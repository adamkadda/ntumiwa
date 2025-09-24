package handler

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamkadda/ntumiwa/public/pagedata"
	"github.com/adamkadda/ntumiwa/shared/models"
	"github.com/stretchr/testify/assert"
)

var validHomeData = &pagedata.HomeData{
	BriefBio: []string{"a valid", "brief", "biography"},
	UpcomingPerformances: []models.Performance{
		{
			Title:      "Paasconcert",
			Venue:      "Kunstkerk, Dordrect",
			ExactDate:  "2025-04-20 19:30",
			Date:       "15 April, 2025",
			TicketLink: "https://example.com/paasconcert",
			Programme: []models.Piece{
				{Composer: "Brahms", Title: "Hungarian Dance No. 5"},
				{Composer: "Mozart", Title: "Sonata KV 381"},
			},
		},
		{
			Title:      "Talent Break",
			Venue:      "De Doelen, Rotterdam",
			ExactDate:  "2024-04-03 12:30",
			Date:       "09 October, 2025",
			TicketLink: "https://example.com/talent-break",
			Programme: []models.Piece{
				{Composer: "Schoenberg", Title: "Chamber Symphony No. 1, Op. 9"},
			},
		},
	},
}

var validBioData = &pagedata.BioData{
	Biography: []string{"a valid", "full", "biography"},
}

var validPerfsData = &pagedata.PerfsData{
	UpcomingPerformances: []models.Performance{
		{
			Title:      "Winter Recital",
			Venue:      "Concertgebouw, Amsterdam",
			ExactDate:  "2025-12-10 20:00",
			Date:       "10 December, 2025",
			TicketLink: "https://example.com/winter-recital",
			Programme: []models.Piece{
				{Composer: "Chopin", Title: "Ballade No. 1 in G minor, Op. 23"},
			},
		},
	},
	PastPerformances: []models.Performance{
		{
			Title:      "Summer Serenade",
			Venue:      "Theater aan het Spui",
			ExactDate:  "2024-07-15 19:00",
			Date:       "15 July, 2024",
			TicketLink: "https://example.com/summer-serenade",
			Programme: []models.Piece{
				{Composer: "Ravel", Title: "Alborada del gracioso"},
			},
		},
	},
}

var validMediaData = &pagedata.MediaData{
	Videos: []models.Video{
		{
			Title:         "Beethoven Sonata Op. 27 No. 2",
			ExtendedTitle: "Beethoven: Sonata No. 14 'Moonlight' – Full Performance",
			EmbedURL:      "https://youtube.com/embed/example1",
		},
		{
			Title:         "Ravel - Gaspard de la nuit",
			ExtendedTitle: "Ravel: Gaspard de la nuit – Ondine, Le Gibet, Scarbo",
			EmbedURL:      "https://youtube.com/embed/example2",
		},
	},
}

var validContactData = &pagedata.ContactData{
	Position: "Clasical Pianist, Educator",
	Location: "Rotterdam, Netherlands",
	TelNum:   "+31612345678",
	TelText:  "+31 (0) 612 345 678",
	Email:    "info@nadiatumiwa.com",
}

type testCase struct {
	name                  string
	handler               func(log *log.Logger, tmpl *template.Template, pages pagedata.Pages) http.Handler
	templateName          string
	templateContent       string
	pages                 pagedata.Pages
	expectedLog           string
	expectedStatus        int
	expectedBodySubstring string
}

func runHandlerTest(
	t *testing.T,
	handlerFunc func(log *log.Logger, tmpl *template.Template, pages pagedata.Pages) http.Handler,
	templateName string,
	templateContent string,
	pages pagedata.Pages,
	expectedLog string,
	expectedStatus int,
	expectedBodySubstring string,
) {
	t.Helper()

	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)

	templates := template.Must(template.New("base").Parse(`
	{{define "` + templateName + `"}}` + templateContent + `{{end}}
	{{define "500"}}<html><body><h1>Internal Server Error</h1></body></html>{{end}}
	`))

	handler := handlerFunc(logger, templates, pages)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	assert.Contains(t, buf.String(), expectedLog)
	assert.Equal(t, expectedStatus, resp.StatusCode)
	assert.Contains(t, string(body), expectedBodySubstring)
}

type mockPages struct{}

func (m *mockPages) GetHomePageData() (*pagedata.HomeData, error) {
	return validHomeData, nil
}

func (m *mockPages) GetBioPageData() (*pagedata.BioData, error) {
	return validBioData, nil
}

func (m *mockPages) GetPerfsPageData() (*pagedata.PerfsData, error) {
	return validPerfsData, nil
}

func (m *mockPages) GetMediaPageData() (*pagedata.MediaData, error) {
	return validMediaData, nil
}

func (m *mockPages) GetContactPageData() (*pagedata.ContactData, error) {
	return validContactData, nil
}

type mockPagesDataError struct{}

func (m *mockPagesDataError) GetHomePageData() (*pagedata.HomeData, error) {
	return nil, fmt.Errorf("simulated data error")
}

func (m *mockPagesDataError) GetBioPageData() (*pagedata.BioData, error) {
	return nil, fmt.Errorf("simulated data error")
}

func (m *mockPagesDataError) GetPerfsPageData() (*pagedata.PerfsData, error) {
	return nil, fmt.Errorf("simulated data error")
}

func (m *mockPagesDataError) GetMediaPageData() (*pagedata.MediaData, error) {
	return nil, fmt.Errorf("simulated data error")
}

func (m *mockPagesDataError) GetContactPageData() (*pagedata.ContactData, error) {
	return nil, fmt.Errorf("simulated data error")
}

func TestHandlers_Success(t *testing.T) {
	tests := []testCase{
		{
			name:                  "Home - Success",
			handler:               Home,
			templateName:          "home",
			templateContent:       `<html><body><h1>Home Page</h1></body></html>`,
			pages:                 &mockPages{},
			expectedStatus:        http.StatusOK,
			expectedLog:           "",
			expectedBodySubstring: "Home Page",
		},
		{
			name:                  "Biography - Success",
			handler:               Biography,
			templateName:          "bio",
			templateContent:       `<html><body><h1>Biography Page</h1></body></html>`,
			pages:                 &mockPages{},
			expectedStatus:        http.StatusOK,
			expectedLog:           "",
			expectedBodySubstring: "Biography Page",
		},
		{
			name:                  "Performances - Success",
			handler:               Performances,
			templateName:          "perfs",
			templateContent:       `<html><body><h1>Performances Page</h1></body></html>`,
			pages:                 &mockPages{},
			expectedStatus:        http.StatusOK,
			expectedLog:           "",
			expectedBodySubstring: "Performances Page",
		},
		{
			name:                  "Media - Success",
			handler:               Media,
			templateName:          "media",
			templateContent:       `<html><body><h1>Media Page</h1></body></html>`,
			pages:                 &mockPages{},
			expectedStatus:        http.StatusOK,
			expectedLog:           "",
			expectedBodySubstring: "Media Page",
		},
		{
			name:                  "Contact - Success",
			handler:               Contact,
			templateName:          "contact",
			templateContent:       `<html><body><h1>Contact Page</h1></body></html>`,
			pages:                 &mockPages{},
			expectedStatus:        http.StatusOK,
			expectedLog:           "",
			expectedBodySubstring: "Contact Page",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runHandlerTest(
				t,
				tc.handler,
				tc.templateName,
				tc.templateContent,
				tc.pages,
				tc.expectedLog,
				tc.expectedStatus,
				tc.expectedBodySubstring,
			)
		})
	}
}

func TestHandlers_DataError(t *testing.T) {
	tests := []testCase{
		{
			name:                  "Home Data Error",
			handler:               Home,
			templateName:          "home",
			templateContent:       `<html><body><h1>Home</h1></body></html>`,
			pages:                 &mockPagesDataError{},
			expectedLog:           "ERROR: Failed to get home page data",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
		{
			name:                  "Biography Data Error",
			handler:               Biography,
			templateName:          "bio",
			templateContent:       `<html><body><h1>Bio</h1></body></html>`,
			pages:                 &mockPagesDataError{},
			expectedLog:           "ERROR: Failed to get biography page data",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
		{
			name:                  "Performances Data Error",
			handler:               Performances,
			templateName:          "perfs",
			templateContent:       `<html><body><h1>Performances</h1></body></html>`,
			pages:                 &mockPagesDataError{},
			expectedLog:           "ERROR: Failed to get performances page data",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
		{
			name:                  "Media Data Error",
			handler:               Media,
			templateName:          "media",
			templateContent:       `<html><body><h1>Media</h1></body></html>`,
			pages:                 &mockPagesDataError{},
			expectedLog:           "ERROR: Failed to get media page data",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
		{
			name:                  "Contact Data Error",
			handler:               Contact,
			templateName:          "contact",
			templateContent:       `<html><body><h1>Contact</h1></body></html>`,
			pages:                 &mockPagesDataError{},
			expectedLog:           "ERROR: Failed to get contact page data",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runHandlerTest(
				t,
				tc.handler,
				tc.templateName,
				tc.templateContent,
				tc.pages,
				tc.expectedLog,
				tc.expectedStatus,
				tc.expectedBodySubstring,
			)
		})
	}
}

func TestHandlers_TemplateExecutionError(t *testing.T) {
	tests := []testCase{
		{
			name:                  "Home Template Error",
			handler:               Home,
			templateName:          "foo", // mismatched name
			templateContent:       `<h1>Home</h1>`,
			pages:                 &mockPages{},
			expectedLog:           "ERROR: Failed to render home.html",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
		{
			name:                  "Biography Template Error",
			handler:               Biography,
			templateName:          "foo",
			templateContent:       `<h1>Bio</h1>`,
			pages:                 &mockPages{},
			expectedLog:           "ERROR: Failed to render biography.html",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
		{
			name:                  "Performances Template Error",
			handler:               Performances,
			templateName:          "foo",
			templateContent:       `<h1>Performances</h1>`,
			pages:                 &mockPages{},
			expectedLog:           "ERROR: Failed to render performances.html",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
		{
			name:                  "Media Template Error",
			handler:               Media,
			templateName:          "foo",
			templateContent:       `<h1>Media</h1>`,
			pages:                 &mockPages{},
			expectedLog:           "ERROR: Failed to render media.html",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
		{
			name:                  "Contact Template Error",
			handler:               Contact,
			templateName:          "foo",
			templateContent:       `<h1>Contact</h1>`,
			pages:                 &mockPages{},
			expectedLog:           "ERROR: Failed to render contact.html",
			expectedStatus:        http.StatusInternalServerError,
			expectedBodySubstring: "Internal Server Error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runHandlerTest(
				t,
				tc.handler,
				tc.templateName,
				tc.templateContent,
				tc.pages,
				tc.expectedLog,
				tc.expectedStatus,
				tc.expectedBodySubstring,
			)
		})
	}
}
