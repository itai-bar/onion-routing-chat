package chat_server

import (
	"log"
	"net"
	"torbasedchat/pkg/tor_server"
)

const (
	DATA_SIZE_SEGMENT_SIZE = 5
)

const (
	REQ_CODE_SIZE = 2
	CODE_AUTH     = "00"
	CODE_UPDATE   = "01"
	CODE_LOGIN    = "02"
	CODE_SIGN_UP  = "03"
	CODE_LOGOUT   = "04"
	CODE_MSG      = "05"
)

var clients map[Cookie]Client

/*
	handles the connection with every client

	conn net.Conn: connection with a client
*/
func HandleClient(conn net.Conn) {
	defer conn.Close()

	msgCode, err := tor_server.ReadSize(conn, REQ_CODE_SIZE)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	switch string(msgCode) {
	case CODE_AUTH:
		cookie, aes, err := Auth(conn)
		log.Println("the cookie is", cookie)
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}

		// creating the new client, name will be set in login
		clients[*cookie] = Client{"", *aes, conn}
	default:
		// TODO: DEAL WITH DEFAULT AND SEND ERROR MESSAGE
	}
}
