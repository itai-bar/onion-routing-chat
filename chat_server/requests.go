package chat_server

// registers a user to the db if his username does not exists already
func Register(req *RegisterRequest) interface{} {
	// username must not be empty (thats how we check if a user is logged or not)
	// TODO: if there any password requirements thats the place to add them..
	if req.Username == "" {
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
	ok, err := db.CreateChatRoomDB(req.Name, req.Password, client.username)

	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}

	// add admin to the chat db members
	ok, err = db.JoinChatRoomDB(req.Name, req.Password, client.username, STATE_NORMAL)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED}
	}

	// initing the live room, adding the client to it
	chatRoomsMx.Lock()
	chatRooms[req.Name] = &ChatRoom{onlineMembers: make([]*Client, 0)}
	chatRooms[req.Name].onlineMembers = append(chatRooms[req.Name].onlineMembers, client)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_SUCCESS}
}

func DeleteChatRoom(req *DeleteChatRoomRequest, client *Client) interface{} {
	ok, err := db.DeleteChatRoomDB(req.Name, req.Password, client.username)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED}
	}

	chatRoomsMx.Lock()
	delete(chatRooms, req.Name)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_SUCCESS}
}

func JoinChatRoom(req *JoinChatRoomRequest, client *Client, state bool) interface{} {
	ok, err := db.JoinChatRoomDB(req.Name, req.Password, client.username, state)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED}
	}

	chatRoomsMx.Lock()
	chatRooms[req.Name].onlineMembers = append(chatRooms[req.Name].onlineMembers, client)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_SUCCESS}
}

func KickFromChatRoom(req *KickFromChatRoomRequest, client *Client) interface{} {
	ok, err := db.KickFromChatRoomDB(req.Name, req.Username, client.username)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	// chatRoomsMx.Lock()
	//TODO:remove req.Username from charRooms[req.Name].onlineMembers
	// chatRoomsMx.Unlock()
	return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_SUCCESS}
}

func BanFromChatRoom(req *BanFromChatRoomRequest, client *Client) interface{} {
	//in case that user not in room so we add him and change the state to STATE_BAN
	ok, err := db.BanFromChatRoomDB(req.Name, req.Username, client.username)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}
	if !ok {
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED}
	}

	// chatRoomsMx.Lock()
	//TODO:in case that user already at room so remove req.Username from charRooms[req.Name].onlineMembers
	// chatRoomsMx.Unlock()
	return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_SUCCESS}
}
