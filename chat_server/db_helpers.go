package chat_server

import (
	"database/sql"
	"fmt"
)

func (db *ChatDb) _deleteRoomMessages(roomID int, roomPassword string, adminID int) error {
	sql := `
		DELETE FROM messages WHERE chatID = ?;
	`

	err := db._execNoneResponseQuery(sql, roomID)
	if err != nil {
		return err
	}

	return nil
}
func (db *ChatDb) _deleteRoomMembers(roomID int, roomPassword string, adminID int) error {
	sql := `
		DELETE FROM chats_members WHERE chatID = ?
	`
	err := db._execNoneResponseQuery(sql, roomID)
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

func (db *ChatDb) _deleteRoom(roomID int, roomPassword string, adminID int) error {
	sql := `
		DELETE FROM chats WHERE ID = ? AND password = ? AND adminID = ?;
	`
	err := db._execNoneResponseQuery(sql, roomID, roomPassword, adminID)
	if err != nil {
		return err
	}
	return nil
}

func (db *ChatDb) _getChatRoomID(roomName string) (int, error) {
	var roomID int
	sql := `
		SELECT ID FROM chats WHERE name = ?;
	`

	err := db.QueryRow(sql, roomName).Scan(&roomID)
	if err != nil {
		return WITHOUT_ID, err // room not found
	}

	return roomID, nil
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

func (db *ChatDb) _isAdminOfRoom(roomID int, userID int) bool {
	return db._rowExists("SELECT * FROM chats WHERE ID = ? AND adminID = ?", roomID, userID)
}

func (db *ChatDb) _isUserInRoom(roomID int, userID int) bool {
	return db._rowExists("SELECT * FROM chats_members WHERE userID = ? AND chatID = ?", userID, roomID)
}

func (db *ChatDb) _isUserInBan(roomID int, userID int) bool {
	return db._rowExists("SELECT * FROM chats_members WHERE userID = ? AND chatID = ? AND state = 1",
		userID, roomID)
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
		logger.Err.Fatalf("error checking if row exists '%s' %v", args, err)
	}
	return exists
}
