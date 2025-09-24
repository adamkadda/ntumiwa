package pagedata

import (
	"time"

	"github.com/adamkadda/ntumiwa/shared/api"
	"github.com/adamkadda/ntumiwa/shared/cache"
)

type Pages interface {
	GetHomePageData() (*HomeData, error)
	GetBioPageData() (*BioData, error)
	GetPerfsPageData() (*PerfsData, error)
	GetMediaPageData() (*MediaData, error)
	GetContactPageData() (*ContactData, error)
}

type PageData struct {
	apiClient *api.APIClient
	home      *cache.Cache[HomeData]
	bio       *cache.Cache[BioData]
	perfs     *cache.Cache[PerfsData]
	media     *cache.Cache[MediaData]
	contact   *cache.Cache[ContactData]
}

func New(apiClient *api.APIClient) (*PageData, error) {
	pageData := &PageData{
		apiClient: apiClient,
		home:      cache.New[HomeData](1 * time.Hour),
		bio:       cache.New[BioData](1 * time.Hour),
		perfs:     cache.New[PerfsData](1 * time.Hour),
		media:     cache.New[MediaData](1 * time.Hour),
		contact:   cache.New[ContactData](1 * time.Hour),
	}

	homeData, err := pageData.GetHomePageData()
	if err != nil {
		return nil, err
	}
	pageData.home.Set(homeData)

	bioData, err := pageData.GetBioPageData()
	if err != nil {
		return nil, err
	}
	pageData.bio.Set(bioData)

	perfsData, err := pageData.GetPerfsPageData()
	if err != nil {
		return nil, err
	}
	pageData.perfs.Set(perfsData)

	mediaData, err := pageData.GetMediaPageData()
	if err != nil {
		return nil, err
	}
	pageData.media.Set(mediaData)

	contactData, err := pageData.GetContactPageData()
	if err != nil {
		return nil, err
	}
	pageData.contact.Set(contactData)

	return pageData, nil
}

func (pageData *PageData) GetHomePageData() (*HomeData, error) {
	data := pageData.home.Get()

	if data == nil {
		briefBio, err := pageData.apiClient.FetchBriefBio()
		if err != nil {
			return nil, err
		}

		upcomingPerformances, err := pageData.apiClient.FetchUpcomingPerformances()
		if err != nil {
			return nil, err
		}

		data = &HomeData{
			BriefBio:             briefBio,
			UpcomingPerformances: upcomingPerformances,
		}
	}

	return data, nil
}

func (pageData *PageData) GetBioPageData() (*BioData, error) {
	data := pageData.bio.Get()
	if data == nil {
		biography, err := pageData.apiClient.FetchFullBio()
		if err != nil {
			return nil, err
		}
		data = &BioData{
			Biography: biography,
		}
	}
	return data, nil
}

func (pageData *PageData) GetPerfsPageData() (*PerfsData, error) {
	data := pageData.perfs.Get()
	if data == nil {
		upcoming, err := pageData.apiClient.FetchUpcomingPerformances()
		if err != nil {
			return nil, err
		}
		past, err := pageData.apiClient.FetchPastPerformances()
		if err != nil {
			return nil, err
		}
		data = &PerfsData{
			UpcomingPerformances: upcoming,
			PastPerformances:     past,
		}
	}
	return data, nil
}

func (pageData *PageData) GetMediaPageData() (*MediaData, error) {
	data := pageData.media.Get()
	if data == nil {
		videos, err := pageData.apiClient.FetchVideos()
		if err != nil {
			return nil, err
		}
		data = &MediaData{
			Videos: videos,
		}
	}
	return data, nil
}

func (pageData *PageData) GetContactPageData() (*ContactData, error) {
	data := pageData.contact.Get()
	if data == nil {
		contact, err := pageData.apiClient.FetchContactDetails()
		if err != nil {
			return nil, err
		}
		data = &ContactData{
			Position: contact.Position,
			Location: contact.Location,
			TelNum:   contact.TelNum,
			TelText:  contact.TelText,
			Email:    contact.Email,
		}
	}
	return data, nil
}

// low priority feature
func (pageData *PageData) checkHealth() error {

	/*
		Call Get() on each field/cache, check for nil.
		Fetch relevant data from API if Get() returns nil.
		Update data in field/cache with Set().

		Ideally a goroutine periodically calls checkHealth().
	*/

	return nil
}
