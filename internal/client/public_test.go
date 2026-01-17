package client

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adamkadda/ntumiwa/internal/config"
	"github.com/adamkadda/ntumiwa/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestFetchBriefBio_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `["a valid", "brief", "biography"]`)
	}))
	defer server.Close()

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: server.URL,
		},
	}

	api := NewAPIClient(config)

	briefBio, err := api.FetchBriefBio()

	assert.Nil(t, err)
	assert.Equal(t, []string{"a valid", "brief", "biography"}, briefBio)
}

func TestFetchBriefBio_CannotReachServer(t *testing.T) {
	badURL := "http://127.0.0.1:55555"

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: badURL,
		},
	}

	api := NewAPIClient(config)

	briefBio, err := api.FetchBriefBio()

	assert.Nil(t, briefBio)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "refused")
}

func TestFetchBriefBio_UnexpectedStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
	}))

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: server.URL,
		},
	}

	api := NewAPIClient(config)

	briefBio, err := api.FetchBriefBio()

	assert.Nil(t, briefBio)
	assert.Contains(t, err.Error(), "unexpected status code")
}

func TestFetchBriefBio_InvalidResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `"an invalid response body"`)
	}))
	defer server.Close()

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: server.URL,
		},
	}

	api := NewAPIClient(config)

	briefBio, err := api.FetchBriefBio()

	assert.Nil(t, briefBio)
	assert.Contains(t, err.Error(), "failed to unmarshal")
}

func TestFetchUpcomingEvents_Success(t *testing.T) {
	body := `
	[
	  {
		"title": "Paasconcert",
		"venue": "Kunstkerk, Dordrect",
		"exactDate": "2025-04-20 19:30",
		"date": "15 April, 2025",
		"ticketLink": "https://example.com/paasconcert",
		"programme": [
		  {
			"composer": "Brahms",
			"title": "Hungarian Dance No. 5"
		  },
		  {
			"composer": "Mozart",
			"title": "Sonata KV 381"
		  }
		]
	  },
	  {
		"title": "Talent Break",
		"venue": "De Doelen, Rotterdam",
		"exactDate": "2024-04-03 12:30",
		"date": "09 October, 2025",
		"ticketLink": "https://example.com/talent-break",
		"programme": [
		  {
			"composer": "Schoenberg",
			"title": "Chamber Symphony No. 1, Op. 9"
		  }
		]
	  }
	]
	`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer server.Close()

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: server.URL,
		},
	}

	api := NewAPIClient(config)

	upcomingEvents, err := api.FetchUpcomingPerformances()

	expected := []models.Performance{
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
	}

	assert.Nil(t, err)
	assert.Equal(t, upcomingEvents, expected)

}

func TestFetchUpcomingEvents_CannotReachServer(t *testing.T) {
	badURL := "http://127.0.0.1:55555"

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: badURL,
		},
	}

	api := NewAPIClient(config)

	upcomingEvents, err := api.FetchUpcomingPerformances()

	assert.Nil(t, upcomingEvents)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "refused")
}

func TestFetchUpcomingEvents_UnexpectedStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error": "internal server error"}`, http.StatusInternalServerError)
	}))

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: server.URL,
		},
	}

	api := NewAPIClient(config)

	upcomingEvents, err := api.FetchUpcomingPerformances()

	assert.Nil(t, upcomingEvents)
	assert.Contains(t, err.Error(), "unexpected status code")

}

func TestFetchUpcomingEvents_InvalidResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `"an invalid response body"`)
	}))
	defer server.Close()

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: server.URL,
		},
	}

	api := NewAPIClient(config)

	upcomingEvents, err := api.FetchUpcomingPerformances()

	assert.Nil(t, upcomingEvents)
	assert.Contains(t, err.Error(), "failed to unmarshal")

}
