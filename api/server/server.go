package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adamkadda/ntumiwa-site/api/handler"
	"github.com/adamkadda/ntumiwa-site/internal/auth"
	"github.com/adamkadda/ntumiwa-site/internal/db"
	"github.com/adamkadda/ntumiwa-site/internal/hash"
	"github.com/adamkadda/ntumiwa-site/internal/session"
	"github.com/adamkadda/ntumiwa-site/shared/config"
	"github.com/adamkadda/ntumiwa-site/shared/logging"
	"github.com/adamkadda/ntumiwa-site/shared/middleware"
	"github.com/joho/godotenv"
)

type Server struct {
	db     *db.DB
	addr   string
	router *http.ServeMux
}

func New() *Server {
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

	fmt.Print("Config loaded !\n")

	// DB setup,
	// construct connection string (DSN),
	// pass base timeout value for queries
	db := db.New(config.DB.DSN(), config.DB.Timeout)
	fmt.Print("DB initialized !\n")

	hash.Setup(config.Hash)
	fmt.Print("Hash setup complete !\n")

	logging.Setup(config.Logging)
	fmt.Print("Log setup complete !\n")

	manager := session.NewSessionManager(
		session.NewSessionStore(),
		config.Session.GCInterval,
		config.Session.AbsoluteExpiration,
		config.Session.IdleExpiration,
		config.Session.Domain,
		config.Session.CookieName,
		config.SecretKey,
	)

	fmt.Print("Session manager initialized !\n")

	// TODO: Apply to handlers after testing
	_ = middleware.NewStack(
		logging.Middleware(manager),
		session.Middleware(manager),
		auth.Middleware(manager, db),
	)

	// fmt.Print("Middleware ready !")

	eventHandler := handler.NewEventHandler(db)
	programmeHandler := handler.NewProgrammeHandler(db)
	pieceHandler := handler.NewPieceHandler(db)
	composerHandler := handler.NewComposerHandler(db)
	venueHandler := handler.NewVenueHandler(db)

	router := http.NewServeMux()

	router.Handle("events/{id}", eventHandler)
	router.HandleFunc("/events/{id}/draft", eventHandler.DraftEvent)
	router.HandleFunc("/events/{id}/publish", eventHandler.PublishEvent)
	router.HandleFunc("/events/{id}/archive", eventHandler.ArchiveEvent)

	router.Handle("/programmes/{id}", programmeHandler)

	router.Handle("/pieces/{id}", pieceHandler)

	router.Handle("/composers/{id}", composerHandler)

	router.Handle("/venues/{id}", venueHandler)

	// TODO: Implement biography endpoint

	// TODO: Implement media (photos/videos?) endpoint

	// TODO: Implement contact details endpoint

	// TODO: Implement login & logout endpoints

	fmt.Print("Routes registered !\n")

	return &Server{
		db:     db,
		addr:   config.Port,
		router: router,
	}
}

func (s *Server) Run() error {
	fmt.Printf("Listening on port %s...\n", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}
