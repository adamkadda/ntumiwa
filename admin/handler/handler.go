package handler

import (
	"html/template"
	"log"
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/tmpl"
)

func Dashboard(logger *log.Logger, template *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := template.ExecuteTemplate(w, "base", nil)
		if err != nil {
			logger.Printf("ERROR: Failed to render dashboard: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}
	})
}

func Login(logger *log.Logger, templates tmpl.TemplateMap) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err := templates["login"].ExecuteTemplate(w, "base", nil)
			if err != nil {
				logger.Printf("ERROR: Failed to render login page: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
				return
			}
		case http.MethodPost:
			// TODO: Pass on to backend API
			if err := r.ParseForm(); err != nil {
				logger.Printf("ERROR: Error parsing login form: %v", err)
				http.Error(w, "Bad Request (#`^`)/", http.StatusBadRequest)
			}

			// username := r.FormValue("username")
			// password := r.FormValue("password")

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed"))
		}
	})
}
