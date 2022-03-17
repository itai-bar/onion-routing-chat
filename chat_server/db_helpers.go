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
	dbMx.Lock()
	defer dbMx.Unlock()

	db._saveCurrentState()
	stmt, err := db.Prepare(query)
	if err != nil {
		db._revertChanges()
		return err
	}

	defer stmt.Close()
	_, err = stmt.Exec(args...)
	if err != nil {
		db._revertChanges()
		return err
	}

	db._saveChanges()
	return nil
}

func (db *ChatDb) _deleteRoom(roomID int, roomPassword string, adminID int) error {
	sql := `
		DELETE FROM chats WHERE ID = ? AND password = ? AND adminID = ?;
	`
	return db._execNoneResponseQuery(sql, roomID, roomPassword, adminID)
}

func (db *ChatDb) _getChatRoomID(roomName string) (int, error) {
	var roomID int
	sql := `
		SELECT ID FROM chats WHERE name = ?;
	`
	dbMx.Lock()
	defer dbMx.Unlock()
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

	dbMx.Lock()
	defer dbMx.Unlock()
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
	return db._rowExists("SELECT * FROM chats_members WHERE userID = ? AND chatID = ? AND state = 0", userID, roomID)
}

func (db *ChatDb) _isUserInBan(roomID int, userID int) bool {
	return db._rowExists("SELECT * FROM chats_members WHERE userID = ? AND chatID = ? AND state = 1",
		userID, roomID)
}

func (db *ChatDb) _saveCurrentState() error {
	sql := `
		BEGIN TRANSACTION;
	`
	_, err := db.Exec(sql)
	return err
}

func (db *ChatDb) _saveChanges() error {
	sql := `
		END TRANSACTION;
	`
	_, err := db.Exec(sql)
	return err
}

func (db *ChatDb) _revertChanges() error {
	sql := `
		ROLLBACK;
	`
	_, err := db.Exec(sql)
	return err
}

// helper to check if x exist
func (db *ChatDb) _rowExists(query string, args ...interface{}) bool {
	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)

	dbMx.Lock()
	defer dbMx.Unlock()

	err := db.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		logger.Err.Fatalf("error checking if row exists '%s' %v", args, err)
	}

	return exists
}
