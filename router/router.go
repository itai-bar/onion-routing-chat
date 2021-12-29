package router

import (
	"log"
	"net"
	"sync"
)

const (
	REQ_CODE_SIZE  = 2
	CODE_NODE_CONN = "00"
	CODE_NODE_DIS  = "01"
	CODE_ROUTE     = "11"
)

var networkLock sync.Mutex

type TorNetwork map[string]struct{}

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

	network := make(TorNetwork)

	for {
		conn, err := listener.Accept() // new client
		if err != nil {
			log.Fatal("Error on accepting client:\t", err)
		}
		log.Println("New client:\t", conn.RemoteAddr().String())

		go HandleClient(conn, network) // new thread to handle the client
	}
}

/*
	the router will be handling 3 different requests:
		* node connection
		* node disconnection
		* tor client route request
*/
func HandleClient(conn net.Conn, network TorNetwork) {
	msgCode := make([]byte, REQ_CODE_SIZE)
	_, err := conn.Read(msgCode)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	// the map is a mutual resource
	networkLock.Lock()
	defer networkLock.Unlock()

	switch string(msgCode) {
	case CODE_NODE_CONN:
		network[conn.RemoteAddr().String()] = struct{}{} // init en empty struct to the map
	case CODE_NODE_DIS:
		delete(network, conn.RemoteAddr().String())
	case CODE_ROUTE:
		// TODO: client route
	default:
		// TODO: send error msg to client
		log.Println("invalid req code")
		return
	}
}
