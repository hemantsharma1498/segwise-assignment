package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hemantsharma1498/segwise-assignment/pkg/scraper"
	"github.com/hemantsharma1498/segwise-assignment/pkg/utils"
)

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	d := &LoginReq{}
	if err := utils.DecodeReqBody(r, d); err != nil {
		utils.WriteResponse(w, "Encountered an error. Please try again", http.StatusInternalServerError)
		return
	}
	if !utils.ValidEmail(d.Email) {
		utils.WriteResponse(w, "invalid email", http.StatusBadRequest)
		return
	}

	//Code here
	/*
		token, err := auth.GenerateJWT(users[0].UserID, d.Email)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
	*/

}

func (s *Server) Home(w http.ResponseWriter, r *http.Request) {
	d := &HomeReq{}
	if err := utils.DecodeReqBody(r, d); err != nil {
		utils.WriteResponse(w, "Encountered an error. Please try again", http.StatusInternalServerError)
		return
	}
	if !utils.ValidEmail(d.Email) {
		utils.WriteResponse(w, "invalid email", http.StatusBadRequest)
		return
	}

	scraper, err := scraper.NewScraper(d.Email, d.Password, d.LinkedinUrl)
	if err != nil {
		log.Printf("error while getting posts: %v\n", err)
	}
	if err := scraper.GetRecentPosts(); err != nil {
		log.Printf("error while getting posts: %v\n", err)
	}

	fmt.Println(len(scraper.Profile.Posts))

	//If posts are less than 2, get user information
	if len(scraper.Profile.Posts) <= 2 {
		if err := scraper.GetNameAndLocation(); err != nil {
			log.Printf("error while getting name && location: %v\n", err)
		}
		if err := scraper.GetExperiences(); err != nil {
			log.Printf("error while getting experiences: %v\n", err)
		}
		if err := scraper.GetEducation(); err != nil {
			log.Printf("error while getting education: %v\n", err)
			utils.WriteResponse(w, "server encountered an error, please try again later", 500)
		}
	}
	go scraper.Close()
	fmt.Printf("%v+\n", scraper.Profile)

	//Generate message from llm
	/*
	 * msg := GetMessage(scraper.Profile)
	 */
	res := &HomeRes{Msg: "Hello"}
	utils.WriteResponse(w, res, 200)
}

type OpenAIReq struct {
	Model    string `json:"model"`
	Messages string `json:"messages"`
}

type OpenAIRole struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func getMessage() {

}
