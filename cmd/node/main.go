package main

import (
	"torbasedchat/node"
)

func main() {
	//routerIp := "static-ip:7777" // TODO: change it to the static-ip of router when task is done
	node.RunNode("0.0.0.0:8989") // TODO: remove this line when finishing with router static-ip
	/*if node.NetworkLogon(routerIp){
		log.Println("Performed 'Network-Logon' with tor-network.")
		node.RunNode("0.0.0.0:8989")
	}else{
		log.Println("Couldn't perform 'Network-Logon' with tor-network.")
	}*/ //TODO:change from comment to valid code when router-static ip is avialble
}
