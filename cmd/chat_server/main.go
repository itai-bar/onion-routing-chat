package main

import (
	"torbasedchat/chat_server"
	"torbasedchat/pkg/tor_server"
)

func main() {
	chat_server.InitializeClientsMap()
	tor_server.RunServer("0.0.0.0:8989", chat_server.HandleClient)
}
