package server

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type HomeReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	LinkedinUrl string `json:"linkedinUrl"`
}

type LoginRes struct {
	UserID int    `json:"userId"`
	Token  string `json:"token"`
}

type HomeRes struct {
	Msg string `json:"msg"`
}
