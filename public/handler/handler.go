package handler

import (
	"html/template"
	"log"
	"net/http"

	"github.com/adamkadda/ntumiwa/public/pagedata"
)

func Home(logger *log.Logger, templates *template.Template, pageData pagedata.Pages) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := pageData.GetHomePageData()
		if err != nil {
			logger.Printf("ERROR: Failed to get home page data: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err = templates.ExecuteTemplate(w, "base", data)
		if err != nil {
			logger.Printf("ERROR: Failed to render home.html: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
		}
	})
}

func Biography(logger *log.Logger, templates *template.Template, pageData pagedata.Pages) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := pageData.GetBioPageData()
		if err != nil {
			logger.Printf("ERROR: Failed to get biography page data: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err = templates.ExecuteTemplate(w, "base", data)
		if err != nil {
			logger.Printf("ERROR: Failed to render biography.html: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
		}
	})
}

// TODO: define Performances handler
func Performances(logger *log.Logger, templates *template.Template, pageData pagedata.Pages) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := pageData.GetPerfsPageData()
		if err != nil {
			logger.Printf("ERROR: Failed to get performances page data: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err = templates.ExecuteTemplate(w, "base", data)
		if err != nil {
			logger.Printf("ERROR: Failed to render performances.html: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
		}
	})
}

// TODO: define Media handler
func Media(logger *log.Logger, templates *template.Template, pageData pagedata.Pages) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := pageData.GetMediaPageData()
		if err != nil {
			logger.Printf("ERROR: Failed to get media page data: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err = templates.ExecuteTemplate(w, "base", data)
		if err != nil {
			logger.Printf("ERROR: Failed to render media.html: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
		}
	})
}

// TODO: define Contact handler
func Contact(logger *log.Logger, templates *template.Template, pageData pagedata.Pages) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := pageData.GetContactPageData()
		if err != nil {
			logger.Printf("ERROR: Failed to get contact page data: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		err = templates.ExecuteTemplate(w, "base", data)
		if err != nil {
			logger.Printf("ERROR: Failed to render contact.html: %v", err)
			InternalServerError(logger, templates).ServeHTTP(w, r)
		}
	})
}

func NotFound(logger *log.Logger, templates *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("ERROR: Path `%s` not found", r.URL.EscapedPath())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		w.WriteHeader(http.StatusNotFound)

		if err := templates.ExecuteTemplate(w, "base", nil); err != nil {
			logger.Printf("ERROR: Failed to execute notFound template: %v", err)
			w.Write([]byte("404 - Not Found"))
		}
	})
}

func InternalServerError(logger *log.Logger, templates *template.Template) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		w.WriteHeader(http.StatusInternalServerError)

		if err := templates.ExecuteTemplate(w, "base", nil); err != nil {
			logger.Printf("ERROR: Failed to execute 500 template: %v", err)
			w.Write([]byte("500 - Internal Server Error"))
		}
	})
}
