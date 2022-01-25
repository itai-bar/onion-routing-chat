package main

import (
	"os"
	"torbasedchat/chat_server"
	"torbasedchat/pkg/tor_logger"
	"torbasedchat/pkg/tor_server"
)

func main() {
	logger := tor_logger.NewTorLogger(os.Getenv("CHAT_LOG"))
	defer chat_server.CloseDB()
	tor_server.RunServer("0.0.0.0:8989", chat_server.HandleClient, logger)
}
