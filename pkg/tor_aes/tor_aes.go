package tor_aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"log"
)

type Aes struct {
	key []byte
}

/*
	creates new object of aes key with given key size

	size int: key size
*/
func NewAesSize(size int) *Aes {
	key := make([]byte, size)
	rand.Read(key)
	return &Aes{key}
}

/*
	creates new object of aes key with given key

	givenKey []byte: aes key
*/
func NewAesGiveKey(givenKey []byte) *Aes {
	return &Aes{givenKey}
}

/*
	encrypt plaintext with the aes key 'self.key'

	plaintext []byte: plaintext
*/
func (self *Aes) Encrypt(data []byte) []byte {
	c, err := aes.NewCipher(self.key)
	if err != nil {
		log.Println(err)
		return nil
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		log.Println(err)
		return nil
	}

	nonce := make([]byte, gcm.NonceSize())

	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Println(err)
		return nil
	}

	return gcm.Seal(nonce, nonce, data, nil)
}

/*
	decrypt ciphertext with the aes key 'self.key'

	ciphertext []byte: encrypted text
*/
func (self *Aes) Decrypt(ciphertext []byte) []byte {
	c, err := aes.NewCipher(self.key)
	if err != nil {
		log.Println(err)
		return nil
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		log.Println(err)
		return nil
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		log.Println(err)
		return nil
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Println(err)
		return nil
	}

	return plaintext
}
