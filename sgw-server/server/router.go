package server

import (
	"net/http"

	"github.com/hemantsharma1498/segwise-assignment/pkg/utils"
)

func (s *Server) Routes() {
	s.Router.HandleFunc("/api/home", utils.WithCORS(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
		s.Home(w, r)
	})))
}
