package main

import (
	"log"

	"github.com/Go-Chat/internal/server"
)

func main() {
	s, err := server.NewServer()
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}
	log.Println("Server starting on :8080")
	if err := s.Start(":8080"); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
