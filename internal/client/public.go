package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/adamkadda/ntumiwa/internal/models"
)

func (api *APIClient) FetchBriefBio() ([]string, error) {
	apiURL := fmt.Sprintf("%s/public/content/bio?type=short", api.baseURL)

	resp, err := api.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch brief bio: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var briefBio []string

	err = json.Unmarshal(body, &briefBio)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal brief bio: %v", err)
	}

	return briefBio, nil
}

func (api *APIClient) FetchFullBio() ([]string, error) {
	apiURL := fmt.Sprintf("%s/public/content/bio?type=long", api.baseURL)

	resp, err := api.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch full bio: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var fullBio []string

	err = json.Unmarshal(body, &fullBio)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal full bio: %v", err)
	}

	return fullBio, nil
}

func (api *APIClient) FetchUpcomingPerformances() ([]models.Performance, error) {
	apiURL := fmt.Sprintf("%s/public/performances?type=upcoming", api.baseURL)

	resp, err := api.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch upcoming performances: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var upcomingPerformances []models.Performance

	err = json.Unmarshal(body, &upcomingPerformances)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal upcoming performances: %v", err)
	}

	return upcomingPerformances, nil
}

func (api *APIClient) FetchPastPerformances() ([]models.Performance, error) {
	apiURL := fmt.Sprintf("%s/public/performances?type=past", api.baseURL)

	resp, err := api.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch past performances: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pastPerformances []models.Performance

	err = json.Unmarshal(body, &pastPerformances)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal past performances: %v", err)
	}

	return pastPerformances, nil
}

func (api *APIClient) FetchVideos() ([]models.Video, error) {
	apiURL := fmt.Sprintf("%s/public/media/videos", api.baseURL)

	resp, err := api.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch videos: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var videos []models.Video

	err = json.Unmarshal(body, &videos)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal videos: %v", err)
	}

	return videos, nil
}

func (api *APIClient) FetchContactDetails() (*models.ContactDetails, error) {
	apiURL := fmt.Sprintf("%s/public/contact-details", api.baseURL)

	resp, err := api.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch contact details: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var contactDetails models.ContactDetails
	err = json.Unmarshal(body, &contactDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal contact details: %v", err)
	}

	return &contactDetails, nil
}
