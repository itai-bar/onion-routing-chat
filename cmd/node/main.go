package main

import (
	"log"
	"torbasedchat/node"
)

func main() {
	routerIp := "172.20.0.2:7777"
	if node.NetworkLogon(routerIp) {
		defer node.NetworkLogout("172.20.0.2:7777")
		log.Println("Performed 'Network-Logon' with tor-network.")
		node.RunNode("0.0.0.0:8989")
	} else {
		log.Println("Couldn't perform 'Network-Logon' with tor-network.")
	}
}
