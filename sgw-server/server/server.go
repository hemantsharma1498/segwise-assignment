package server

import (
	"log"
	"net/http"
)

type Server struct {
	Router *http.ServeMux
}

func InitServer() *Server {
	s := &Server{Router: http.NewServeMux()}
	s.Routes()
	return s
}

func (m *Server) Start(port string) error {
	log.Printf("Starting auction server at address: %s\n", port)
	if err := http.ListenAndServe(":"+port, m.Router); err != nil {
		return err
	}
	return nil
}
