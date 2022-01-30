package chat_server

import (
	"database/sql"

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

func (db *ChatDb) SendMessageDB(content string, roomName string, senderName string) (bool, error) {
	// 1. check if user in room CHECK
	// 2. check if user isnt banned from room CHECK
	// 3. write message to db message
	if inRoom, err := db._isUserInRoom(roomName, senderName); err != nil || !inRoom {
		return false, err
	}
	if inBan, err := db._isUserInBan(roomName, senderName); err != nil || inBan {
		return false, err
	}

	sql := `
		INSERT INTO messages ( senderID, chatID, content, time ) VALUES ( ?, ?, ?, datetime('now') );
	`

	userID, _ := db._getUserID(senderName)
	chatID, _ := db._getChatRoomID(roomName)

	err := db._execNoneResponseQuery(sql, userID, chatID, content)
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
	err := db._execNoneResponseQuery(sql, username, password)
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
	var dbPassword string
	sql := `
		SELECT password FROM users WHERE username = ?;
	`

	err := db.QueryRow(sql, username).Scan(&dbPassword)
	if err != nil {
		return false // username not found
	}

	return dbPassword == password
}

func (db *ChatDb) CheckChatRoomPassword(roomName string, roomPassword string) bool {
	var dbPassword string
	sql := `
		SELECT password FROM chats WHERE name = ?;
	`

	err := db.QueryRow(sql, roomName).Scan(&dbPassword)
	if err != nil {
		return false // chatRoom not found
	}

	return dbPassword == roomPassword
}

/*
	creating room to DB with given parameters
*/
func (db *ChatDb) CreateChatRoomDB(roomName string, roomPassword string, adminName string) (bool, error) {
	adminID, err := db._getUserID(adminName)
	if err != nil || adminID == WITHOUT_ID {
		return false, err
	}

	sql := `
		INSERT INTO chats ( name, password, adminID ) VALUES ( ?, ?, ? );
	`

	err = db._execNoneResponseQuery(sql, roomName, roomPassword, adminID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *ChatDb) DeleteChatRoomDB(roomName string, roomPassword string, adminName string) (bool, error) {
	adminID, err := db._getUserID(adminName)
	if err != nil || adminID == WITHOUT_ID {
		return false, err
	}

	if !db._rowExists("SELECT * FROM chats WHERE name = ? AND password = ? AND adminID = ?", roomName, roomPassword, adminID) {
		return false, err // not all credentials are right
	}

	err = db._deleteRoomMessages(roomName, roomPassword, adminID) //TODO: test this function
	if err != nil {
		return false, err
	}

	err = db._deleteRoomMembers(roomName, roomPassword, adminID)
	if err != nil {
		return false, err
	}

	err = db._deleteRoom(roomName, roomPassword, adminID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *ChatDb) JoinChatRoomDB(roomName string, roomPassword string, username string, banState int) (bool, error) {
	// using the userID in db
	userId, err := db._getUserID(username)
	if err != nil || userId == WITHOUT_ID {
		return false, err
	}

	// using the the chatID in db
	chatID, err := db._getChatRoomID(roomName)
	if err != nil || chatID == WITHOUT_ID {
		return false, err
	}

	if inBan, err := db._isUserInBan(roomName, username); err != nil || inBan {
		logger.Info.Println("user:", username, " in ban")
		return false, nil
	}

	// password has to match or giving ban
	if !db.CheckChatRoomPassword(roomName, roomPassword) && banState == 0 {
		return false, nil
	}

	sql := `
		INSERT INTO chats_members ( userID, chatID, state ) VALUES ( ?, ?, ? );
	`

	err = db._execNoneResponseQuery(sql, userId, chatID, banState)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *ChatDb) KickFromChatRoomDB(roomName string, username string, adminName string) (bool, error) {
	isAdmin, err := db._isAdminOfRoom(roomName, adminName)
	if !isAdmin || err != nil {
		return false, err
	}

	userID, err := db._getUserID(username)
	if err != nil || userID == WITHOUT_ID {
		return false, err // username not exists at all
	}

	chatID, err := db._getChatRoomID(roomName)
	if err != nil || chatID == WITHOUT_ID {
		return false, err // room not exists at all
	}

	sql := `
		DELETE FROM chats_members WHERE userID = ? AND chatID = ?;
	`

	err = db._execNoneResponseQuery(sql, userID, chatID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *ChatDb) BanFromChatRoomDB(roomName string, username string, adminName string) (bool, error) {
	isAdmin, err := db._isAdminOfRoom(roomName, adminName)
	if !isAdmin || err != nil {
		return false, err
	}

	isUserInRoom, err := db._isUserInRoom(roomName, username)
	if err != nil {
		return false, err
	}

	userID, err := db._getUserID(username)
	if err != nil {
		return false, err
	}

	chatID, err := db._getChatRoomID(roomName)
	if err != nil {
		return false, err
	}

	sql := `
		UPDATE chats_members SET state = ? WHERE userID = ? AND chatID = ?;
	`
	if isUserInRoom {
		err = db._execNoneResponseQuery(sql, STATE_BAN, userID, chatID)
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		return db.JoinChatRoomDB(roomName, "", username, STATE_BAN)
	}
}

func (db *ChatDb) UnBanFromChatRoomDB(roomName string, username string, adminName string) (bool, error) {
	adminID, err := db._getUserID(adminName)
	if err != nil || adminID == WITHOUT_ID {
		return false, err
	}

	if !db._rowExists("SELECT * FROM chats WHERE name = ? AND adminID = ?", roomName, adminID) {
		return false, err // not all credentials are right
	}

	userID, err := db._getUserID(username)
	if err != nil || userID == WITHOUT_ID {
		return false, err // check if user exists
	}

	chatID, err := db._getChatRoomID(roomName)
	if err != nil || chatID == WITHOUT_ID {
		return false, err // check if room exists
	}

	if !db._rowExists("SELECT * FROM chats_members WHERE userID = ? AND chatID = ? AND state = 1", userID, chatID) {
		return false, err //user not in ban
	}

	sql := `
		DELETE FROM chats_members WHERE userID = ? AND chatID = ? AND state = 1;
	`
	err = db._execNoneResponseQuery(sql, userID, chatID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func CloseDB() {
	db.Close()
}
