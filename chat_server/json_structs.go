package chat_server

const (
	STATUS_SUCCESS = 1
	STATUS_FAILED  = 0

	STATE_NORMAL	= 0
	STATE_BAN	 	= 1

	CODE_AUTH             	   = "00"
	CODE_UPDATE           	   = "01"
	CODE_LOGIN            	   = "02"
	CODE_REGISTER         	   = "03"
	CODE_LOGOUT           	   = "04"
	CODE_CREATE_CHAT_ROOM 	   = "05"
	CODE_DELETE_CHAT_ROOM 	   = "06"
	CODE_JOIN_CHAT_ROOM   	   = "07"
	CODE_KICK_FROM_CHAT_ROOM   = "08"
	CODE_BAN_FROM_CHAT_ROOM	   = "09"
	CODE_UNBAN_FROM_CHAT_ROOM  = "10"
	CODE_ERR              	   = "99"
)

type GeneralResponse struct {
	Code   string `json:"code"`
	Status int    `json:"status"`
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
	Name     string `json:"name"`
	Password string `json:"password"`
}

type JoinChatRoomRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type DeleteChatRoomRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type KickFromChatRoomRequest struct {
	Username string `json:"username"`
	Name	 string `json:"name"`
}

type BanFromChatRoomRequest struct {
	Username string `json:"username"`
	Name	 string `json:"name"`
}

type UnBanFromChatRoomRequest struct {
	Username string `json:"username"`
	Name	 string `json:"name"`
}

type ErrorResponse struct {
	Code string `json:"code"`
	Err  string `json:"error"`
}

// functions to insert the resp code automatically

func MakeErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{CODE_ERR, err}
}
