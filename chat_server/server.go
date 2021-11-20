package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

const (
	DATA_SIZE_SEGMENT_SIZE = 5
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
	defer conn.Close()

	for {
		dataSizeBuf := make([]byte, DATA_SIZE_SEGMENT_SIZE)
		_, err := conn.Read(dataSizeBuf)
		if err != nil {
			continue
		}

		// data size is zero padded
		dataSize, err := strconv.Atoi(string(RemoveLeadingChars(dataSizeBuf, '0')))
		if err != nil {
			continue
		}

		dataBuf := make([]byte, dataSize)
		_, err = conn.Read(dataBuf)
		if err != nil {
			continue
		}
		log.Printf("Client %s: %s\n", conn.RemoteAddr().String(), string(dataBuf))

		// zero padding the size again
		resp := fmt.Sprintf("%05d", len(dataBuf)) + string(dataBuf)
		conn.Write([]byte(resp))

		if string(dataBuf) == "Exit" {
			break
		}
	}
}

/*
	removes every char c from the start of byte array s
*/
func RemoveLeadingChars(s []byte, c byte) []byte {
	for i := range s {
		if s[i] != c {
			return s[i:]
		}
	}
	return []byte{}
}
