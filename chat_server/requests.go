package chat_server

// registers a user to the db if his username does not exists already
func Register(req *RegisterRequest) interface{} {
	if req.Username == "" || !isValidPassword(req.Password) {
		return GeneralResponse{CODE_REGISTER, STATUS_FAILED}
	}

	ok, err := db.RegisterDB(req.Username, req.Password)
	logger.Info.Println("got the register ans and it is : ", ok)

	if err != nil {
		logger.Err.Println(err)
		return MakeErrorResponse("db error")
	}

	if ok {
		return GeneralResponse{CODE_REGISTER, STATUS_SUCCESS}
	}
	return GeneralResponse{CODE_REGISTER, STATUS_FAILED}
}

// logs the user into the system if his password and username are correct
func Login(req *LoginRequest, client *Client) interface{} {
	if db.CheckUsersPassword(req.Username, req.Password) {
		client.username = req.Username
		return GeneralResponse{CODE_LOGIN, STATUS_SUCCESS}
	}
	return GeneralResponse{CODE_LOGIN, STATUS_FAILED}
}

// creates a new room and joins the client as the admin
func CreateChatRoom(req *CreateChatRoomRequest, client *Client) interface{} {
	adminID, err := db._getUserID(client.username)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}
	ok, err := db.CreateChatRoomDB(req.RoomName, req.Password, adminID)

	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}

	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}

	// add admin to the chat db members
	ok, err = db.JoinChatRoomDB(roomID, req.Password, adminID, STATE_NORMAL)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}

	// initing the live room, adding the client to it
	chatRoomsMx.Lock()
	chatRooms[req.RoomName] = &ChatRoom{onlineMembers: make([]*Client, 0)}
	chatRooms[req.RoomName].onlineMembers = append(chatRooms[req.RoomName].onlineMembers, client)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_SUCCESS}
}

func DeleteChatRoom(req *DeleteChatRoomRequest, client *Client) interface{} {
	//TODO: check credentials(in all things that needs that)
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED}
	}

	if !(db._isAdminOfRoom(roomID, adminID) && db.isRoomPassword(roomID, req.Password)) {
		logger.Err.Println("Wrong credentials")
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED}
	}

	ok, err := db.DeleteChatRoomDB(roomID, req.Password, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED}
	}

	chatRoomsMx.Lock()
	delete(chatRooms, req.RoomName)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_SUCCESS}
}

func JoinChatRoom(req *JoinChatRoomRequest, client *Client, state int) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED}
	}

	ok, err := db.JoinChatRoomDB(roomID, req.Password, userID, state)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED}
	}

	chatRoomsMx.Lock()
	chatRooms[req.RoomName].onlineMembers = append(chatRooms[req.RoomName].onlineMembers, client)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_SUCCESS}
}

func KickFromChatRoom(req *KickFromChatRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to kick")
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	ok, err := db.KickFromChatRoomDB(roomID, userID, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	RemoveMemberFromChat(req.RoomName, req.Username)

	return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_SUCCESS}
}

func BanFromChatRoom(req *BanFromChatRoomRequest, client *Client) interface{} {
	//in case that user not in room so we add him and change the state to STATE_BAN
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to ban")
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	ok, err := db.BanFromChatRoomDB(roomID, userID, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	RemoveMemberFromChat(req.RoomName, req.Username)

	return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_SUCCESS}
}

func UnBanFromChatRoom(req *UnBanFromChatRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to unBan")
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	ok, err := db.UnBanFromChatRoomDB(roomID, userID, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_SUCCESS}
}

func SendMessage(req *SendMessageRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED}
	}

	inRoom := db._isUserInRoom(roomID, roomID)
	if !inRoom {
		logger.Err.Println("user not in room")
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED}
	}

	inBan := db._isUserInBan(roomID, userID)
	if inBan {
		logger.Err.Println("user in ban")
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED}
	}

	ok, err := db.SendMessageDB(req.Content, roomID, userID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED}
	}

	// TODO: walk through all client in chosen room and update them about the message!

	return GeneralResponse{CODE_SEND_MESSAGE, STATUS_SUCCESS}
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
