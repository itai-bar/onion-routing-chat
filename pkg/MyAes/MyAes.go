package myaes

import (
	"crypto/aes"
	"encoding/hex"
	"log"
)

/*
	encrypt the plaintext with given aes key

	key []byte: text of 128(16 bytes) bits or 256(32 bytes) bits
	plaintext string: text to encrypt
*/
func EncryptAES(key []byte, plaintext string) string {
	cipher, err := aes.NewCipher(key)
	CheckError(err)

	output := make([]byte, len(plaintext))

	cipher.Encrypt(output, []byte(plaintext))

	return hex.EncodeToString(output)
}

/*
	decrypt the ciphertext with given aes key

	key []byte: text of 128(16 bytes) bits or 256(32 bytes) bits
	ct string: text to decrypt
*/
func DecryptAES(key []byte, ct string) string {
	ciphertext, _ := hex.DecodeString(ct)

	cipher, err := aes.NewCipher(key)
	CheckError(err)

	plaintext := make([]byte, len(ciphertext))

	cipher.Decrypt(plaintext, []byte(ciphertext))

	return string(plaintext)
}

/*
	check wheather happen error or not

	err error: the given error
*/
func CheckError(err error) {
	if err != nil {
		log.Println("error occured:" + err.Error())
	}
}
