package node

import (
	"bytes"
	"log"
	"net"
)

/*
	trying to perform with router, login to tor-network

	routerAddress string: "ip:port"

	return-value: true - connection agreed. false - connection refused.
*/
func NetworkLogon(routerAddress string) bool {
	routerConn, err := net.Dial("tcp", routerAddress)
	if err != nil {
		log.Println("ERROR: ", err)
		return false
	}

	routerConn.Write([]byte(CODE_NODE_CONN))

	respBuf := make([]byte, 1)
	_, err = routerConn.Read(respBuf)
	if err != nil {
		return false
	}

	routerConn.Close()

	passed := []byte("1") // "1" for true and performed logon succesfully
	res := bytes.Compare(respBuf, passed)

	return res == 0 //if res==0-> connection passed. else-> connection refused
}

/*
	disconnect from tor-network

	routerAddress string: "ip:port"
*/
func NetworkLogout(routerAddress string) {
	routerConn, err := net.Dial("tcp", routerAddress)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	routerConn.Write([]byte(CODE_NODE_DIS))
	routerConn.Close()

	log.Println("Logged out from tor-network")
}
