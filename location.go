package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

type LocationService interface {
	ListLocations(ctx context.Context) ([]*Location, error)
	GetLocationByID(ctx context.Context, locationID int) (*Location, error)
	GetLocationByName(ctx context.Context, locationName string) (*Location, error)
	GetLocationByNameFuzzy(ctx context.Context, search string) ([]*Location, error)
	ListCountries(ctx context.Context) ([]*Country, error)
	ListMarketCodes(ctx context.Context) ([]string, error)
	IsValidMarketCode(ctx context.Context, marketCode string) (*bool, error)
	FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations []*Location) ([]*Location, error)
	FilterLocationsByMcrAvailability(ctx context.Context, mcrAvailable bool, locations []*Location) []*Location
	GetRandom(ctx context.Context, marketCode string) (*Location, error)
}

type LocationServiceOp struct {
	Client *Client
}

func NewLocationService(c *Client) *LocationServiceOp {
	return &LocationServiceOp{
		Client: c,
	}
}

type Location struct {
	Name             string                 `json:"name"`
	Country          string                 `json:"country"`
	LiveDate         int                    `json:"liveDate"`
	SiteCode         string                 `json:"siteCode"`
	NetworkRegion    string                 `json:"networkRegion"`
	Address          map[string]string      `json:"address"`
	Campus           string                 `json:"campus"`
	Latitude         float64                `json:"latitude"`
	Longitude        float64                `json:"longitude"`
	Products         map[string]interface{} `json:"products"`
	Market           string                 `json:"market"`
	Metro            string                 `json:"metro"`
	VRouterAvailable bool                   `json:"vRouterAvailable"`
	ID               int                    `json:"id"`
	Status           string                 `json:"status"`
}

type Country struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	SiteCount int    `json:"siteCount"`
}

type LocationResponse struct {
	Message string      `json:"message"`
	Terms   string      `json:"terms"`
	Data    []*Location `json:"data"`
}

type CountryResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []*CountryInnerResponse `json:"data"`
}

type CountryInnerResponse struct {
	Countries     []*Country `json:"countries"`
	NetworkRegion string     `json:"networkRegion"`
}

func (svc *LocationServiceOp) ListLocations(ctx context.Context) ([]*Location, error) {
	path := "/v2/locations"
	url := svc.Client.BaseURL.JoinPath(path).String()
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return nil, resErr
	}
	defer response.Body.Close()
	body, fileErr := io.ReadAll(response.Body)

	if fileErr != nil {
		return nil, fileErr
	}

	locationResponse := &LocationResponse{}

	unmarshalErr := json.Unmarshal(body, locationResponse)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return locationResponse.Data, nil
}

func (svc *LocationServiceOp) GetLocationByID(ctx context.Context, locationID int) (*Location, error) {
	allLocations, locErr := svc.ListLocations(ctx)
	if locErr != nil {
		return nil, locErr
	}
	for _, location := range allLocations {
		if location.ID == locationID {
			return location, nil
		}
	}
	return nil, errors.New(ERR_LOCATION_NOT_FOUND)
}

func (svc *LocationServiceOp) GetLocationByName(ctx context.Context, locationName string) (*Location, error) {
	allLocations, locErr := svc.ListLocations(ctx)
	if locErr != nil {
		return nil, locErr
	}
	for _, location := range allLocations {
		if location.Name == locationName {
			return location, nil
		}
	}
	return nil, errors.New(ERR_LOCATION_NOT_FOUND)
}

func (svc *LocationServiceOp) GetLocationByNameFuzzy(ctx context.Context, search string) ([]*Location, error) {
	locations, err := svc.ListLocations(ctx)
	if err != nil {
		return nil, err
	}
	var matchedLocations []*Location

	for i := 0; i < len(locations); i++ {
		if fuzzy.Match(search, locations[i].Name) {
			matchedLocations = append(matchedLocations, locations[i])
		}
	}

	if len(matchedLocations) > 0 {
		return matchedLocations, nil
	} else {
		return matchedLocations, errors.New(ERR_NO_MATCHING_LOCATIONS)
	}
}

func (svc *LocationServiceOp) ListCountries(ctx context.Context) ([]*Country, error) {
	path := "/v2/networkRegions"
	url := svc.Client.BaseURL.JoinPath(path).String()
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return nil, resErr
	}
	defer response.Body.Close()

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	countryResponse := CountryResponse{}

	unmarshalErr := json.Unmarshal(body, &countryResponse)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	allCountries := make([]*Country, 0)

	for i := 0; i < len(countryResponse.Data); i++ {
		if countryResponse.Data[i].NetworkRegion == "MP1" {
			allCountries = countryResponse.Data[i].Countries
		}
	}

	return allCountries, nil
}

func (svc *LocationServiceOp) ListMarketCodes(ctx context.Context) ([]string, error) {
	countries, countriesErr := svc.ListCountries(ctx)
	if countriesErr != nil {
		return nil, countriesErr
	}
	var marketCodes []string
	for i := 0; i < len(countries); i++ {
		marketCodes = append(marketCodes, countries[i].Prefix)
	}

	return marketCodes, nil
}

func (svc *LocationServiceOp) IsValidMarketCode(ctx context.Context, marketCode string) (*bool, error) {
	marketCodes, err := svc.ListMarketCodes(ctx)
	if err != nil {
		return nil, err
	}
	found := false

	for i := 0; i < len(marketCodes); i++ {
		if marketCodes[i] == marketCode {
			found = true
		}
	}

	return PtrTo(found), nil
}

func (svc *LocationServiceOp) FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations []*Location) ([]*Location, error) {
	existingLocations := locations
	toReturn := []*Location{}
	isValid, err := svc.IsValidMarketCode(ctx, marketCode)
	if err != nil {
		return nil, err
	}
	if *isValid {
		for _, loc := range existingLocations {
			if loc.Market == marketCode {
				toReturn = append(toReturn, loc)
			}
		}
	}
	return toReturn, nil
}

func (svc *LocationServiceOp) FilterLocationsByMcrAvailability(ctx context.Context, mcrAvailable bool, locations []*Location) []*Location {
	existingLocations := locations
	toReturn := []*Location{}
	for _, location := range existingLocations {
		if _, ok := location.Products["mcr2"]; ok {
			toReturn = append(toReturn, location)
		}
	}
	return toReturn
}

func (svc *LocationServiceOp) GetRandom(ctx context.Context, marketCode string) (*Location, error) {
	testLocations, err := svc.ListLocations(ctx)
	if err != nil {
		return nil, err
	}
	filtered, err := svc.FilterLocationsByMarketCode(ctx, marketCode, testLocations)
	if err != nil {
		return nil, err
	}
	filteredByMCR := svc.FilterLocationsByMcrAvailability(ctx, true, filtered)
	testLocation := filteredByMCR[GenerateRandomNumber(0, len(filteredByMCR)-1)]
	return testLocation, nil
}
