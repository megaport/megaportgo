package megaport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// LocationService is an interface for interfacing with the Location endpoints of the Megaport API.
type LocationService interface {
	ListLocations(ctx context.Context) ([]*Location, error)
	GetLocationByID(ctx context.Context, locationID int) (*Location, error)
	GetLocationByName(ctx context.Context, locationName string) (*Location, error)
	GetLocationByNameFuzzy(ctx context.Context, search string) ([]*Location, error)
	ListCountries(ctx context.Context) ([]*Country, error)
	ListMarketCodes(ctx context.Context) ([]string, error)
	IsValidMarketCode(ctx context.Context, marketCode string) (bool, error)
	FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations []*Location) ([]*Location, error)
	FilterLocationsByMcrAvailability(ctx context.Context, mcrAvailable bool, locations []*Location) []*Location
	GetRandom(ctx context.Context, marketCode string) (*Location, error)
}

// LocationServiceOp handles communication with Location methods of the Megaport API.
type LocationServiceOp struct {
	Client *Client
}

func NewLocationService(c *Client) *LocationServiceOp {
	return &LocationServiceOp{
		Client: c,
	}
}

// Location represents a location in the Megaport API.
type Location struct {
	Name             string                 `json:"name"`
	Country          string                 `json:"country"`
	LiveDate         *Time                  `json:"liveDate"`
	SiteCode         string                 `json:"siteCode"`
	NetworkRegion    string                 `json:"networkRegion"`
	Address          map[string]string      `json:"address"`
	Campus           string                 `json:"campus"`
	Latitude         float64                `json:"latitude"`
	Longitude        float64                `json:"longitude"`
	Products         *LocationProducts 		`json:"products"`
	Market           string                 `json:"market"`
	Metro            string                 `json:"metro"`
	VRouterAvailable bool                   `json:"vRouterAvailable"`
	ID               int                    `json:"id"`
	Status           string                 `json:"status"`
}

// LocationProducts represent the products available at a location in the Megaport API.
type LocationProducts struct {
	MCR 		bool 			`json:"mcr"`
	MCRVersion 	int 			`json:"mcrVersion"`
	Megaport 	[]int  			`json:"megaport"`
	MVE 		[]LocationMVE 	`json:"mve"`
	MCR1 		[]int 			`json:"mcr1"`
	MCR2		[]int			`json:"mcr2"`
}

// LocationMVE represents the MVE product available at a location in the Megaport API.
type LocationMVE struct {
	Sizes 				[]string 			 	`json:"sizes"`
	Details 			[]LocationMVEDetails 	`json:"details"`
	MaxCPUCount 		int 			 		`json:"maxCpuCount"`
	Version 			string					`json:"version"`
	Product 			string					`json:"product"`
	Vendor 				string					`json:"vendor"`
	VendorDescription 	string					`json:"vendorDescription"`
	ID					int						`json:"id"`	
	ReleaseImage 		bool					`json:"releaseImage"`
}

// LocationMVEDetails represents the details of the MVE product available at a location in the Megaport API.
type LocationMVEDetails struct {
	Size 			string 		`json:"size"`
	Label 			string 		`json:"label"`
	CPUCoreCount 	int 		`json:"cpuCoreCount"`
	RamGB 			int 		`json:"ramGB"`
	BandwidthMbps	int 		`json:"bandwidthMbps"`
}

// Country represents a country in the Megaport Locations API.
type Country struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	SiteCount int    `json:"siteCount"`
}

// LocationsResponse represents the response from the Megaport Locations API.
type LocationResponse struct {
	Message string      `json:"message"`
	Terms   string      `json:"terms"`
	Data    []*Location `json:"data"`
}

// CountryResponse represents the response from the Megaport Network Regions API.
type CountryResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []*CountryInnerResponse `json:"data"`
}

// CountriesInnerResponse represents the inner response from the Megaport Network Regions API.
type CountryInnerResponse struct {
	Countries     []*Country `json:"countries"`
	NetworkRegion string     `json:"networkRegion"`
}

// ListLocations returns a list of all locations in the Megaport Locations API.
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

// GetLocationByID returns a location by its ID in the Megaport Locations API.
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
	return nil, ErrLocationNotFound
}

// GetLocationByName returns a location by its name in the Megaport Locations API.
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
	return nil, ErrLocationNotFound
}

// GetLocationByNameFuzzy returns a location by its name in the Megaport Locations API using fuzzy search.
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
		return matchedLocations, ErrNoMatchingLocations
	}
}

// ListCountries returns a list of all countries in the Megaport Network Regions API.
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

// ListMarketCodes returns a list of all market codes in the Megaport Network Regions API.
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

// IsValidMarketCode checks if a market code is valid in the Megaport Network Regions API.
func (svc *LocationServiceOp) IsValidMarketCode(ctx context.Context, marketCode string) (bool, error) {
	found := false

	marketCodes, err := svc.ListMarketCodes(ctx)
	if err != nil {
		return found, err
	}

	for i := 0; i < len(marketCodes); i++ {
		if marketCodes[i] == marketCode {
			found = true
		}
	}

	return found, nil
}

// FilterLocationsByMarketCode filters locations by market code in the Megaport Locations API.
func (svc *LocationServiceOp) FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations []*Location) ([]*Location, error) {
	existingLocations := locations
	toReturn := []*Location{}
	isValid, err := svc.IsValidMarketCode(ctx, marketCode)
	if err != nil {
		return nil, err
	}
	if isValid {
		for _, loc := range existingLocations {
			if loc.Market == marketCode {
				toReturn = append(toReturn, loc)
			}
		}
	}
	return toReturn, nil
}

// FilterLocationsByMcrAvailability filters locations by MCR availability in the Megaport Locations API.
func (svc *LocationServiceOp) FilterLocationsByMcrAvailability(ctx context.Context, mcrAvailable bool, locations []*Location) []*Location {
	existingLocations := locations
	toReturn := []*Location{}
	for _, location := range existingLocations {
		if location.Products.MCR == mcrAvailable {
			toReturn = append(toReturn, location)
		}
	}
	return toReturn
}

// GetRandom returns a random location in the Megaport Locations API with MCR Availability. Used for integration testing.
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
