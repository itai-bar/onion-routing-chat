package main

import (
	"torbasedchat/pkg/tor_server"
	"torbasedchat/router"
)

func main() {
	go router.CheckNodes()
	tor_server.RunServer("172.20.0.2:7777", router.HandleClient)
}
