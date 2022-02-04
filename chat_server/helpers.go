package chat_server

import "bytes"

/*
	check if password with chars for sql injection, if it is - invalid password.
*/
func isValidPassword(password string) bool {
	invalidChars := []byte{33, 34, 35, 37, 39, 42, 43, 45, 47, 59, 60, 61, 62, 92, 96, 124}
	for _, invalidChar := range invalidChars {
		if bytes.Contains([]byte(password), []byte{invalidChar}) {
			return false
		}
	}
	return true
}
