package tor_rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"log"
)

type Rsa struct {
	PublicKey *rsa.PublicKey
}

/*
	***The key using: sha256 and empty label***
	This C'tor get RSA public_pem_key as bytes and return parsed key to user

	public_pem_key []byte: -----BEGIN PUBLIC KEY-----key-----BEGIN PUBLIC KEY-----
*/
func NewRsaGivenPemPublicKey(public_pem_key []byte) *Rsa {
	r := &Rsa{BytesToPublicKey(public_pem_key)}
	return r
}

func BytesToPublicKey(pub []byte) *rsa.PublicKey {
	block, _ := pem.Decode(pub)
	b := block.Bytes

	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		log.Println(err)
	}

	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		log.Println("not ok")
	}

	return key
}

/*
	Encrypts a given buffer with the rsa public key

	data byte[]: data to encrypt
*/
func (r *Rsa) Encrypt(data []byte) []byte {
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, r.PublicKey, data, []byte{})
	if err != nil {
		log.Println(err)
	}

	return ciphertext
}
