package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/adamkadda/ntumiwa/api/handler"
	"github.com/adamkadda/ntumiwa/internal/auth"
	"github.com/adamkadda/ntumiwa/internal/config"
	"github.com/adamkadda/ntumiwa/internal/db"
	"github.com/adamkadda/ntumiwa/internal/hash"
	"github.com/adamkadda/ntumiwa/internal/logging"
	"github.com/adamkadda/ntumiwa/internal/middleware"
	"github.com/adamkadda/ntumiwa/internal/session"
	"github.com/joho/godotenv"
)

type Server struct {
	db     *db.DB
	port   string
	router http.Handler
}

func New() *Server {
	err := godotenv.Load(".env.api")
	if err != nil {
		log.Fatalf("godotenv failure: %v\n", err)
	}

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[CONFIG] Load failure: %v\n", err)
	}

	log.Printf("[CONFIG] Load finished !\n")

	logging.Setup(config.Logging)

	// DB setup,
	// construct connection string (DSN),
	// pass base timeout value for queries
	db := db.New(config.DB.DSN(), config.DB.Timeout)

	hash.Setup(config.Hash)

	layers := []middleware.Middleware{
		logging.Middleware(),
	}

	if config.AppEnv != "TEST" {
		store := session.NewSessionStore()
		manager := session.NewSessionManager(config.Session, store)

		layers = append(layers,
			session.Middleware(manager),
			auth.Middleware(manager, db),
		)
	}

	stack := middleware.NewStack(layers...)

	router := http.NewServeMux()

	venueHandler := handler.NewVenueHandler(db)
	venueHandler.RegisterRoutes(router)

	composerHandler := handler.NewComposerHandler(db)
	composerHandler.RegisterRoutes(router)

	pieceHandler := handler.NewPieceHandler(db)
	pieceHandler.RegisterRoutes(router)

	programmeHandler := handler.NewProgrammeHandler(db)
	programmeHandler.RegisterRoutes(router)

	eventHandler := handler.NewEventHandler(db)
	eventHandler.RegisterRoutes(router)

	// TODO: Implement biography endpoint

	// TODO: Implement media (photos/videos?) endpoint

	// TODO: Implement contact details endpoint

	// TODO: Implement login & logout endpoints

	return &Server{
		db:     db,
		port:   config.Port,
		router: stack(router),
	}
}

func (s *Server) Run() error {
	fmt.Printf("Listening on port %s...\n", s.port)
	return http.ListenAndServe(":"+s.port, s.router)
}
