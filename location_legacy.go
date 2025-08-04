package megaport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// Location represents a location in the Megaport API v2 (DEPRECATED).
// This struct is maintained for backward compatibility but should not be used for new code.
// Use LocationV3 for new implementations.
type Location struct {
	Name             string            `json:"name"`
	Country          string            `json:"country"`
	LiveDate         *Time             `json:"liveDate"`
	SiteCode         string            `json:"siteCode"`
	NetworkRegion    string            `json:"networkRegion"`
	Address          map[string]string `json:"address"`
	Campus           string            `json:"campus"`
	Latitude         float64           `json:"latitude"`
	Longitude        float64           `json:"longitude"`
	Products         *LocationProducts `json:"products"`
	Market           string            `json:"market"`
	Metro            string            `json:"metro"`
	VRouterAvailable bool              `json:"vRouterAvailable"`
	ID               int               `json:"id"`
	Status           string            `json:"status"`
}

// LocationProducts represent the products available at a location in the Megaport API.
type LocationProducts struct {
	MCR        bool          `json:"mcr"`
	MCRVersion int           `json:"mcrVersion"`
	Megaport   []int         `json:"megaport"`
	MVE        []LocationMVE `json:"mve"`
	MCR1       []int         `json:"mcr1"`
	MCR2       []int         `json:"mcr2"`
}

// LocationMVE represents the MVE product available at a location in the Megaport API.
type LocationMVE struct {
	Sizes             []string             `json:"sizes"`
	Details           []LocationMVEDetails `json:"details"`
	MaxCPUCount       int                  `json:"maxCpuCount"`
	Version           string               `json:"version"`
	Product           string               `json:"product"`
	Vendor            string               `json:"vendor"`
	VendorDescription string               `json:"vendorDescription"`
	ID                int                  `json:"id"`
	ReleaseImage      bool                 `json:"releaseImage"`
}

// LocationMVEDetails represents the details of the MVE product available at a location in the Megaport API.
type LocationMVEDetails struct {
	Size          string `json:"size"`
	Label         string `json:"label"`
	CPUCoreCount  int    `json:"cpuCoreCount"`
	RamGB         int    `json:"ramGB"`
	BandwidthMbps int    `json:"bandwidthMbps"`
}

// Country represents a country in the Megaport Locations API.
type Country struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	SiteCount int    `json:"siteCount"`
}

// ToLegacyLocation converts a LocationV3 to the legacy Location struct for backward compatibility.
// Note: Some fields cannot be accurately converted due to structural differences between v2 and v3.
func (l *LocationV3) ToLegacyLocation() *Location {
	// Convert address from structured format to map[string]string
	addressMap := map[string]string{
		"street":   l.Address.Street,
		"suburb":   l.Address.Suburb,
		"city":     l.Address.City,
		"state":    l.Address.State,
		"postcode": l.Address.Postcode,
		"country":  l.Address.Country,
	}

	// Try to create a compatible Products struct
	var products *LocationProducts
	if l.DiversityZones != nil {
		products = &LocationProducts{
			MCR:      l.HasMCRSupport(),
			Megaport: l.GetMegaportSpeeds(),
			MCR2:     l.GetMCRSpeeds(),
			// MCRVersion, MVE, MCR1 fields cannot be accurately mapped from v3
		}
	}

	return &Location{
		ID:        l.ID,
		Name:      l.Name,
		Country:   l.Address.Country,
		Address:   addressMap,
		Latitude:  l.Latitude,
		Longitude: l.Longitude,
		Products:  products,
		Market:    l.Market,
		Metro:     l.Metro,
		Status:    l.Status,
		// Fields that cannot be converted:
		// - LiveDate: Not available in v3
		// - SiteCode: Not available in v3
		// - NetworkRegion: Not available in v3
		// - Campus: Not available in v3
		// - VRouterAvailable: Not available in v3
	}
}

// ==========================================
// DEPRECATED V2 API IMPLEMENTATION METHODS
// ==========================================

// ListLocations returns a list of all locations in the Megaport Locations API.
// Deprecated: Use ListLocationsV3 instead. The v2 API will be removed in a future version.
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
// Deprecated: Use GetLocationByIDV3 instead. The v2 API will be removed in a future version.
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
// Deprecated: Use GetLocationByNameV3 instead. The v2 API will be removed in a future version.
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
// Deprecated: Use GetLocationByNameFuzzyV3 instead. The v2 API will be removed in a future version.
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

// FilterLocationsByMarketCode filters locations by market code in the Megaport Locations API.
// Deprecated: Use FilterLocationsByMarketCodeV3 instead. The v2 API will be removed in a future version.
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
// Deprecated: Use FilterLocationsByMcrAvailabilityV3 instead. The v2 API will be removed in a future version.
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
