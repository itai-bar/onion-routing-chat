package chat_server

import (
	"database/sql"
	"encoding/json"
	"fmt"
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
	CODE_REGISTER = "03"
	CODE_LOGOUT   = "04"
	CODE_MSG      = "05"
)

var clients map[Cookie]Client
var db *sql.DB

func init() {
	var err error

	clients = make(map[Cookie]Client)
	db, err = InitDb("/app/db.sqlite")
	if err != nil {
		log.Fatal("ERROR: ", err)
	}
}

/*
	handles the connection with every client

	conn net.Conn: connection with a client
*/
func HandleClient(conn net.Conn) {
	defer conn.Close()

	// SIZE CODE RSA_KEY
	// SIZE CODE COOKIE ( AES )

	allData, err := tor_server.ReadDataFromSizeHeader(conn, DATA_SIZE_SEGMENT_SIZE)
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	code, data := string(allData[:REQ_CODE_SIZE]), allData[REQ_CODE_SIZE:]

	if code == CODE_AUTH {
		cookie, aes, err := Auth(conn, data)
		if err != nil {
			log.Println("ERROR: ", err)
			return
		}

		log.Println("the cookie is", *cookie)

		// creating the new client, name will be set in login
		clients[*cookie] = Client{username: "", aesObj: aes}
		return
	}

	cookie, err := InitCookie(data[:COOKIE_SIZE])
	log.Println("got cookie: ", *cookie)
	if err != nil {
		log.Println("ERROR: ", err)
		// TODO: send error resp
		return
	}

	if _, inMap := clients[*cookie]; !inMap {
		log.Println("cookie not in map")
		// TODO: send error resp
		return
	}

	log.Println("cookie found in map")

	decrypted, err := clients[*cookie].aesObj.Decrypt(data)
	if err != nil {
		log.Println("ERROR: ", err)
		// TODO: send error resp
		return
	}

	log.Println("decrypted the data: ", decrypted)

	// chat server logic, created the response in json
	jsonResp := HandleRequests(code, decrypted)
	// encrypting the json with the aes key saved for the specific cookie
	encryptpedResp, err := clients[*cookie].aesObj.Encrypt([]byte(jsonResp))
	if err != nil {
		log.Println("ERROR: ", err)
		// TODO: send error resp
	}

	// the network requires a data size header
	serializedResp := fmt.Sprintf("%05d", len(encryptpedResp)) + string(encryptpedResp)
	conn.Write([]byte(serializedResp))
}

/*
	gets the request code and data, proccess it and returns the resp json
*/
func HandleRequests(code string, data []byte) string {
	var v interface{}

	switch code {
	case CODE_REGISTER:
		var req RegisterRequest
		err := json.Unmarshal(data, &req)
		if err != nil {
			return Marshal(ErrorResponse{"invalid json"})
		}

		v = Register(&req)
	default:
		v = ErrorResponse{"undefined request"}
	}

	return Marshal(v)
}

func Marshal(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		log.Println("ERROR: ", err)
		return ""
	}
	return string(s)
}

func Register(req *RegisterRequest) interface{} {
	// TODO: check if there is another one with this username
	err := RegisterDB(db, req.Username, req.Password)
	if err != nil {
		return ErrorResponse{"db error"}
	}

	return RegisterResponse{1}
}

func CloseDB() {
	db.Close()
}
