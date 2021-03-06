package chat_server

import (
	"database/sql"
	"time"

	"crypto/sha256"

	_ "github.com/mattn/go-sqlite3"
)

type ChatDb struct {
	*sql.DB
}

const WITHOUT_ID = 0

func InitChatDb(db *sql.DB) *ChatDb {
	return &ChatDb{db}
}

func InitDb(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	//creating users table
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS users(
			ID 			INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			username 	TEXT NOT NULL UNIQUE,
			password 	TEXT NOT NULL
		);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	//creating chats table
	sqlStmt = `
		CREATE TABLE IF NOT EXISTS chats(
			ID 			INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			name	 	TEXT NOT NULL UNIQUE,
			password 	TEXT NOT NULL,
			adminID		INTEGER NOT NULL,
			FOREIGN KEY(adminID) REFERENCES users(ID)
		);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	// creating chat members table
	// state field: there are two states of connections to a room.
	// 		0 - a room member
	//		1 - banned from the room by an admin

	sqlStmt = `
		CREATE TABLE IF NOT EXISTS chats_members(
			ID 			INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			userID	 	TEXT NOT NULL, 
			chatID	 	TEXT NOT NULL,
			state		INTEGER NOT NULL,
			FOREIGN KEY(userID) REFERENCES users(ID),
			FOREIGN KEY(chatID) REFERENCES chats(ID),
			CONSTRAINT room_member UNIQUE (userID, chatID)
		);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	sqlStmt = `
		CREATE TABLE IF NOT EXISTS messages(
			ID 			INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			senderID 	INTEGER NOT NULL, 
			chatID	 	INTEGER NOT NULL,
			content		TEXT NOT NULL,
			time		TIME NOT NULL,
			FOREIGN KEY(senderID) REFERENCES users(ID),
			FOREIGN KEY(chatID) REFERENCES chats(ID)
		);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *ChatDb) LoadRoomsFromDB() error {
	rooms, err := db.GetRoomsDB(-1)
	if err != nil {
		return err
	}
	for _, room := range rooms {
		chatRooms[room] = &ChatRoom{onlineMembers: make([]*Client, 0)}
	}
	return nil
}

func (db *ChatDb) SendMessageDB(content string, roomID int, senderID int, time time.Time) (bool, error) {
	sql := `
		INSERT INTO messages ( senderID, chatID, content, time ) VALUES ( ?, ?, ?, ? );
	`

	err := db._execNoneResponseQuery(sql, senderID, roomID, content, time)
	if err != nil {
		return false, err
	}

	return true, err
}

/*
	checks if a username exists in database, if not it registers a new user
*/
func (db *ChatDb) RegisterDB(username string, password string) (bool, error) {
	if db._rowExists("SELECT username FROM users WHERE username = ?", username) {
		return false, nil
	}

	sql := `
		INSERT INTO users ( username, password ) VALUES ( ?, ? );
	`
	hashOfPassword := sha256.Sum256([]byte(password))
	err := db._execNoneResponseQuery(sql, username, string(hashOfPassword[:]))
	if err != nil {
		return false, err
	}

	return true, nil
}

/*
	checks if the password of a given username is the given password
	helps for login
*/
func (db *ChatDb) CheckUsersPassword(username string, password string) bool {
	var dbPasswordHash string
	sql := `
		SELECT password FROM users WHERE username = ?;
	`

	err := db.QueryRow(sql, username).Scan(&dbPasswordHash)
	if err != nil {
		return false // username not found
	}
	hashOfPassword := sha256.Sum256([]byte(password))
	// does the password input match the db password
	return dbPasswordHash == string(hashOfPassword[:])
}

func (db *ChatDb) isRoomPassword(roomID int, roomPassword string) bool {
	var dbPassword string
	sql := `
		SELECT password FROM chats WHERE ID = ?;
	`

	err := db.QueryRow(sql, roomID).Scan(&dbPassword)
	if err != nil {
		return false // chatRoom not found
	}

	return dbPassword == roomPassword
}

/*
	creating room to DB with given parameters
*/
func (db *ChatDb) CreateChatRoomDB(roomName string, roomPassword string, adminID int) (bool, error) {
	sql := `
		INSERT INTO chats ( name, password, adminID ) VALUES ( ?, ?, ? );
	`

	err := db._execNoneResponseQuery(sql, roomName, roomPassword, adminID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *ChatDb) DeleteChatRoomDB(roomID int, roomPassword string, adminID int) (bool, error) {
	err := db._deleteRoomMessages(roomID, roomPassword, adminID) //TODO: test this function
	if err != nil {
		return false, err
	}

	err = db._deleteRoomMembers(roomID, roomPassword, adminID)
	if err != nil {
		return false, err
	}

	err = db._deleteRoom(roomID, roomPassword, adminID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *ChatDb) JoinChatRoomDB(roomID int, roomPassword string, userID int, banState int) (bool, error) {
	inBan := db._isUserInBan(roomID, userID)
	if inBan {
		logger.Info.Println("user:", userID, " in ban")
		return false, nil
	}

	// password has to match or giving ban
	if !db.isRoomPassword(roomID, roomPassword) && banState == 0 {
		return false, nil
	}

	sql := `
		INSERT INTO chats_members ( userID, chatID, state ) VALUES ( ?, ?, ? );
	`

	err := db._execNoneResponseQuery(sql, userID, roomID, banState)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *ChatDb) KickFromChatRoomDB(roomID int, userID int, adminID int) (bool, error) {
	sql := `
		DELETE FROM chats_members WHERE userID = ? AND chatID = ?;
	`

	err := db._execNoneResponseQuery(sql, userID, roomID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *ChatDb) BanFromChatRoomDB(roomID int, userID int, adminID int) (bool, error) {
	isUserInRoom := db._isUserInRoom(roomID, userID)

	sql := `
		UPDATE chats_members SET state = ? WHERE userID = ? AND chatID = ?;
	`
	if isUserInRoom {
		err := db._execNoneResponseQuery(sql, STATE_BAN, userID, roomID)
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		return db.JoinChatRoomDB(roomID, "", userID, STATE_BAN)
	}
}

func (db *ChatDb) UnBanFromChatRoomDB(roomID int, userID int, adminID int) (bool, error) {
	if !db._isUserInBan(roomID, userID) {
		return false, nil //user not in ban
	}

	sql := `
		DELETE FROM chats_members WHERE userID = ? AND chatID = ? AND state = 1;
	`
	err := db._execNoneResponseQuery(sql, userID, roomID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *ChatDb) LoadLastMessages(roomId int, amount int, offset int) ([]Message, error) {
	sql := `
		SELECT chats.name, users.username, messages.content, strftime("%s", datetime(messages.time)) 
		FROM messages
		INNER JOIN users
		ON users.ID = messages.senderID
		INNER JOIN chats
		ON chats.ID = messages.chatID
		WHERE messages.chatID = ? ORDER BY messages.time DESC LIMIT ? OFFSET ?;
	`

	var messages []Message

	rows, err := db.Query(sql, roomId, amount, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var roomName string
		var sender string
		var content string
		var unixTime int64

		err = rows.Scan(&roomName, &sender, &content, &unixTime)
		if err != nil {
			return nil, err
		}

		messages = append(messages, Message{roomName, content, sender, time.Unix(unixTime, 0)})
	}

	return messages, err
}

// set userID to -1 to get all room on the db
// else the function will get all the rooms the user
// isn't banned from
func (db *ChatDb) GetRoomsDB(userID int) ([]string, error) {
	sql := "SELECT name FROM chats"
	if userID != -1 {
		// selecting all rooms but the ones that userID is banned on
		sql = `
			SELECT name
			FROM chats

			WHERE id NOT IN (
				SELECT chats.ID
				FROM chats
				INNER JOIN chats_members
				ON chats_members.chatID = chats.ID
				WHERE chats_members.userID = ? AND chats_members.state = 1
			)
		`
	}

	var rooms []string

	rows, err := db.Query(sql, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var roomName string

		err = rows.Scan(&roomName)
		if err != nil {
			return nil, err
		}

		rooms = append(rooms, roomName)
	}

	return rooms, nil
}

func (db *ChatDb) LeaveRoomDB(roomID int, userID int) (bool, error) {
	sql := `DELETE FROM chats_members WHERE chatID = ? AND userID = ? AND state = 0`
	err := db._execNoneResponseQuery(sql, roomID, userID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *ChatDb) GetOfflineMembersInRoomDB(roomID int, onlineMembersNames []string) ([]string, error) {
	sql := `
		SELECT users.username
		FROM chats_members
		INNER JOIN users
		ON users.ID = chats_members.userID
		WHERE chats_members.chatID = ? AND chats_members.state = 0;
	`
	var membersNames []string

	rows, err := db.Query(sql, roomID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var memberName string

		err = rows.Scan(&memberName)
		if err != nil {
			return nil, err
		}

		membersNames = append(membersNames, memberName)
	}

	//Erase online members from all members, to seperate offline members from online members
	//TODO:fix bug of empty name in list (occur when user re-enter to chat after x button)
	for _, onlineMemberName := range onlineMembersNames {
		for memberIdx, memberName := range membersNames {
			if onlineMemberName == memberName {
				membersNames = append(membersNames[:memberIdx], membersNames[memberIdx+1:]...)
				break
			}
		}
	}

	return membersNames, nil
}

func (db *ChatDb) GetBannedMembersDB(roomID int) ([]string, error) {
	sql := `
		SELECT users.username
		FROM chats_members
		INNER JOIN users
		ON users.ID = chats_members.userID
		WHERE chats_members.chatID = ? AND chats_members.state = 1;
	`
	var bannedMembers []string

	rows, err := db.Query(sql, roomID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var memberName string

		err = rows.Scan(&memberName)
		if err != nil {
			return nil, err
		}

		bannedMembers = append(bannedMembers, memberName)
	}
	return bannedMembers, nil
}
func CloseDB() {
	db.Close()
}
