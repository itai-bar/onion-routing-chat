package chat_server

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// registers a user to the db if his username does not exists already
func Register(req *RegisterRequest) interface{} {
	if req.Username == "" || !isValidPassword(req.Password) {
		return GeneralResponse{CODE_REGISTER, STATUS_FAILED, "invalid password or username"}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_REGISTER, STATUS_FAILED, "something went wrong"}
	}

	ok, err := db.RegisterDB(req.Username, string(hash))

	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_REGISTER, STATUS_FAILED, "something went wrong"}
	}

	if ok {
		return GeneralResponse{CODE_REGISTER, STATUS_SUCCESS, "registered successfuly"}
	}
	return GeneralResponse{CODE_REGISTER, STATUS_FAILED, "username already exists"}
}

// logs the user into the system if his password and username are correct
func Login(req *LoginRequest, client *Client) interface{} {
	if db.CheckUsersPassword(req.Username, req.Password) {
		if IsUserLoggedin(req.Username) {
			return GeneralResponse{CODE_LOGIN, STATUS_FAILED, "user already logged in"}
		}
		client.username = req.Username
		return GeneralResponse{CODE_LOGIN, STATUS_SUCCESS, "logged in successfuly"}
	}
	return GeneralResponse{CODE_LOGIN, STATUS_FAILED, "username or password is not correct"}
}

func Logout(client *Client) interface{} {
	found := false
	username := client.username

	clientsMx.Lock()
	for cookie, c := range clients {
		if c == client {
			clients[cookie].username = "" // not logged
			found = true
			break
		}
	}
	clientsMx.Unlock()

	if !found {
		return GeneralResponse{CODE_LOGOUT, STATUS_FAILED, "user was not logged in"}
	}

	for chatName := range chatRooms {
		SetOfflineUserInRoom(chatName, username)
	}

	return GeneralResponse{CODE_LOGOUT, STATUS_SUCCESS, "logged out successfuly"}
}

// creates a new room and joins the client as the admin
func CreateChatRoom(req *CreateChatRoomRequest, client *Client) interface{} {
	if req.Password == "" {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "password cant be empty!"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "user not exists"}
	}
	ok, err := db.CreateChatRoomDB(req.RoomName, req.Password, adminID)

	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "room name might be existing already"}
	}
	if !ok {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "room name might be existing already"}
	}

	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "failed to create room"}
	}

	// add admin to the chat db members
	ok, err = db.JoinChatRoomDB(roomID, req.Password, adminID, STATE_NORMAL)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "cant enter admin to room"}
	}
	if !ok {
		return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_FAILED, "user in ban or with invalid password"}
	}

	// initing the live room, adding the client to it
	chatRoomsMx.Lock()
	chatRooms[req.RoomName] = &ChatRoom{onlineMembers: make([]*Client, 0)}
	//SetOnlineUserInRoom(req.RoomName, client.username)
	chatRooms[req.RoomName].onlineMembers = append(chatRooms[req.RoomName].onlineMembers, client)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_CREATE_CHAT_ROOM, STATUS_SUCCESS, "entered successfuly"}
}

func DeleteChatRoom(req *DeleteChatRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "room doesn't exists"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "user not exists"}
	}

	if !(db._isAdminOfRoom(roomID, adminID) && db.isRoomPassword(roomID, req.Password)) {
		logger.Err.Println("Wrong credentials")
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "Wrong room credentials(username or/and password is/are wrong)"}
	}

	ok, err := db.DeleteChatRoomDB(roomID, req.Password, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "deleting room-related things failed"}
	}
	if !ok {
		return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_FAILED, "something went wrong"}
	}

	chatRoomsMx.Lock()
	delete(chatRooms, req.RoomName)
	chatRoomsMx.Unlock()

	return GeneralResponse{CODE_DELETE_CHAT_ROOM, STATUS_SUCCESS, "room deleted successfuly"}
}

func JoinChatRoom(req *JoinChatRoomRequest, client *Client, state int) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "room not exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "user not exists"}
	}

	if db._isUserInRoom(roomID, userID) {
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "user already in room"}
	}

	ok, err := db.JoinChatRoomDB(roomID, req.Password, userID, state)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "something went wrong"}
	}
	if !ok {
		return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_FAILED, "user in ban or with wrong password"}
	}

	SetOnlineUserInRoom(req.RoomName, client.username)

	return GeneralResponse{CODE_JOIN_CHAT_ROOM, STATUS_SUCCESS, "joined room successfuly"}
}

func KickFromChatRoom(req *KickFromChatRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "room doesn't exists"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "action performer not exists"}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "user to kick doesn't exists"}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to kick")
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "you are not the admin"}
	}

	ok, err := db.KickFromChatRoomDB(roomID, userID, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "something went wrong"}
	}
	if !ok {
		return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_FAILED, "something went wrong"}
	}

	SetOfflineUserInRoom(req.RoomName, req.Username)

	return GeneralResponse{CODE_KICK_FROM_CHAT_ROOM, STATUS_SUCCESS, "kicked user from room successfuly"}
}

func BanFromChatRoom(req *BanFromChatRoomRequest, client *Client) interface{} {
	//in case that user not in room so we add him and change the state to STATE_BAN
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "room doesn't exists"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "action performer not exists"}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "user to ban doesn't exists"}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to ban")
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "you are not the admin"}
	}

	ok, err := db.BanFromChatRoomDB(roomID, userID, adminID)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "something went wrong"}
	}
	if !ok {
		return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_FAILED, "something went wrong"}
	}

	SetOfflineUserInRoom(req.RoomName, req.Username)

	return GeneralResponse{CODE_BAN_FROM_CHAT_ROOM, STATUS_SUCCESS, "banned user from room successfuly"}
}

func UnBanFromChatRoom(req *UnBanFromChatRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "room doesn't exists"}
	}

	adminID, err := db._getUserID(client.username)
	if err != nil || adminID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "action performer not exists"}
	}

	userID, err := db._getUserID(req.Username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "user to unban doesn't exists"}
	}

	if !db._isAdminOfRoom(roomID, adminID) {
		logger.Err.Println("not admin trying to unBan")
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "you are not the admin"}
	}

	ok, err := db.UnBanFromChatRoomDB(roomID, userID, adminID)
	if err != nil || !ok {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_FAILED, "something went wrong"}
	}

	return GeneralResponse{CODE_UNBAN_FROM_CHAT_ROOM, STATUS_SUCCESS, "unbanned user from room successfuly"}
}

func SendMessage(req *SendMessageRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "sender doesn't exists"}
	}

	inRoom := db._isUserInRoom(roomID, userID)
	if !inRoom {
		logger.Err.Println("user not in room")
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "sender not in room"}
	}

	inBan := db._isUserInBan(roomID, userID)
	if inBan {
		logger.Err.Println("user in ban")
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "sender in ban"}
	}

	messageTime := time.Now()
	ok, err := db.SendMessageDB(req.Content, roomID, userID, messageTime)
	if err != nil || !ok {
		logger.Err.Println(err)
		return GeneralResponse{CODE_SEND_MESSAGE, STATUS_FAILED, "something went wrong"}
	}

	newMsg := Message{req.RoomName, req.Content, client.username, messageTime}

	// notifying every member of the room about the new message
	logger.Info.Println(chatRooms[req.RoomName].onlineMembers)
	for _, member := range chatRooms[req.RoomName].onlineMembers {
		member.Lock()
		member.messages = append(member.messages, newMsg)
		member.Unlock()
		member.cond.Signal()
	}

	return GeneralResponse{CODE_SEND_MESSAGE, STATUS_SUCCESS, "messages sent successfuly"}
}

func UpdateMessages(req *UpdateMessagesRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UPDATE, STATUS_FAILED, "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_UPDATE, STATUS_FAILED, "user not exists"}
	}

	inRoom := db._isUserInRoom(roomID, userID)
	if !inRoom {
		logger.Err.Println("user not in room")
		return GeneralResponse{CODE_UPDATE, STATUS_FAILED, "user not in room"}
	}

	if len(client.messages) != 0 {
		messages := client.messages
		client.messages = client.messages[:0] // cleaning the messages
		return UpdateMessagesResponse{GeneralResponse{CODE_UPDATE, STATUS_SUCCESS, "updated messages successfuly"}, messages}
	}

	client.Lock()
	client.cond.Wait() // waiting for a message

	messages := client.messages
	client.messages = client.messages[:0] // cleaning the messages

	client.Unlock()

	return UpdateMessagesResponse{GeneralResponse{CODE_UPDATE, STATUS_SUCCESS, "updated messages successfuly"}, messages}
}

func LoadMessages(req *LoadRoomsMessagesRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_LOAD_MESSAGES, STATUS_FAILED, "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_LOAD_MESSAGES, STATUS_FAILED, "user doesn't exists"}
	}

	inRoom := db._isUserInRoom(roomID, userID)
	if !inRoom {
		logger.Err.Println("user not in room")
		return GeneralResponse{CODE_LOAD_MESSAGES, STATUS_FAILED, "user not in room"}
	}

	messages, err := db.LoadLastMessages(roomID, req.Amount, req.Offset)
	logger.Info.Println(messages)

	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_LOAD_MESSAGES, STATUS_FAILED, "something went wrong"}
	}
	return LoadRoomsMessagesResponse{GeneralResponse{CODE_LOAD_MESSAGES, STATUS_SUCCESS, "load messages successfuly"}, messages}
}

func GetRooms() interface{} {
	rooms, err := db.GetRoomsDB()
	if err != nil {
		return GeneralResponse{CODE_GET_ROOMS, STATUS_FAILED, "something went wrong"}
	}
	return GetRoomsResponse{GeneralResponse{CODE_GET_ROOMS, STATUS_SUCCESS, "got rooms successfuly"}, rooms}
}

func SetOnlineUserInRoom(roomName string, username string) {
	for _, client := range clients {
		if client.username == username {
			SetOfflineUserInRoom(roomName, username)
			chatRoomsMx.Lock()
			chatRooms[roomName].onlineMembers = append(chatRooms[roomName].onlineMembers, client)
			chatRoomsMx.Unlock()
		}
	}
}

func IsUserInRoom(req *UserInRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_IS_USER_IN_ROOM, STATUS_FAILED, "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_IS_USER_IN_ROOM, STATUS_FAILED, "user doesn't exists"}
	}

	inRoom := db._isUserInRoom(roomID, userID)
	if inRoom {
		SetOnlineUserInRoom(req.RoomName, client.username)
		return GeneralResponse{CODE_IS_USER_IN_ROOM, STATUS_FAILED, "user in room"}
	}
	return GeneralResponse{CODE_IS_USER_IN_ROOM, STATUS_SUCCESS, "user not in room"}
}

func CancelUpdate(req *CancelUpdateRequest, client *Client) interface{} {
	for _, clientIterator := range clients {
		if clientIterator.username == client.username {
			clientIterator.cond.Signal()
		}
	}

	if req.RoomName == "" {
		return GeneralResponse{CODE_CANCEL_UPDATE, STATUS_FAILED, "no room for release cancel"}
	}

	if !SetOfflineUserInRoom(req.RoomName, client.username) {
		return GeneralResponse{CODE_CANCEL_UPDATE, STATUS_FAILED, "couldn't release the update request"}
	}

	return GeneralResponse{CODE_CANCEL_UPDATE, STATUS_SUCCESS, "released update request successfully"}
}

func LeaveRoom(req *LeaveRoomRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_LEAVE_ROOM, STATUS_FAILED, "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_LEAVE_ROOM, STATUS_FAILED, "user doesn't exists"}
	}

	if db._isUserInRoom(roomID, userID) {
		success, err := db.LeaveRoomDB(roomID, userID)
		if err != nil || !success {
			return GeneralResponse{CODE_LEAVE_ROOM, STATUS_FAILED, "something went wrong, maybe user not in room"}
		}
		SetOfflineUserInRoom(req.RoomName, client.username)
		return GeneralResponse{CODE_LEAVE_ROOM, STATUS_SUCCESS, "quited successfully *in case user not in ban..*"}
	}
	return GeneralResponse{CODE_LEAVE_ROOM, STATUS_FAILED, "user not in room"}
}

func SetOfflineUserInRoom(roomName string, username string) bool {
	if roomName == "" {
		return false
	}

	chatRoomsMx.Lock()
	defer chatRoomsMx.Unlock()

	// this function can be called on a closed room in bad timing
	// which will lead to dereferencing a nil pointer
	if _, roomExists := chatRooms[roomName]; !roomExists {
		return false
	}

	//TODO: add request to get deleted from clients(that what cause us to use onlineMember.username == "" in next lines)
	for i, onlineMember := range chatRooms[roomName].onlineMembers {
		if onlineMember.username == username || onlineMember.username == "" { // wanted user || not logged in
			// removing the username by appending without it
			chatRooms[roomName].onlineMembers = append(chatRooms[roomName].onlineMembers[:i],
				chatRooms[roomName].onlineMembers[i+1:]...)
		}
	}

	return true
}

func GetRoomMembers(req *GetRoomMembersRequest, client *Client) interface{} {
	roomID, err := db._getChatRoomID(req.RoomName)
	if err != nil || roomID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_GET_ROOM_MEMBERS, STATUS_FAILED, "room doesn't exists"}
	}

	userID, err := db._getUserID(client.username)
	if err != nil || userID == WITHOUT_ID {
		logger.Err.Println(err)
		return GeneralResponse{CODE_GET_ROOM_MEMBERS, STATUS_FAILED, "user doesn't exists"}
	}

	if !db._isUserInRoom(roomID, userID) {
		return GeneralResponse{CODE_GET_ROOM_MEMBERS, STATUS_FAILED, "user not in room"}
	}

	var onlineMembersNames []string
	logger.Info.Println("online members before are:", onlineMembersNames)
	for _, onlineMember := range chatRooms[req.RoomName].onlineMembers {
		logger.Info.Println("Adding", onlineMember.username, "to online members")
		onlineMembersNames = append(onlineMembersNames, onlineMember.username)
	}
	logger.Info.Println("online members after are:", onlineMembersNames)

	offlineMembersNames, err := db.GetOfflineMembersInRoomDB(roomID, onlineMembersNames)
	if err != nil {
		logger.Err.Println(err)
		return GeneralResponse{CODE_GET_ROOM_MEMBERS, STATUS_FAILED, "something went wrong"}
	}

	adminName := db._getAdminRoom(roomID)

	return GetRoomMembersResponse{GeneralResponse{CODE_GET_ROOM_MEMBERS, STATUS_SUCCESS, "got room members successfully"}, adminName, onlineMembersNames, offlineMembersNames}
}

func IsUserLoggedin(username string) bool {
	for _, client := range clients {
		if client.username == username {
			return true
		}
	}
	return false
}
