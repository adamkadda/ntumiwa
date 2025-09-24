package pagedata

import "github.com/adamkadda/ntumiwa/shared/models"

type HomeData struct {
	BriefBio             []string             `json:"briefBio"`
	UpcomingPerformances []models.Performance `json:"upcomingPerformances"`
}

type BioData struct {
	Biography []string `json:"biography"`
}

type PerfsData struct {
	UpcomingPerformances []models.Performance `json:"upcomingPerformances"`
	PastPerformances     []models.Performance `json:"pastPerformances"`
}

type MediaData struct {
	Videos []models.Video `json:"videos"`
}

type ContactData = models.ContactDetails
