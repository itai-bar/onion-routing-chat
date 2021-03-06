package chat_server

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"torbasedchat/pkg/tor_aes"
	"torbasedchat/pkg/tor_rsa"
)

const (
	COOKIE_SIZE = 15
)

// used to create a more statefull connection with the clients
type Cookie struct {
	data [15]byte
}

/*
	generates a random unique cookie to identify clients
*/
func CreateCookie() *Cookie {
	var c Cookie
	data := make([]byte, COOKIE_SIZE)

	for {
		rand.Read(data)

		for i, data := range data {
			c.data[i] = data
		}
		_, inMap := clients[c]

		if !inMap {
			break
		}
	}

	return &c
}

func InitCookie(data []byte) (*Cookie, error) {
	var c Cookie

	if len(data) != COOKIE_SIZE {
		return nil, errors.New("cookie size not 15")
	}

	for i, data := range data {
		c.data[i] = data
	}

	return &c, nil
}

/*
	performs a key change with the client using a given RSA key
	returns aes key to use with the client, cookie to identify a client and error
*/
func Auth(conn net.Conn, pemKey []byte) (*Cookie, *tor_aes.Aes, error) {
	// inits a rsa object with the key we got from the client
	// creating the aes key for the rest of comm
	rsa, err := tor_rsa.NewRsaGivenPemPublicKey(pemKey)
	aes := tor_aes.NewAesRandom()
	if err != nil {
		return nil, nil, err
	}

	// adding the aes key and the cookie for the client
	cookie := CreateCookie()

	buf, err := rsa.Encrypt(append(cookie.data[:], aes.Key...))
	if err != nil {
		return nil, nil, err
	}

	// padding the length with to 5 chars
	paddedLen := fmt.Sprintf("%05d", len(buf))
	buf = append([]byte(paddedLen), buf...)

	conn.Write(buf)

	logger.Info.Printf("performed auth with new client ")
	return cookie, aes, nil
}
