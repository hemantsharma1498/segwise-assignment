package server

type HomeReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	LinkedinUrl string `json:"linkedinUrl"`
}

type HomeRes struct {
	Msg         string   `json:"msg"`
	ParamsUsed  []string `json:"paramsUsed"`
	RecentPosts string   `json:"recentPosts"`
}
