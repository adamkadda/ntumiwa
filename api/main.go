package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/adamkadda/ntumiwa-site/internal/hash"
	"github.com/adamkadda/ntumiwa-site/internal/session"
	"github.com/adamkadda/ntumiwa-site/shared/config"
	"github.com/adamkadda/ntumiwa-site/shared/middleware"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Failed to load .env: %v", err)
	} else {
		fmt.Println("Loaded .env successfully")
	}

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	sessionManager := session.NewSessionManager(
		session.NewSessionStore(),
		config.Session.GCInterval,
		config.Session.IdleExpiration,
		config.Session.AbsoluteExpiration,
		config.Session.CookieName,
		config.Session.Domain,
		config.SecretKey,
	)

	hash.Setup(config.Hash)

	logger := log.New(os.Stdout, "["+config.ServerType+"]", log.LstdFlags)

	mux := http.NewServeMux()

	adminStack := middleware.NewStack(
		middleware.LoggingMiddleware,
		sessionManager.Middleware,
	)

	server := http.Server{
		Addr:    config.Port,
		Handler: adminStack(mux),
	}

	logger.Printf("Listening on port %s ...\n", config.Port)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
