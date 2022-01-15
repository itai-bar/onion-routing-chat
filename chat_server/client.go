package chat_server

import "torbasedchat/pkg/tor_aes"

type Client struct {
	username string
	aesObj   *tor_aes.Aes
}
