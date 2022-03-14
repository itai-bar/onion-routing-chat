package chat_server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"torbasedchat/pkg/tor_aes"
	"torbasedchat/pkg/tor_logger"
	"torbasedchat/pkg/tor_server"
)

const (
	DATA_SIZE_SEGMENT_SIZE = 5
	REQ_CODE_SIZE          = 2
)

type Client struct {
	sync.Mutex
	username string
	aesObj   *tor_aes.Aes
	messages []Message
	cond     *sync.Cond
}

type ChatRoom struct {
	onlineMembers []*Client
}

var clients map[Cookie]*Client
var chatRooms map[string]*ChatRoom

var db *ChatDb

var clientsMx sync.Mutex
var chatRoomsMx sync.Mutex
var dbMx sync.Mutex

var logger *tor_logger.TorLogger

func init() {
	var err error

	logger = tor_logger.NewTorLogger(os.Getenv("CHAT_LOG"))

	clients = make(map[Cookie]*Client)
	chatRooms = make(map[string]*ChatRoom)

	sqlDb, err := InitDb("/app/db.sqlite")
	if err != nil {
		log.Fatal("ERROR: ", err)
	}

	db = InitChatDb(sqlDb)
	db.LoadRoomsFromDB()
}

/*
	handles the connection with every client

	conn net.Conn: connection with a client
*/
func HandleClient(conn net.Conn) {
	defer conn.Close()

	allData, err := tor_server.ReadDataFromSizeHeader(conn, DATA_SIZE_SEGMENT_SIZE)
	if err != nil {
		logger.Err.Println(err)
		return
	}

	code, data := string(allData[:REQ_CODE_SIZE]), allData[REQ_CODE_SIZE:]

	if code == CODE_AUTH {
		cookie, aes, err := Auth(conn, data)
		if err != nil {
			logger.Err.Println(err)
			return
		}

		logger.Info.Println("the cookie is", *cookie)

		// creating the new client, name will be set in login

		clientsMx.Lock()
		client := &Client{username: "", aesObj: aes}
		client.cond = sync.NewCond(client)
		clients[*cookie] = client
		clientsMx.Unlock()

		return
	}

	// client found in map

	cookie, err := InitCookie(data[:COOKIE_SIZE])
	if err != nil {
		logger.Err.Println(err)
		return
	}

	clientsMx.Lock()
	currentClient, inMap := clients[*cookie]
	clientsMx.Unlock()

	if !inMap {
		logger.Err.Println("cookie not in map")
		return
	}

	decrypted, err := currentClient.aesObj.Decrypt(data[COOKIE_SIZE:])
	logger.Info.Printf("client: %s. req: %s+%s", currentClient.username, code, string(decrypted))

	if err != nil {
		logger.Err.Println(err)
		return
	}

	// chat server logic, created the response in json
	jsonResp := HandleRequests(code, decrypted, currentClient)
	SendResponse(conn, jsonResp, currentClient)
}

func SendResponse(conn net.Conn, resp string, client *Client) {

	// encrypting the json with the aes key saved for the specific cookie
	encryptpedResp, err := client.aesObj.Encrypt([]byte(resp))
	if err != nil {
		logger.Err.Println(err)
		conn.Write(nil)
	}

	// the network requires a data size header
	serializedResp := fmt.Sprintf("%05d", len(encryptpedResp)) + string(encryptpedResp)
	conn.Write([]byte(serializedResp))
}

/*
	gets the request code and data, proccess it and returns the resp json
*/
func HandleRequests(code string, data []byte, client *Client) string {
	var resp interface{}

	// those requests require to be logged in
	if code != CODE_REGISTER && code != CODE_LOGIN {
		if client.username == "" {
			// not logged in
			return Marshal(GeneralResponse{code, STATUS_FAILED, "Must log in to use this request"})
		}
	}

	switch code {
	case CODE_REGISTER:
		var req RegisterRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = Register(&req)

	case CODE_LOGIN:
		var req LoginRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = Login(&req, client)

	case CODE_LOGOUT:
		resp = Logout(client)

	case CODE_CREATE_CHAT_ROOM:
		var req CreateChatRoomRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = CreateChatRoom(&req, client)

	case CODE_DELETE_CHAT_ROOM:
		var req DeleteChatRoomRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = DeleteChatRoom(&req, client)

	case CODE_JOIN_CHAT_ROOM:
		var req JoinChatRoomRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = JoinChatRoom(&req, client, STATE_NORMAL)

	case CODE_KICK_FROM_CHAT_ROOM:
		var req KickFromChatRoomRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = KickFromChatRoom(&req, client)

	case CODE_BAN_FROM_CHAT_ROOM:
		var req BanFromChatRoomRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = BanFromChatRoom(&req, client)

	case CODE_UNBAN_FROM_CHAT_ROOM:
		var req UnBanFromChatRoomRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = UnBanFromChatRoom(&req, client)

	case CODE_SEND_MESSAGE:
		var req SendMessageRequest
		if errMsg := Unmarshal(CODE_SEND_MESSAGE, data, &req); errMsg != "" {
			return errMsg
		}

		resp = SendMessage(&req, client)

	case CODE_UPDATE:
		var req UpdateMessagesRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = UpdateMessages(&req, client)

	case CODE_LOAD_MESSAGES:
		var req LoadRoomsMessagesRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		}

		resp = LoadMessages(&req, client)

	case CODE_GET_ROOMS:
		resp = GetRooms()
	
	case CODE_IS_USER_IN_ROOM:
		var req UserInRoomRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		} 
		resp = IsUserInRoom(&req, client)

	case CODE_CANCEL_UPDATE:
		resp = CancelUpdate(client)
	
	case CODE_LEAVE_ROOM:
		var req LeaveRoomRequest
		if errMsg := Unmarshal(code, data, &req); errMsg != "" {
			return errMsg
		} 
		resp = LeaveRoom(&req, client)

	default:
		resp = GeneralResponse{code, STATUS_FAILED, "undefined request"}
	}
	return Marshal(resp)
}

// less code when using this instead of directly calling marshal
func Marshal(v interface{}) string {
	s, err := json.Marshal(v)
	if err != nil {
		logger.Info.Println(err)
		return ""
	}
	return string(s)
}

func Unmarshal(code string, data []byte, v interface{}) string {
	err := json.Unmarshal(data, &v)
	if err != nil {
		logger.Info.Println(err)
		return Marshal(GeneralResponse{code, STATUS_FAILED, "invalid request arguments"})
	}
	return ""
}
