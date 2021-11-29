package tor_rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

type Rsa struct {
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

// creates a new rsa structure with a random key
func NewRsa() (*Rsa, error) {
	// a 256 byte randomly generated key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	r := &Rsa{&privateKey.PublicKey, privateKey}
	return r, nil
}

// encrypts a given buffer with the rsa public key
func (r *Rsa) Encrypt(buf []byte) []byte {
	cipher, _ := rsa.EncryptOAEP(sha256.New(), rand.Reader, r.PublicKey, buf, []byte{})
	return cipher
}

// decrypts given buffer with the rsa private key
func (r *Rsa) Decrypt(buf []byte) []byte {
	plaintext, _ := rsa.DecryptOAEP(sha256.New(), rand.Reader, r.PrivateKey, buf, []byte{})
	return plaintext
}
