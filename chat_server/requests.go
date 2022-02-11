package chat_server

import "time"

// registers a user to the db if his username does not exists already
func Register(req *RegisterRequest) interface{} {
	if req.Username == "" || !isValidPassword(req.Password) {
		return GeneralResponse{CODE_REGISTER, STATUS_FAILED, "register error", "invalid password or username"}
	}

	ok, err := db.RegisterDB(req.Username, req.Password)
	logger.Info.Println("got the register ans and it is : ", ok)

	if err != nil {
		logger.Err.Println(err)
		return MakeErrorResponse("db error")
	}

	if ok {
		return GeneralResponse{CODE_REGISTER, STATUS_SUCCESS, "register success", "registered successfuly"}
	}
	return GeneralResponse{CODE_REGISTER, STATUS_FAILED, "register error", "username already exists"}
}

// logs the user into the system if his password and username are correct
func Login(req *LoginRequest, client *Client) interface{} {
	if db.CheckUsersPassword(req.Username, req.Password) {
		if IsUserLoggedin(req.Username) {
			return GeneralResponse{CODE_LOGIN, STATUS_FAILED, "login error", "user already logged in"}
		}
		client.username = req.Username
		return GeneralResponse{CODE_LOGIN, STATUS_SUCCESS, "login success", "logged in successfuly"}
	}
	return GeneralResponse{CODE_LOGIN, STATUS_FAILED, "login error", "username or password is not correct"}
}

// creates a new room and joins the client as the admin
func CreateChatRoom(req *CreateChatRoomRequest, client *Client) interface{} {
	adminID, err := db._getUserID(client.username)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "create room error", "user not exists"}
	}
	ok, err := db.CreateChatRoomDB(req.RoomName, req.Password, adminID)

	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "create room error", "room name might be existing already"}
	}
	if !ok {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "create room error", "room name might be existing already"}
	}

	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "create room error", "failed to create room"}
	}

	// add admin to the chat db members
	ok, err = db.JoinChatRoomDB(roomID, req.Password, adminID, STATE_NORMAL)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "create room error", "cant enter admin to room"}
	}
	if !ok {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "create room error", "user in ban or with invalid password"}
	}

	// initing the live room, adding the client to it
	chatRoomsMx.Lock()
	chatRooms[req.RoomName] = &ChatRoom{onlineMembers: make([]*Client, 0)}
	chatRooms[req.RoomName].onlineMembers = append(chatRooms[req.RoomName].onlineMembers, client)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_SUCCESS, "create room success", "entered successfuly"}
}

func DeleteChatRoom(req *DeleteChatRoomRequest, client *Client) interface{} {
	//TODO: check credentials(in all things that needs that)
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "delete room error", "room doesn't exists"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "delete room error", "user not exists"}
	}

	if !(db._isAdminOfRoom(roomID, adminID) && db.isRoomPassword(roomID, req.Password)) {
		logger.Err.Println("Wrong credentials")
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "delete room error", "Wrong room credentials(username or/and password is/are wrong)"}
	}

	ok, err := db.DeleteChatRoomDB(roomID, req.Password, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "delete room error", "deleting room-related things failed"}
	}
	if !ok {
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "delete room error", "something went wrong"}
	}

	chatRoomsMx.Lock()
	delete(chatRooms, req.RoomName)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_SUCCESS, "delete room success", "room deleted successfuly"}
}

func JoinChatRoom(req *JoinChatRoomRequest, client *Client, state int) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "join room error", "room not exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "join room error", "user not exists"}
	}

	ok, err := db.JoinChatRoomDB(roomID, req.Password, userID, state)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "join room error", "something went wrong"}
	}
	if !ok {
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "join room error", "user in ban or with wrong password"}
	}

	chatRoomsMx.Lock()
	chatRooms[req.RoomName].onlineMembers = append(chatRooms[req.RoomName].onlineMembers, client)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_SUCCESS, "join room success", "joined room successfuly"}
}

func KickFromChatRoom(req *KickFromChatRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "kick user from room error", "room doesn't exists"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "kick user from room error", "action performer not exists"}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "kick user from room error", "user to kick doesn't exists"}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to kick")
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "kick user from room error", "you are not the admin"}
	}

	ok, err := db.KickFromChatRoomDB(roomID, userID, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "kick user from room error", "something went wrong"}
	}
	if !ok {
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "kick user from room error", "something went wrong"}
	}

	RemoveMemberFromChat(req.RoomName, req.Username)

	return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_SUCCESS, "kick user from room success", "kicked user from room successfuly"}
}

func BanFromChatRoom(req *BanFromChatRoomRequest, client *Client) interface{} {
	//in case that user not in room so we add him and change the state to STATE_BAN
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "ban user from room error", "room doesn't exists"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "ban user from room error", "action performer not exists"}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "ban user from room error" , "user to ban doesn't exists"}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to ban")
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "ban user from room error", "you are not the admin"}
	}

	ok, err := db.BanFromChatRoomDB(roomID, userID, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "ban user from room error", "something went wrong"}
	}
	if !ok {
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "ban user from room error", "something went wrong"}
	}

	RemoveMemberFromChat(req.RoomName, req.Username)

	return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_SUCCESS, "ban user from room success", "banned user from room successfuly"}
}

func UnBanFromChatRoom(req *UnBanFromChatRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "unban user from room error", "room doesn't exists"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "unban user from room error", "action performer not exists"}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "unban user from room error", "user to unban doesn't exists"}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to unBan")
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "unban user from room error", "you are not the admin"}
	}

	ok, err := db.UnBanFromChatRoomDB(roomID, userID, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "unban user from room error", "something went wrong"}
	}
	if !ok {
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "unban user from room error", "something went wrong"}
	}

	return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_SUCCESS, "unban user from room success", "unbanned user from room successfuly"}
}

func SendMessage(req *SendMessageRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "send message error", "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "send message error", "sender doesn't exists"}
	}

	inRoom := db._isUserInRoom(roomID, roomID)
	if !inRoom {
		logger.Err.Println("user not in room")
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "send message error", "sender not in room"}
	}

	inBan := db._isUserInBan(roomID, userID)
	if inBan {
		logger.Err.Println("user in ban")
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "send message error", "sender in ban"}
	}

	messageTime := time.Now()
	ok, err := db.SendMessageDB(req.Content, roomID, userID, messageTime)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "send message error", "something went wrong"}
	}
	if !ok {
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "send message error", "something went wrong"}
	}

	newMsg := Message{req.RoomName, req.Content, client.username, messageTime}

	// notifying every member of the room about the new message
	for _, member := range chatRooms[req.RoomName].onlineMembers {
		member.Lock()
		member.messages = append(member.messages, newMsg)
		member.Unlock()
		member.cond.Signal()
	}

	return GeneralResponse{CODE_SEND_MESSAGE, STATUS_SUCCESS, "send message success", "messages sent successfuly"}
}

func UpdateMessages(req *UpdateMessagesRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UPDATE, STATUS_FAILED, "update messages error", "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UPDATE, STATUS_FAILED, "update messages error", "user not exists"}
	}

	inRoom := db._isUserInRoom(roomID, roomID)
	if !inRoom {
		logger.Err.Println("user not in room")
		return GeneralResponse{CODE_UPDATE, STATUS_FAILED, "update messages error", "user not in room"}
	}

	if len(client.messages) != 0 {
		messages := client.messages
		client.messages = client.messages[:0] // cleaning the messages
		return UpdateMessagesResponse{GeneralResponse{CODE_UPDATE, STATUS_SUCCESS, "update messages success", "updated messages successfuly"}, messages}
	}

	client.Lock()
	client.cond.Wait() // waiting for a message

	messages := client.messages
	client.messages = client.messages[:0] // cleaning the messages

	client.Unlock()

	return UpdateMessagesResponse{GeneralResponse{CODE_UPDATE, STATUS_SUCCESS, "update messages success", "updated messages successfuly"}, messages}
}

func LoadMessages(req *LoadRoomsMessagesRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_LOAD_MESSAGES, STATUS_FAILED, "load messages error", "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_LOAD_MESSAGES, STATUS_FAILED, "load messages error", "user doesn't exists"}
	}

	inRoom := db._isUserInRoom(roomID, roomID)
	if !inRoom {
		logger.Err.Println("user not in room")
		return GeneralResponse{CODE_LOAD_MESSAGES, STATUS_FAILED, "load messages error", "user not in room"}
	}

	messages, err := db.LoadLastMessages(roomID, req.Amount, req.Offset)
	logger.Info.Println(messages)

	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_LOAD_MESSAGES, STATUS_FAILED, "load messages error", "something went wrong"}
	}
	return LoadRoomsMessagesResponse{GeneralResponse{CODE_LOAD_MESSAGES, STATUS_SUCCESS, "load messages success", "load messages successfuly"}, messages}
}

func RemoveMemberFromChat(roomName string, username string) {
	chatRoomsMx.Lock()

	for i, v := range chatRooms[roomName].onlineMembers {
		if v.username == username {
			// removing the username by appending without it
			chatRooms[roomName].onlineMembers = append(chatRooms[roomName].onlineMembers[:i],
				chatRooms[roomName].onlineMembers[i+1:]...)
		}
	}

	chatRoomsMx.Unlock()
}

func IsUserLoggedin(username string) bool {
	for _, client := range clients {
		if client.username == username {
			return true
		}
	}
	return false
}
