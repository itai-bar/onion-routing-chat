package tor_rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type Rsa struct {
	PublicKey *rsa.PublicKey
}

/*
	***The key using: sha256 and empty label***
	This C'tor get RSA public_pem_key as bytes and return parsed key to user

	public_pem_key []byte: -----BEGIN PUBLIC KEY-----key-----BEGIN PUBLIC KEY-----
*/
func NewRsaGivenPemPublicKey(public_pem_key []byte) (*Rsa, error) {
	pk, err := BytesToPublicKey(public_pem_key)
	if err != nil {
		return nil, err
	}

	r := &Rsa{pk}
	return r, nil
}

func BytesToPublicKey(pub []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pub)
	if block == nil {
		return nil, errors.New("Invalid public key")
	}
	b := block.Bytes

	ifc, err := x509.ParsePKIXPublicKey(b)
	if err != nil {
		return nil, err
	}
	key, ok := ifc.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("ifc not ok")
	}

	return key, nil
}

/*
	Encrypts a given buffer with the rsa public key

	data byte[]: data to encrypt
*/
func (r *Rsa) Encrypt(data []byte) ([]byte, error) {
	rng := rand.Reader
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rng, r.PublicKey, data, []byte{})
	if err != nil {
		return nil, err
	}

	return ciphertext, nil
}
