package chat_server

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDb(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	sqlStmt := `
		CREATE TABLE IF NOT EXISTS USERS(
			ID 			INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
			USERNAME 	TEXT NOT NULL,
			PASSWORD 	TEXT NOT NULL
		);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func RegisterDB(db *sql.DB, username string, password string) error {
	// TODO: check if username exists already, add return value..
	sql := `
		INSERT INTO USERS ( USERNAME, PASSWORD ) VALUES ( ?, ? );
	`

	stmt, err := db.Prepare(sql)
	if err != nil {
		return err
	}
	stmt.Exec(username, password)
	stmt.Close()

	return nil
}

/*
	checks if the password of a given username is the given password
	helps for login
*/
func CheckUsersPassword(db *sql.DB, username string, password string) bool {
	var dbPassword string
	sql := `
		SELECT PASSWORD FROM USERS WHERE USERNAME = ?;
	`

	err := db.QueryRow(sql, username).Scan(&dbPassword)
	if err != nil {
		return false
	}

	return dbPassword == password
}
