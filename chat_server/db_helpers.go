package chat_server

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

func (db *ChatDb) _deleteRoomMembers(roomName string, roomPassword string, adminID int) error {
	if !db._rowExists("SELECT * FROM chats WHERE name = ? AND password = ? AND adminID = ?", roomName, roomPassword, adminID) {
		return errors.New("wrong credentials, can't delete room members") // not all credentials are right
	}

	chatID, err := db._getRoomID(roomName)
	if err != nil || chatID == WITHOUT_ID {
		return err
	}

	sql := `
		DELETE FROM chats_members WHERE chatID = ?
	`
	err = db._execNoneResponseQuery(sql, chatID)
	if err != nil {
		return err
	}

	return nil

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

func (db *ChatDb) _getRoomID(roomName string) (int, error) {
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

func (db *ChatDb) _getUserID(username string) (int, error) {
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
	return exists
}
