package main

import (
	"os"
	"torbasedchat/node"
	"torbasedchat/pkg/tor_logger"
	"torbasedchat/pkg/tor_server"
)

const (
	ROUTER_IP = "172.20.0.2:7777"
	SELF_IP   = "0.0.0.0:8989"
)

func main() {
	logger := tor_logger.NewTorLogger(os.Getenv("NODE_LOG"))
	if node.NetworkLogon(ROUTER_IP) {
		defer node.NetworkLogout(ROUTER_IP)

		logger.Info.Println("Performed 'Network-Logon' with tor-network.")
		tor_server.RunServer(SELF_IP, node.HandleClient, logger)
	} else {
		logger.Err.Println("Couldn't perform 'Network-Logon' with tor-network.")
	}
}
