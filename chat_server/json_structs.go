package chat_server

import "time"

const (
	STATUS_SUCCESS = 1
	STATUS_FAILED  = 0

	STATE_NORMAL = 0
	STATE_BAN    = 1

	CODE_AUTH                 = "00"
	CODE_UPDATE               = "01"
	CODE_LOGIN                = "02"
	CODE_REGISTER             = "03"
	CODE_LOGOUT               = "04"
	CODE_CREATE_CHAT_ROOM     = "05"
	CODE_DELETE_CHAT_ROOM     = "06"
	CODE_JOIN_CHAT_ROOM       = "07"
	CODE_KICK_FROM_CHAT_ROOM  = "08"
	CODE_BAN_FROM_CHAT_ROOM   = "09"
	CODE_UNBAN_FROM_CHAT_ROOM = "10"
	CODE_SEND_MESSAGE         = "11"
	CODE_LOAD_MESSAGES        = "12"
	CODE_GET_ROOMS            = "13"
	CODE_IS_USER_IN_ROOM      = "14"
	CODE_CANCEL_UPDATE        = "15"
	CODE_LEAVE_ROOM            = "16"
	CODE_ERR                  = "99"
)

type Message struct {
	RoomName string    `json:"roomName"`
	Content  string    `json:"content"`
	Sender   string    `json:"sender"`
	Time     time.Time `json:"time"`
}

type GeneralResponse struct {
	Code   string `json:"code"`
	Status int    `json:"status"`
	Info   string `json:"info"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateChatRoomRequest struct {
	RoomName string `json:"roomName"`
	Password string `json:"password"`
}

type JoinChatRoomRequest struct {
	RoomName string `json:"roomName"`
	Password string `json:"password"`
}

type DeleteChatRoomRequest struct {
	RoomName string `json:"roomName"`
	Password string `json:"password"`
}

type KickFromChatRoomRequest struct {
	Username string `json:"username"`
	RoomName string `json:"roomName"`
}

type BanFromChatRoomRequest struct {
	Username string `json:"username"`
	RoomName string `json:"roomName"`
}

type UnBanFromChatRoomRequest struct {
	Username string `json:"username"`
	RoomName string `json:"roomName"`
}

type SendMessageRequest struct {
	Content  string `json:"content"`
	RoomName string `json:"roomName"`
}

type UpdateMessagesRequest struct {
	RoomName string `json:"roomName"`
}

type UpdateMessagesResponse struct {
	GeneralResponse
	Messages []Message `json:"messages"`
}

type LoadRoomsMessagesRequest struct {
	RoomName string `json:"roomName"`
	Amount   int    `json:"amount"`
	Offset   int    `json:"offset"`
}

type UserInRoomRequest struct {
	RoomName string `json:"roomName"`
}

type LeaveRoomRequest struct {
	RoomName string `json:"roomName"`
}

type LoadRoomsMessagesResponse struct {
	GeneralResponse
	Messages []Message `json:"messages"`
}

type GetRoomsResponse struct {
	GeneralResponse
	Rooms []string `json:"rooms"`
}

type CancelUpdateRequest struct {
	RoomName string `json:"roomName"`
}
