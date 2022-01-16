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
	REQ_CODE_SIZE          = 2
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
	if err != nil {
		log.Println("ERROR: ", err)
		return
	}

	if _, inMap := clients[*cookie]; !inMap {
		log.Println("cookie not in map")
		// TODO: send error resp
		return
	}

	log.Println("cookie found in map")
	decrypted, err := clients[*cookie].aesObj.Decrypt(data[COOKIE_SIZE:])
	if err != nil {
		log.Println("ERROR: ", err)
		// TODO: send error resp
		return
	}

	// chat server logic, created the response in json
	jsonResp := HandleRequests(code, decrypted, cookie)
	// encrypting the json with the aes key saved for the specific cookie
	encryptpedResp, err := clients[*cookie].aesObj.Encrypt([]byte(jsonResp))
	if err != nil {
		log.Println("ERROR: ", err)
		// TODO: send error resp
		return
	}

	// the network requires a data size header
	serializedResp := fmt.Sprintf("%05d", len(encryptpedResp)) + string(encryptpedResp)
	conn.Write([]byte(serializedResp))
}

/*
	gets the request code and data, proccess it and returns the resp json
*/
func HandleRequests(code string, data []byte, cookie *Cookie) string {
	var v interface{}

	// TODO: find a shorter way to do this with only one unmarshal call..
	switch code {
	case CODE_REGISTER:
		var req RegisterRequest
		err := json.Unmarshal(data, &req)
		if err != nil {
			return Marshal(MakeRegisterResponse(STATUS_FAILED))
		}

		v = Register(&req)

	case CODE_LOGIN:
		var req LoginRequest
		err := json.Unmarshal(data, &req)
		if err != nil {
			return Marshal(MakeLoginResponse(STATUS_FAILED))
		}

		v = Login(&req, cookie)
	default:
		v = MakeErrorResponse("undefined request")
	}

	return Marshal(v)
}

// less code when using this instead of directly calling marshal
func Marshal(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		log.Println("ERROR: ", err)
		return ""
	}
	return string(s)
}

// registers a user to the db if his username does not exists already
func Register(req *RegisterRequest) interface{} {
	// TODO: check if there is another one with this username
	err := RegisterDB(db, req.Username, req.Password)
	if err != nil {
		log.Println("ERROR: ", err)
		return MakeErrorResponse("db error")
	}

	return MakeRegisterResponse(STATUS_SUCCESS)
}

// logs the user into the system if his password and username are correct
func Login(req *LoginRequest, cookie *Cookie) interface{} {
	if CheckUsersPassword(db, req.Username, req.Password) {
		if entry, ok := clients[*cookie]; ok {
			entry.username = req.Username
		}
		return MakeLoginResponse(STATUS_SUCCESS)
	}
	return MakeLoginResponse(STATUS_FAILED)
}

// to use the global var "db" from main
func CloseDB() {
	db.Close()
}
