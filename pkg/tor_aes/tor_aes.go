package tor_aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

const KEY_SIZE = 32

type Aes struct {
	Key []byte
}

/*
	creates new object of aes key with given key size

	size int: key size
*/
func NewAesRandom() *Aes {
	key := make([]byte, KEY_SIZE)
	rand.Read(key)
	return &Aes{key}
}

/*
	creates new object of aes key with given key

	givenKey []byte: aes key
*/
func NewAesGivenkey(givenKey []byte) *Aes {
	return &Aes{givenKey}
}

/*
	encrypt plaintext with the aes key

	plaintext []byte: plaintext
*/
func (a *Aes) Encrypt(data []byte) ([]byte, error) {
	plaintext := []byte(data) // gotta work with bytes
	block, err := aes.NewCipher(a.Key)
	if err != nil {
		return nil, err
	}

	// need to use a random iv for security
	cipherText := make([]byte, aes.BlockSize+len(plaintext))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plaintext)

	return cipherText, nil
}

/*
	decrypt ciphertext with the aes key

	ciphertext []byte: encrypted text
*/
func (a *Aes) Decrypt(cipherText []byte) ([]byte, error) {

	block, err := aes.NewCipher(a.Key)
	if err != nil {
		return []byte{}, err
	}

	if len(cipherText) < aes.BlockSize {
		return []byte{}, errors.New("cipher text len is smaller than aes blocksize")
	}

	iv, cipherText := cipherText[:aes.BlockSize], cipherText[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(cipherText, cipherText)
	return cipherText, nil
}
