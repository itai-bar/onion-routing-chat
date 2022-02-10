package chat_server

import (
	"net"
	"torbasedchat/pkg/tor_aes"
)

type Client struct {
	username string
	aes      tor_aes.Aes
	conn     net.Conn
}
