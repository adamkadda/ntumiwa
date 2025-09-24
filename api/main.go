package main

import (
	"log"

	"github.com/adamkadda/ntumiwa/api/server"
)

func main() {
	server := server.New()

	err := server.Run()
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
