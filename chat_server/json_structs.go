package chat_server

const (
	STATUS_SUCCESS = 1
	STATUS_FAILED  = 0

	CODE_AUTH     = "00"
	CODE_UPDATE   = "01"
	CODE_LOGIN    = "02"
	CODE_REGISTER = "03"
	CODE_LOGOUT   = "04"
	CODE_MSG      = "05"
	CODE_ERR      = "11"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Code   string `json:"code"`
	Status int    `json:"status"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Code   string `json:"code"`
	Status int    `json:"status"`
}

type ErrorResponse struct {
	Code string `json:"code"`
	Err  string `json:"error"`
}

// functions to insert the resp code automatically

func MakeErrorResponse(err string) *ErrorResponse {
	return &ErrorResponse{CODE_ERR, err}
}

func MakeRegisterResponse(status int) *RegisterResponse {
	return &RegisterResponse{CODE_REGISTER, status}
}

func MakeLoginResponse(status int) *LoginResponse {
	return &LoginResponse{CODE_LOGIN, status}
}
