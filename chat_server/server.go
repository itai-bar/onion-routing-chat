package main

import (
	"log"
	"net"
)

/*
	runs the server on a given address
	calls HandleClient for a connected client

	address string: "ip:port"
*/
func RunServer(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Error in listening:\t", err)
	}
	defer listener.Close() // will close the listener when the function exits
	log.Println("Listening on:\t", address)

	for {
		conn, err := listener.Accept() // new client
		if err != nil {
			log.Fatal("Error on accepting client:\t", err)
		}
		log.Println("New client:\t", conn.RemoteAddr().String())

		go HandleClient(conn) // new thread to handle the client
	}
}

/*
	handles the connection with every client
	for now its just exchanging string until a client writes Exit

	conn net.Conn: connection with a client
*/
func HandleClient(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		_, err := conn.Read(buf)
		if err != nil {
			continue
		}

		log.Printf("Client %s:\t%s\n", conn.RemoteAddr().String(), string(buf))
		conn.Write([]byte("hello"))

		if string(buf) == "Exit" {
			break
		}
	}
	conn.Close()
}
