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

package location

import (
	"errors"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBadID(t *testing.T) {
	assert := assert.New(t)
	// Make sure that an id with no record returns an error as expected.
	_, idErr := GetLocationByID(-999999)
	assert.Equal(mega_err.ERR_LOCATION_NOT_FOUND, idErr.Error())
}

func TestBadName(t *testing.T) {
	// Make sure that a name with no record returns an error as expected.
	assert := assert.New(t)
	_, nameErr := GetLocationByName("DefinitelyNotARealName")
	assert.Equal(mega_err.ERR_LOCATION_NOT_FOUND, nameErr.Error())
}

func TestGetLocationByID(t *testing.T) {
	assert := assert.New(t)
	// Make sure our location by id works as expected
	byId, idErr := GetLocationByID(137)
	assert.Nil(idErr)
	assert.Equal("3DC/Telecity Sofia", byId.Name)

	byId, idErr = GetLocationByID(383)
	assert.Nil(idErr)
	assert.Equal("NextDC B2", byId.Name)
}

func TestGetLocationByName(t *testing.T) {
	assert := assert.New(t)
	// Make sure our location by name works as expected
	byName, nameErr := GetLocationByName("3DC/Telecity Sofia")
	assert.Nil(nameErr)
	assert.Equal(137, byName.ID)

	byName, nameErr = GetLocationByName("NextDC B2")
	assert.Nil(nameErr)
	assert.Equal(383, byName.ID)

	byName, nameErr = GetLocationByName("Equinix SY3")
	assert.Nil(nameErr)
	assert.Equal(6, byName.ID)
}

func TestGetLocationByNameFuzzy(t *testing.T) {
	assert := assert.New(t)
	byFuzzy, fuzzyErr := GetLocationByNameFuzzy("NextDC")
	assert.True(len(byFuzzy) > 0)
	assert.NoError(fuzzyErr)

	failFuzzy, failFuzzyErr := GetLocationByNameFuzzy("definitely not a location name at all")
	assert.True(len(failFuzzy) == 0)
	assert.Error(errors.New(mega_err.ERR_NO_MATCHING_LOCATIONS), failFuzzyErr)
}

// first one should always be Australia
func TestGetCountries(t *testing.T) {
	assert := assert.New(t)
	countries, countriesErr := GetCountries()
	assert.NoError(countriesErr)
	assert.Equal("Australia", countries[0].Name)
	assert.Equal("AUS", countries[0].Code)
	assert.Equal("AU", countries[0].Prefix)
	assert.Greater(countries[0].SiteCount, 0)
}

func TestMarketCodes(t *testing.T) {
	assert := assert.New(t)
	marketCodes, _ := GetMarketCodes()
	found := false

	for i := 0; i < len(marketCodes); i++ {
		if marketCodes[i] == "AU" {
			found = true
		}
	}

	assert.True(found)
}

func TestFilterLocationsByMarketCode(t *testing.T) {
	assert := assert.New(t)
	locations, _ := GetAllLocations()
	currentCount := len(locations)
	FilterLocationsByMarketCode("AU", &locations)
	assert.Less(len(locations), currentCount)
	assert.Equal("AU", locations[0].Market)
}

func TestFilterLocationsByMcrAvailability(t *testing.T) {
	assert := assert.New(t)
	locations, _ := GetAllLocations()
	currentCount := len(locations)
	FilterLocationsByMcrAvailability(true, &locations)
	assert.Less(len(locations), currentCount)
}
