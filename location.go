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
	// V3 API methods (RECOMMENDED for new code)
	// ListLocationsV3 returns a list of all locations in the Megaport Locations API v3.
	ListLocationsV3(ctx context.Context) ([]*LocationV3, error)
	// GetLocationByIDV3 returns a location by its ID in the Megaport Locations API v3.
	GetLocationByIDV3(ctx context.Context, locationID int) (*LocationV3, error)
	// GetLocationByNameV3 returns a location by its name in the Megaport Locations API v3.
	GetLocationByNameV3(ctx context.Context, locationName string) (*LocationV3, error)
	// GetLocationByNameFuzzyV3 returns a location by its name in the Megaport Locations API v3 using fuzzy search.
	GetLocationByNameFuzzyV3(ctx context.Context, search string) ([]*LocationV3, error)
	// FilterLocationsByMarketCodeV3 filters locations by market code in the Megaport Locations API v3.
	FilterLocationsByMarketCodeV3(ctx context.Context, marketCode string, locations []*LocationV3) ([]*LocationV3, error)
	// FilterLocationsByMcrAvailabilityV3 filters locations by MCR availability in the Megaport Locations API v3.
	FilterLocationsByMcrAvailabilityV3(ctx context.Context, mcrAvailable bool, locations []*LocationV3) []*LocationV3
	// FilterLocationsByMetroV3 filters locations by metro name in the Megaport Locations API v3.
	FilterLocationsByMetroV3(ctx context.Context, metro string, locations []*LocationV3) []*LocationV3

	// Shared methods (work with both v2 and v3)
	// ListCountries returns a list of all countries in the Megaport Network Regions API.
	ListCountries(ctx context.Context) ([]*Country, error)
	// ListMarketCodes returns a list of all market codes in the Megaport Network Regions API.
	ListMarketCodes(ctx context.Context) ([]string, error)
	// IsValidMarketCode checks if a market code is valid in the Megaport Network Regions API.
	IsValidMarketCode(ctx context.Context, marketCode string) (bool, error)

	// ListLocations returns a list of all locations in the Megaport Locations API v2.
	//
	// Deprecated: Use ListLocationsV3 instead. The v2 API will be removed in a future version.
	ListLocations(ctx context.Context) ([]*Location, error)
	// GetLocationByID returns a location by its ID in the Megaport Locations API v2.
	//
	// Deprecated: Use GetLocationByIDV3 instead. The v2 API will be removed in a future version.
	GetLocationByID(ctx context.Context, locationID int) (*Location, error)
	// GetLocationByName returns a location by its name in the Megaport Locations API v2.
	//
	// Deprecated: Use GetLocationByNameV3 instead. The v2 API will be removed in a future version.
	GetLocationByName(ctx context.Context, locationName string) (*Location, error)
	// GetLocationByNameFuzzy returns a location by its name in the Megaport Locations API v2 using fuzzy search.
	//
	// Deprecated: Use GetLocationByNameFuzzyV3 instead. The v2 API will be removed in a future version.
	GetLocationByNameFuzzy(ctx context.Context, search string) ([]*Location, error)
	// FilterLocationsByMarketCode filters locations by market code in the Megaport Locations API v2.
	//
	// Deprecated: Use FilterLocationsByMarketCodeV3 instead. The v2 API will be removed in a future version.
	FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations []*Location) ([]*Location, error)
	// FilterLocationsByMcrAvailability filters locations by MCR availability in the Megaport Locations API v2.
	//
	// Deprecated: Use FilterLocationsByMcrAvailabilityV3 instead. The v2 API will be removed in a future version.
	FilterLocationsByMcrAvailability(ctx context.Context, mcrAvailable bool, locations []*Location) []*Location
}

// LocationServiceOp handles communication with Location methods of the Megaport API.
type LocationServiceOp struct {
	Client *Client
}

// NewLocationService creates a new instance of the Location Service.
func NewLocationService(c *Client) *LocationServiceOp {
	return &LocationServiceOp{
		Client: c,
	}
}

// LocationV3 represents a location in the Megaport API v3.
// This struct should be used for all new implementations.
type LocationV3 struct {
	// Core location identifiers (preserved from v2)
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Metro  string `json:"metro"`
	Market string `json:"market"`
	Status string `json:"status"`

	// Geographic information (preserved from v2)
	Address   LocationV3Address `json:"address"`
	Latitude  float64           `json:"latitude"`
	Longitude float64           `json:"longitude"`

	// Data center information (NEW in v3)
	DataCentre LocationV3DataCentre `json:"dataCentre"`

	// Product and availability information (NEW structure in v3)
	DiversityZones *LocationV3DiversityZones `json:"diversityZones"`
	ProductAddOns  *LocationV3ProductAddOns  `json:"productAddOns"`

	OrderingMessage *string `json:"orderingMessage"`

	// DEPRECATED/REMOVED FIELDS from v2 that are no longer available in v3:
	// - NetworkRegion: No longer provided in v3 API
	// - SiteCode: No longer provided in v3 API
	// - Campus: No longer provided in v3 API (was deprecated in v2)
	// - LiveDate: No longer provided in v3 API
	// - VRouterAvailable: No longer provided in v3 API
	// - Products: Completely restructured as DiversityZones in v3
}

// LocationV3Address represents the address structure in v3 API
type LocationV3Address struct {
	Street   string `json:"street"`
	Suburb   string `json:"suburb"`
	City     string `json:"city"`
	State    string `json:"state"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`
}

// LocationV3DataCentre represents data center information in v3 API
type LocationV3DataCentre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// LocationV3DiversityZones represents the diversity zones and product availability in v3 API
type LocationV3DiversityZones struct {
	Red  *LocationV3DiversityZone `json:"red,omitempty"`
	Blue *LocationV3DiversityZone `json:"blue,omitempty"`
}

// LocationV3DiversityZone represents a single diversity zone with product availability
type LocationV3DiversityZone struct {
	McrSpeedMbps       []int `json:"mcrSpeedMbps,omitempty"`
	MegaportSpeedMbps  []int `json:"megaportSpeedMbps,omitempty"`
	MveMaxCpuCoreCount *int  `json:"mveMaxCpuCoreCount,omitempty"`
	MveAvailable       bool  `json:"mveAvailable"`
}

// LocationV3ProductAddOns represents additional product options available at the location
type LocationV3ProductAddOns struct {
	CrossConnect *LocationV3CrossConnect `json:"crossConnect,omitempty"`
}

// LocationV3CrossConnect represents cross-connect availability and type
type LocationV3CrossConnect struct {
	Available bool    `json:"available"`
	Type      *string `json:"type,omitempty"`
}

// LocationsResponse represents the response from the Megaport Locations API.
type LocationResponse struct {
	Message string      `json:"message"`
	Terms   string      `json:"terms"`
	Data    []*Location `json:"data"`
}

// LocationV3Response represents the response from the Megaport Locations API v3.
type LocationV3Response struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    []*LocationV3 `json:"data"`
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

// ProductLocationDetails represents the location details of a product.
type ProductLocationDetails struct {
	Name    string `json:"name"`
	City    string `json:"city"`
	Metro   string `json:"metro"`
	Country string `json:"country"`
}

// ==========================================
// V3 API IMPLEMENTATION METHODS
// ==========================================

// ListLocationsV3 returns a list of all locations using the v3 API.
func (svc *LocationServiceOp) ListLocationsV3(ctx context.Context) ([]*LocationV3, error) {
	path := "/v3/locations"
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

	locationResponse := &LocationV3Response{}

	unmarshalErr := json.Unmarshal(body, locationResponse)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return locationResponse.Data, nil
}

// GetLocationByIDV3 returns a location by its ID using the v3 API.
func (svc *LocationServiceOp) GetLocationByIDV3(ctx context.Context, locationID int) (*LocationV3, error) {
	allLocations, locErr := svc.ListLocationsV3(ctx)
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

// GetLocationByNameV3 returns a location by its name using the v3 API.
func (svc *LocationServiceOp) GetLocationByNameV3(ctx context.Context, locationName string) (*LocationV3, error) {
	allLocations, locErr := svc.ListLocationsV3(ctx)
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

// GetLocationByNameFuzzyV3 returns locations by name using fuzzy search with the v3 API.
func (svc *LocationServiceOp) GetLocationByNameFuzzyV3(ctx context.Context, search string) ([]*LocationV3, error) {
	locations, err := svc.ListLocationsV3(ctx)
	if err != nil {
		return nil, err
	}
	var matchedLocations []*LocationV3

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

// FilterLocationsByMarketCodeV3 filters locations by market code using the v3 API.
func (svc *LocationServiceOp) FilterLocationsByMarketCodeV3(ctx context.Context, marketCode string, locations []*LocationV3) ([]*LocationV3, error) {
	existingLocations := locations
	toReturn := []*LocationV3{}
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

// FilterLocationsByMcrAvailabilityV3 filters locations by MCR availability using the v3 API.
func (svc *LocationServiceOp) FilterLocationsByMcrAvailabilityV3(ctx context.Context, mcrAvailable bool, locations []*LocationV3) []*LocationV3 {
	existingLocations := locations
	toReturn := []*LocationV3{}
	for _, location := range existingLocations {
		hasAnyMCR := location.HasMCRSupport()
		if hasAnyMCR == mcrAvailable {
			toReturn = append(toReturn, location)
		}
	}
	return toReturn
}

// FilterLocationsByMetroV3 filters locations by metro name using the v3 API.
func (svc *LocationServiceOp) FilterLocationsByMetroV3(ctx context.Context, metro string, locations []*LocationV3) []*LocationV3 {
	toReturn := []*LocationV3{}
	for _, loc := range locations {
		if loc.Metro == metro {
			toReturn = append(toReturn, loc)
		}
	}
	return toReturn
}

// ==========================================
// HELPER METHODS FOR LOCATIONV3
// ==========================================

// HasMCRSupport checks if the location supports MCR based on v3 diversity zones.
func (l *LocationV3) HasMCRSupport() bool {
	if l.DiversityZones == nil {
		return false
	}

	// Check if either red or blue zones have MCR support
	if l.DiversityZones.Red != nil && len(l.DiversityZones.Red.McrSpeedMbps) > 0 {
		return true
	}
	if l.DiversityZones.Blue != nil && len(l.DiversityZones.Blue.McrSpeedMbps) > 0 {
		return true
	}

	return false
}

// GetMCRSpeeds returns all available MCR speeds from both diversity zones.
func (l *LocationV3) GetMCRSpeeds() []int {
	var allSpeeds []int

	if l.DiversityZones != nil {
		if l.DiversityZones.Red != nil {
			allSpeeds = append(allSpeeds, l.DiversityZones.Red.McrSpeedMbps...)
		}
		if l.DiversityZones.Blue != nil {
			allSpeeds = append(allSpeeds, l.DiversityZones.Blue.McrSpeedMbps...)
		}
	}

	// Remove duplicates
	speedMap := make(map[int]bool)
	var uniqueSpeeds []int
	for _, speed := range allSpeeds {
		if !speedMap[speed] {
			speedMap[speed] = true
			uniqueSpeeds = append(uniqueSpeeds, speed)
		}
	}

	return uniqueSpeeds
}

// GetMegaportSpeeds returns all available Megaport speeds from both diversity zones.
func (l *LocationV3) GetMegaportSpeeds() []int {
	var allSpeeds []int

	if l.DiversityZones != nil {
		if l.DiversityZones.Red != nil {
			allSpeeds = append(allSpeeds, l.DiversityZones.Red.MegaportSpeedMbps...)
		}
		if l.DiversityZones.Blue != nil {
			allSpeeds = append(allSpeeds, l.DiversityZones.Blue.MegaportSpeedMbps...)
		}
	}

	// Remove duplicates
	speedMap := make(map[int]bool)
	var uniqueSpeeds []int
	for _, speed := range allSpeeds {
		if !speedMap[speed] {
			speedMap[speed] = true
			uniqueSpeeds = append(uniqueSpeeds, speed)
		}
	}

	return uniqueSpeeds
}

// HasMVESupport checks if the location supports MVE.
func (l *LocationV3) HasMVESupport() bool {
	if l.DiversityZones == nil {
		return false
	}

	// Check if either red or blue zones have MVE support
	if l.DiversityZones.Red != nil && l.DiversityZones.Red.MveAvailable {
		return true
	}
	if l.DiversityZones.Blue != nil && l.DiversityZones.Blue.MveAvailable {
		return true
	}

	return false
}

// GetMVEMaxCpuCores returns the maximum MVE CPU cores available across all zones.
func (l *LocationV3) GetMVEMaxCpuCores() *int {
	var maxCores int
	var hasValue bool

	if l.DiversityZones != nil {
		if l.DiversityZones.Red != nil && l.DiversityZones.Red.MveMaxCpuCoreCount != nil {
			if !hasValue || *l.DiversityZones.Red.MveMaxCpuCoreCount > maxCores {
				maxCores = *l.DiversityZones.Red.MveMaxCpuCoreCount
				hasValue = true
			}
		}
		if l.DiversityZones.Blue != nil && l.DiversityZones.Blue.MveMaxCpuCoreCount != nil {
			if !hasValue || *l.DiversityZones.Blue.MveMaxCpuCoreCount > maxCores {
				maxCores = *l.DiversityZones.Blue.MveMaxCpuCoreCount
				hasValue = true
			}
		}
	}

	if hasValue {
		return &maxCores
	}
	return nil
}

// HasCrossConnectSupport checks if the location supports cross-connects.
func (l *LocationV3) HasCrossConnectSupport() bool {
	return l.ProductAddOns != nil &&
		l.ProductAddOns.CrossConnect != nil &&
		l.ProductAddOns.CrossConnect.Available
}

// GetCrossConnectType returns the cross-connect type if available.
func (l *LocationV3) GetCrossConnectType() string {
	if l.HasCrossConnectSupport() && l.ProductAddOns.CrossConnect.Type != nil {
		return *l.ProductAddOns.CrossConnect.Type
	}
	return ""
}

// GetDataCenterName returns the data center name.
func (l *LocationV3) GetDataCenterName() string {
	return l.DataCentre.Name
}

// GetDataCenterID returns the data center ID.
func (l *LocationV3) GetDataCenterID() int {
	return l.DataCentre.ID
}

// GetCountry returns the country from the address.
func (l *LocationV3) GetCountry() string {
	return l.Address.Country
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
