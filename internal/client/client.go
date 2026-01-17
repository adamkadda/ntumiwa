package client

import (
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/config"
)

type APIClient struct {
	client  *http.Client
	baseURL string
}

/*
	This constructor function should not fail.
	Errors should be caught immediately after loading the
	config into memory.

	Therefore we do not need to handle for missing field errors.
*/

func NewAPIClient(config *config.Config) *APIClient {
	transport := &http.Transport{
		MaxIdleConns:        config.API.MaxIdleConns,
		MaxIdleConnsPerHost: config.API.MaxIdleConnsPerHost,
		IdleConnTimeout:     config.API.IdleConnTimeout,
	}

	client := &http.Client{
		Timeout:   config.API.Timeout,
		Transport: transport,
	}

	// TODO: Add auth token field for admin, empty string for public

	return &APIClient{
		client:  client,
		baseURL: config.API.BaseURL,
	}
}
