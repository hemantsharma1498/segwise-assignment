package main

import (
	"github.com/hemantsharma1498/segwise-assignment/server"
	"log"
	"os"
)

func main() {
	log.Printf("Initialising service")

	port := os.Getenv("PORT")
	if port == "" {
		port = "3100"
	}
	s := server.InitServer()
	if err := s.Start(port); err != nil {
		log.Panicf("Failed to initialise server at %s, error: %s\n", port, err)
	}
}
