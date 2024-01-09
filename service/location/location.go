// Copyright 2020 Megaport Pty Ltd
//
// Licensed under the Mozilla Public License, Version 2.0 (the
// "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//       https://mozilla.org/MPL/2.0/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The `location` package is used to do lookups of locations. This has some simple lookups, including finding a location
// by it's ID and a simple EXACT name lookup.
//
// If you want to find a location, you can use https://www.megaport.com/megaport-enabled-locations/ and then simply
// do a name lookup on the "Location" field from this page. For more complex lookups, please use GetAllLocations()
// to iterate through and find the location you need.
package location

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
)

type Location struct {
	*config.Config
}

// NewLocation
func New(cfg *config.Config) *Location {
	return &Location{
		Config: cfg,
	}
}

// GetLocationByID looks up locations based on the IDs that are exposed by the API. These IDs can be found by querying
// the API directly or iterating over GetAllLocations.
func (l *Location) GetLocationByID(locationID int) (types.Location, error) {
	allLocations, locErr := l.GetAllLocations()

	if locErr != nil {
		return types.Location{}, locErr
	}

	for i := 0; i < len(allLocations); i++ {
		if allLocations[i].ID == locationID {
			return allLocations[i], nil
		}
	}

	return types.Location{}, errors.New(mega_err.ERR_LOCATION_NOT_FOUND)
}

// GetLocationByName is an exact name lookup for Megaport Locations. This is not fuzzy, if the exact Location name is
// not passed in, you will not get a result. This is supposed to return a single entry.
func (l *Location) GetLocationByName(locationName string) (types.Location, error) {
	allLocations, locErr := l.GetAllLocations()

	if locErr != nil {
		return types.Location{}, locErr
	}

	for i := 0; i < len(allLocations); i++ {
		if allLocations[i].Name == locationName {
			return allLocations[i], nil
		}
	}

	return types.Location{}, errors.New(mega_err.ERR_LOCATION_NOT_FOUND)
}

// GetAllLocations retrieves all Megaport locations from the API.
func (l *Location) GetAllLocations() ([]types.Location, error) {
	locationUrl := "/v2/locations"
	response, resErr := l.Config.MakeAPICall("GET", locationUrl, nil)
	defer response.Body.Close()

	isResErr, compiledResError := l.Config.IsErrorResponse(response, &resErr, 200)

	if isResErr {
		return nil, compiledResError
	}

	body, fileErr := io.ReadAll(response.Body)

	if fileErr != nil {
		return nil, fileErr
	}

	locationResponse := types.LocationResponse{}

	unmarshalErr := json.Unmarshal(body, &locationResponse)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return locationResponse.Data, nil
}

func (l *Location) GetLocationByNameFuzzy(search string) ([]types.Location, error) {
	locations, _ := l.GetAllLocations()
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

func (l *Location) GetCountries() ([]types.Country, error) {
	marketCodeUrl := "/v2/networkRegions"
	response, resErr := l.Config.MakeAPICall("GET", marketCodeUrl, nil)
	allCountries := make([]types.Country, 0)
	defer response.Body.Close()

	isResErr, compiledResError := l.Config.IsErrorResponse(response, &resErr, 200)

	if isResErr {
		return nil, compiledResError
	}

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return nil, fileErr
	}

	countryResponse := types.CountryResponse{}

	unmarshalErr := json.Unmarshal(body, &countryResponse)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	for i := 0; i < len(countryResponse.Data); i++ {
		if countryResponse.Data[i].NetworkRegion == "MP1" {
			allCountries = countryResponse.Data[i].Countries
		}
	}

	return allCountries, nil
}

func (l *Location) GetMarketCodes() ([]string, error) {
	countries, countriesErr := l.GetCountries()
	var marketCodes []string

	if countriesErr != nil {
		return nil, countriesErr
	}

	for i := 0; i < len(countries); i++ {
		marketCodes = append(marketCodes, countries[i].Prefix)
	}

	return marketCodes, nil
}

func (l *Location) IsValidMarketCode(marketCode string) bool {
	marketCodes, _ := l.GetMarketCodes()
	found := false

	for i := 0; i < len(marketCodes); i++ {
		if marketCodes[i] == marketCode {
			found = true
		}
	}

	return found
}

func (l *Location) FilterLocationsByMarketCode(marketCode string, locations *[]types.Location) {
	existingLocations := *locations
	*locations = nil
	if l.IsValidMarketCode(marketCode) {
		for i := 0; i < len(existingLocations); i++ {
			if existingLocations[i].Market == marketCode {
				*locations = append(*locations, existingLocations[i])
			}
		}
	}
}

func (l *Location) FilterLocationsByMcrAvailability(mcrAvailable bool, locations *[]types.Location) {
	existingLocations := *locations
	*locations = nil
	for i := 0; i < len(existingLocations); i++ {

		if _, ok := existingLocations[i].Products["mcr2"]; ok {
			*locations = append(*locations, existingLocations[i])
		}
	}
}

func (l *Location) GetRandom(marketCode string) *types.Location {
	testLocations, _ := l.GetAllLocations()
	l.FilterLocationsByMarketCode(marketCode, &testLocations)
	l.FilterLocationsByMcrAvailability(true, &testLocations)
	testLocation := testLocations[shared.GenerateRandomNumber(0, len(testLocations)-1)]
	return &testLocation
}
