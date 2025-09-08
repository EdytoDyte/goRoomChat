package main

import (
	"log"

	"github.com/Go-Chat/internal/client"
)

func main() {
	c, err := client.NewClient()
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// These will be prompted in the UI in a later step
	roomName := "test"
	username := "test"

	c.JoinRoom(roomName)
	c.SendUsername(username)

	if err := c.Start("localhost:8080"); err != nil {
		log.Fatalf("client failed to start: %v", err)
	}
}
