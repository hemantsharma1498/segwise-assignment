package main

import (
	"github.com/hemantsharma1498/segwise-assignment/server"
	"log"
	"os"
)

func main() {
	log.Printf("Initialising service")

	OpenAIApiKey := os.Getenv("OPENAI_API_KEY")
	if OpenAIApiKey == "" {
		log.Panic("Couldn't find OpenAI API key")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3100"
	}
	s := server.InitServer(OpenAIApiKey)
	if err := s.Start(port); err != nil {
		log.Panicf("Failed to initialise server at %s, error: %s\n", port, err)
	}
}
