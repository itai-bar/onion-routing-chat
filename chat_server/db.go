package chat_server

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

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
			adminID		INTEGER,
			FOREIGN KEY(adminID) REFERENCES users(ID)
		);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	//creating chat members table
	sqlStmt = `
		CREATE TABLE IF NOT EXISTS chats_members(
			ID 			INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			userID	 	TEXT NOT NULL UNIQUE,
			chatID	 	TEXT NOT NULL,
			state		INTEGER,
			FOREIGN KEY(userID) REFERENCES users(ID),
			FOREIGN KEY(chatID) REFERENCES chats(ID)
		);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
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
	adminID, err := db.GetUserID(adminName)
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
	db._saveCurrentState() // in case of error we don't want data to get harm
	adminID, err := db.GetUserID(adminName)
	if err != nil || adminID == WITHOUT_ID {
		return false, err
	}

	if !db._rowExists("SELECT * FROM chats WHERE name = ? AND password = ? AND adminID = ?", roomName, roomPassword, adminID) {
		return false, err // not all credentials are right
	}

	//TODO: delete from messages TABLE, messages that related to deleted chat room

	err = db._deleteRoomMembers(roomName, roomPassword, adminID)
	if err != nil {
		db._revertChanges()
		return false, err
	}

	err = db._deleteRoom(roomName, roomPassword, adminID)
	if err != nil {
		db._revertChanges()
		return false, err
	}

	db._saveChanges() // in case if success we want to save changes

	return true, nil
}

func (db *ChatDb) _deleteRoomMembers(roomName string, roomPassword string, adminID int) error {
	if !db._rowExists("SELECT * FROM chats WHERE name = ? AND password = ? AND adminID = ?", roomName, roomPassword, adminID) {
		return errors.New("wrong credentials, can't delete room members") // not all credentials are right
	}

	chatID, err := db.GetRoomID(roomName)
	if err != nil || chatID == WITHOUT_ID {
		return err
	}

	sql := `
		DELETE FROM chat_members WHERE chatID = ?
	`
	err = db._execNoneResponseQuery(sql, chatID)
	if err != nil {
		return err
	}

	return nil

}

func (db *ChatDb) _deleteRoom(roomName string, roomPassword string, adminID int) error {
	sql := `
		DELETE FROM chats WHERE name = ? AND password = ? AND adminID = ?;
	`
	err := db._execNoneResponseQuery(sql, roomName, roomPassword, adminID)
	if err != nil {
		return err
	}
	return nil
}

func (db *ChatDb) GetRoomID(roomName string) (int, error) {
	var roomId int
	sql := `
		SELECT ID FROM chats WHERE name = ?;
	`

	err := db.QueryRow(sql, roomName).Scan(&roomId)
	if err != nil {
		return WITHOUT_ID, err // room not found
	}

	return roomId, nil
}

func (db *ChatDb) GetUserID(username string) (int, error) {
	var userID int
	sql := `
		SELECT ID FROM users WHERE username = ?;
	`

	err := db.QueryRow(sql, username).Scan(&userID)
	if err != nil {
		return WITHOUT_ID, err // username not found
	}

	return userID, nil
}

func (db *ChatDb) _execNoneResponseQuery(query string, args ...interface{}) error {
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	stmt.Close()
	return nil
}

func (db *ChatDb) _saveCurrentState() error {
	sql := `
		BEGIN TRANSACTION;
	`
	return db._execNoneResponseQuery(sql)
}

func (db *ChatDb) _saveChanges() error {
	sql := `
		END TRANSACTION;
	`
	return db._execNoneResponseQuery(sql)
}

func (db *ChatDb) _revertChanges() error {
	sql := `
		ROLLBACK;
	`
	return db._execNoneResponseQuery(sql)
}

// helper to check if x exist
func (db *ChatDb) _rowExists(query string, args ...interface{}) bool {
	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err := db.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists '%s' %v", args, err)
	}
	log.Printf("%t %s", exists, query)
	return exists
}

func CloseDB() {
	db.Close()
}
