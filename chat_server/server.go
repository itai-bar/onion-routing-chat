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
	l, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("error in listening: ", err)
	}
	defer l.Close() // will close the listener when the function exits
	log.Println("Listening on ", address)

	for {
		conn, err := l.Accept() // new client
		if err != nil {
			log.Fatal("err on accepting client: ", err)
		}
		log.Println("new client: ", conn.RemoteAddr().String())

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

		log.Printf("client %s: %s\n", conn.RemoteAddr().String(), string(buf))

		conn.Write([]byte("hello from the server"))

		if string(buf) == "Exit" {
			break
		}
	}
	conn.Close()
}
