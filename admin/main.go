package main

import (
	"embed"
	"fmt"
	"html/template"

	"github.com/adamkadda/ntumiwa/internal/tmpl"
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

func main() {}
