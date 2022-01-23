package chat_server

const (
	STATUS_SUCCESS = 1
	STATUS_FAILED  = 0

	CODE_AUTH             = "00"
	CODE_UPDATE           = "01"
	CODE_LOGIN            = "02"
	CODE_REGISTER         = "03"
	CODE_LOGOUT           = "04"
	CODE_CREATE_CHAT_ROOM = "05"
	CODE_DELETE_CHAT_ROOM = "06"
	CODE_JOIN_CHAT_ROOM   = "07"
	CODE_ERR              = "11"
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

type ErrorResponse struct {
	Code string `json:"code"`
	Err  string `json:"error"`
}

// functions to insert the resp code automatically

func MakeErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{CODE_ERR, err}
}
