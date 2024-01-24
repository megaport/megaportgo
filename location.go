package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
)

type LocationService interface {
	ListLocations(ctx context.Context) ([]types.Location, error)
	GetLocationByID(ctx context.Context, locationID int) (*types.Location, error)
	GetLocationByName(ctx context.Context, locationName string) (*types.Location, error)
	GetLocationByNameFuzzy(ctx context.Context, search string) ([]types.Location, error)
	ListCountries(ctx context.Context) ([]types.Country, error)
	ListMarketCodes(ctx context.Context) ([]string, error)
	IsValidMarketCode(ctx context.Context, marketCode string) (*bool, error)
	FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations *[]types.Location) error
}

type LocationServiceOp struct {
	Client *Client
}

func NewLocationServiceOp(c *Client) *LocationServiceOp {
	return &LocationServiceOp{
		Client: c,
	}
}

func (svc *LocationServiceOp) ListLocations(ctx context.Context) ([]types.Location, error) {
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
	isResErr, compiledResError := svc.Client.IsErrorResponse(response, &resErr, 200)
	if isResErr {
		return nil, compiledResError
	}
	body, fileErr := io.ReadAll(response.Body)

	if fileErr != nil {
		return nil, fileErr
	}

	locationResponse := &types.LocationResponse{}

	unmarshalErr := json.Unmarshal(body, locationResponse)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return locationResponse.Data, nil
}

func (svc *LocationServiceOp) GetLocationByID(ctx context.Context, locationID int) (*types.Location, error) {
	allLocations, locErr := svc.ListLocations(ctx)
	if locErr != nil {
		return nil, locErr
	}
	for _, location := range allLocations {
		if location.ID == locationID {
			return &location, nil
		}
	}
	return nil, errors.New(mega_err.ERR_LOCATION_NOT_FOUND)
}

func (svc *LocationServiceOp) GetLocationByName(ctx context.Context, locationName string) (*types.Location, error) {
	allLocations, locErr := svc.ListLocations(ctx)
	if locErr != nil {
		return nil, locErr
	}
	for _, location := range allLocations {
		if location.Name == locationName {
			return &location, nil
		}
	}
	return nil, errors.New(mega_err.ERR_LOCATION_NOT_FOUND)
}

func (svc *LocationServiceOp) GetLocationByNameFuzzy(ctx context.Context, search string) ([]types.Location, error) {
	locations, err := svc.ListLocations(ctx)
	if err != nil {
		return nil, err
	}
	var matchedLocations []types.Location

	for i := 0; i < len(locations); i++ {
		if fuzzy.Match(search, locations[i].Name) {
			matchedLocations = append(matchedLocations, locations[i])
		}
	}

	if len(matchedLocations) > 0 {
		return matchedLocations, nil
	} else {
		return matchedLocations, errors.New(mega_err.ERR_NO_MATCHING_LOCATIONS)
	}
}

func (svc *LocationServiceOp) ListCountries(ctx context.Context) ([]types.Country, error) {
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
	isResErr, compiledResError := svc.Client.IsErrorResponse(response, &resErr, 200)
	if isResErr {
		return nil, compiledResError
	}

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	countryResponse := types.CountryResponse{}

	unmarshalErr := json.Unmarshal(body, &countryResponse)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	allCountries := make([]types.Country, 0)

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

func (svc *LocationServiceOp) FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations *[]types.Location) error {
	existingLocations := *locations
	*locations = nil
	isValid, err := svc.IsValidMarketCode(ctx, marketCode)
	if err != nil {
		return err
	}
	if *isValid {
		for i := 0; i < len(existingLocations); i++ {
			if existingLocations[i].Market == marketCode {
				*locations = append(*locations, existingLocations[i])
			}
		}
	}
	return nil
}
