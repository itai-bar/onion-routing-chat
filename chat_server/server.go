package chat_server

import (
	"fmt"
	"log"
	"net"
	"torbasedchat/pkg/tor_server"
)

const (
	DATA_SIZE_SEGMENT_SIZE = 5
)

/*
	handles the connection with every client
	for now its just exchanging string until a client writes Exit

	conn net.Conn: connection with a client
*/
func HandleClient(conn net.Conn) {
	defer conn.Close()

	for {
		dataBuf, err := tor_server.ReadDataFromSizeHeader(conn, DATA_SIZE_SEGMENT_SIZE)
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
