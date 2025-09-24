package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/adamkadda/ntumiwa/public/handler"
	"github.com/adamkadda/ntumiwa/public/pagedata"
	"github.com/adamkadda/ntumiwa/shared/api"
	"github.com/adamkadda/ntumiwa/shared/config"
	"github.com/adamkadda/ntumiwa/shared/middleware"
	"github.com/joho/godotenv"
)

//go:embed templates/*.html
var tmplDir embed.FS

var templates map[string]*template.Template

func loadTemplates() {
	base := template.Must(template.ParseFS(tmplDir, "templates/base.html"))
	pages := []string{"home", "biography", "performances", "media", "contact", "404", "5xx"}

	templates = make(map[string]*template.Template)
	for _, page := range pages {
		tpl := template.Must(base.Clone())
		template.Must(tpl.ParseFiles(fmt.Sprintf("templates/%s.html", page)))
		templates[page] = tpl
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	} else {
		fmt.Println(".env file loaded successfully")
	}

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	loadTemplates()

	apiClient := api.NewAPIClient(config)

	pageData, err := pagedata.New(apiClient)
	if err != nil {
		log.Fatalf("Failed to init page data: %v", err)
	}

	logger := log.New(os.Stdout, "["+config.ServerType+"]", log.LstdFlags)

	fs := http.FileServer(http.Dir("./static/"))

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/{$}", handler.Home(logger, templates["home"], pageData))
	mux.Handle("/biography", handler.Biography(logger, templates["biography"], pageData))
	mux.Handle("/performances", handler.Performances(logger, templates["performances"], pageData))
	mux.Handle("/media", handler.Media(logger, templates["media"], pageData))
	mux.Handle("/contact", handler.Contact(logger, templates["contact"], pageData))
	mux.Handle("/", handler.NotFound(logger, templates["404"]))

	stack := middleware.NewStack(
		middleware.Logging(logger),
	)

	server := http.Server{
		Addr:    config.Port,
		Handler: stack(mux),
	}

	logger.Printf("Listening on port %s ...\n", config.Port)

	err = server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
