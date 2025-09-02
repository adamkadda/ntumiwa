package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/adamkadda/ntumiwa-site/admin/handler"
	"github.com/adamkadda/ntumiwa-site/shared/config"
	"github.com/adamkadda/ntumiwa-site/shared/tmpl"
)

//go:embed templates/*.html
var tmplDir embed.FS

var templates tmpl.TemplateMap

func loadTemplates() {
	base := template.Must(template.ParseFS(tmplDir, "templates/base.html"))
	pages := []string{"login"}

	for _, page := range pages {
		tpl := template.Must(base.Clone())
		template.Must(tpl.ParseFiles(fmt.Sprintf("templates/%s.html", page)))
		templates[page] = tpl
	}
}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := log.New(os.Stdout, "["+config.ServerType+"]", log.LstdFlags)

	mux := http.NewServeMux()

	mux.Handle("/", handler.Dashboard(logger, templates["dash"]))
	mux.Handle("/login", handler.Login(logger, templates))
}
