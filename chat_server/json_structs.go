package chat_server

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Status int `json:"status"`
}

type ErrorResponse struct {
	Err string `json:"error"`
}
