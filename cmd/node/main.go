package main

import (
	"log"

	"torbasedchat/node"
	"torbasedchat/pkg/tor_aes"
)

func main() {
	a := tor_aes.NewAesSize(32)
	log.Println(a.Encrypt([]byte("hello")))
	node.RunNode("0.0.0.0:8989")
}
