package chat_server

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

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
func RegisterDB(db *sql.DB, username string, password string) (bool, error) {
	/*if RowExists("SELECT username FROM users WHERE username = ?", username) {
		return false, nil
	}*/

	sql := `
		INSERT INTO users ( username, password ) VALUES ( ?, ? );
	`

	stmt, err := db.Prepare(sql)
	if err != nil {
		return false, err
	}
	_, err = stmt.Exec(username, password)
	if err != nil {
		return false, err
	}

	stmt.Close()

	return true, nil
}

/*
	checks if the password of a given username is the given password
	helps for login
*/
func CheckUsersPassword(db *sql.DB, username string, password string) bool {
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

// helper to check if x exist
func RowExists(query string, args ...interface{}) bool {
	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err := db.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists '%s' %v", args, err)
	}
	log.Printf("%t %s", exists, query)
	return exists
}
