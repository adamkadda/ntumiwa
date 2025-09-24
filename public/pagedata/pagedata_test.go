package pagedata

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/adamkadda/ntumiwa/shared/api"
	"github.com/adamkadda/ntumiwa/shared/cache"
	"github.com/adamkadda/ntumiwa/shared/config"
	"github.com/adamkadda/ntumiwa/shared/models"
	"github.com/stretchr/testify/assert"
)

var briefBio = `["a valid", "brief", "biography"]`

var biography = `["a valid", "full", "biography"]`

var upcomingPerformances = `
	[
	  {
		"title": "Winter Concert",
		"venue": "TivoliVredenburg, Utrecht",
		"exactDate": "2025-12-01 18:30",
		"date": "12 December, 2025",
		"ticketLink": "https://example.com/winter-concert",
		"programme": [
		  {
			"composer": "Chopin",
			"title": "Ballade No. 1 in G minor, Op. 23"
		  },
		  {
			"composer": "Ravel",
			"title": "Alborada de gracioso"
		  }
		]
	  }
	]
	`

var pastPerformances = `
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

var videos = `
	[
	  {
		"title": "Summer Concert 2024",
		"extendedTitle": "Summer Concert 2024 - Full Performance",
		"embedURL": "https://www.youtube.com/embed/example1"
	  },
	  {
		"title": "Masterclass Session",
		"extendedTitle": "Masterclass: Advanced Techniques",
		"embedURL": "https://www.youtube.com/embed/example2"
	  }
	]
	`

var contactDetails = `
	{
	  "position": "Classical Pianist, Educator",
	  "location": "Rotterdam, Netherlands",
	  "telNum": "+31612345678",
	  "telText": "+31 (6) 612 345 678",
	  "email": "info@nadiatumiwa.com"
	}
	`

var server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	q := r.URL.Query().Get("type")

	switch {
	case path == "/public/content/bio" && q == "short":
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(briefBio))
	case path == "/public/content/bio" && q == "long":
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(briefBio))
	case path == "/public/performances" && q == "upcoming":
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(upcomingPerformances))
	case path == "/public/performances" && q == "past":
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(upcomingPerformances))
	case path == "/public/media/videos":
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(videos))
	case path == "/public/contact-details":
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(contactDetails))
	default:
		http.NotFound(w, r)
	}
}))

func TestGetHomePageData_Success(t *testing.T) {

	config := &config.Config{
		API: config.APIClientConfig{
			BaseURL: server.URL,
		},
	}

	client := api.NewAPIClient(config)

	pagedata := &PageData{
		apiClient: client,
		home:      cache.New[HomeData](1 * time.Hour),
	}

	homeData, err := pagedata.GetHomePageData()

	expected := &HomeData{
		BriefBio: []string{"a valid", "brief", "biography"},
		UpcomingPerformances: []models.Performance{
			{
				Title:      "Winter Concert",
				Venue:      "TivoliVredenburg, Utrecht",
				ExactDate:  "2025-12-01 18:30",
				Date:       "12 December, 2025",
				TicketLink: "https://example.com/winter-concert",
				Programme: []models.Piece{
					{Composer: "Chopin", Title: "Ballade No. 1 in G minor, Op. 23"},
					{Composer: "Ravel", Title: "Alborada de gracioso"},
				},
			},
		},
	}

	assert.Nil(t, err)
	assert.Equal(t, homeData, expected)
}
