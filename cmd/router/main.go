package main

import (
	"os"
	"torbasedchat/pkg/tor_logger"
	"torbasedchat/pkg/tor_server"
	"torbasedchat/router"
)

func main() {
	logger := tor_logger.NewTorLogger(os.Getenv("ROUTER_LOG"))
	go router.CheckNodes()
	tor_server.RunServer("172.20.0.2:7777", router.HandleClient, logger)
}
