package router

import (
	"log"
	"net"
)

const (
	REQ_CODE_SIZE  = 2
	CODE_NODE_CONN = "00"
	CODE_NODE_DIS  = "01"
	CODE_ROUTE     = "11"
)

/*
	runs the server on a given address
	calls HandleClient for a connected client

	address string: "ip:port"
*/
func RunRouter(address string) {
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
	the router will be handling 3 different requests:
		* node connection
		* node disconnection
		* tor client route request
*/
func HandleClient(conn net.Conn) {
	msgCode := make([]byte, REQ_CODE_SIZE)
	_, err := conn.Read(msgCode)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	switch string(msgCode) {
	case CODE_NODE_CONN:
		// TODO: node connection
	case CODE_NODE_DIS:
		// TODO: node disconnection
	case CODE_ROUTE:
		// TODO: client route
	default:
		// TODO: send error msg to client
		log.Println("invalid req code")
		return
	}
}
