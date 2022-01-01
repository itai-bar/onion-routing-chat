package main

import (
	"torbasedchat/pkg/tor_server"
	"torbasedchat/router"
)

func main() {
	tor_server.RunServer("172.20.0.2:7777", router.HandleClient)
}
