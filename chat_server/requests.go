package chat_server

import "log"

// registers a user to the db if his username does not exists already
func Register(req *RegisterRequest) interface{} {
	ok, err := RegisterDB(db, req.Username, req.Password)

	if err != nil {
		log.Println("ERROR: ", err)
		return MakeErrorResponse("db error")
	}

	if ok {
		return GeneralResponse{CODE_REGISTER, STATUS_SUCCESS}
	}
	return GeneralResponse{CODE_REGISTER, STATUS_FAILED}
}

// logs the user into the system if his password and username are correct
func Login(req *LoginRequest, client *Client) interface{} {
	if CheckUsersPassword(db, req.Username, req.Password) {
		client.username = req.Username
		return GeneralResponse{CODE_LOGIN, STATUS_SUCCESS}
	}
	return GeneralResponse{CODE_LOGIN, STATUS_FAILED}
}
