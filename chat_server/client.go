package chat_server

import (
	"torbasedchat/pkg/tor_aes"
)

type Client struct {
	username string
	aes      tor_aes.Aes
}
