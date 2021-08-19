// +build integration

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
	"os"
	"testing"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/service/authentication"
	"github.com/stretchr/testify/assert"
)

var logger *config.DefaultLogger
var cfg config.Config

const (
	MEGAPORTURL = "https://api-staging.megaport.com/"
)

func TestMain(m *testing.M) {
	logger = config.NewDefaultLogger()
	logger.SetLevel(config.DebugLevel)

	username := os.Getenv("MEGAPORT_USERNAME")
	password := os.Getenv("MEGAPORT_PASSWORD")
	otp := os.Getenv("MEGAPORT_MFA_OTP_KEY")
	logLevel := os.Getenv("LOG_LEVEL")

	if logLevel != "" {
		logger.SetLevel(config.StringToLogLevel(logLevel))
	}

	if username == "" {
		logger.Error("MEGAPORT_USERNAME environment variable not set.")
		os.Exit(1)
	}

	if password == "" {
		logger.Error("MEGAPORT_PASSWORD environment variable not set.")
		os.Exit(1)
	}

	cfg = config.Config{
		Log:      logger,
		Endpoint: MEGAPORTURL,
	}

	auth := authentication.New(&cfg, username, password, otp)
	token, loginErr := auth.Login()

	if loginErr != nil {
		logger.Errorf("LoginError: %s", loginErr.Error())
	}

	cfg.SessionToken = token
	os.Exit(m.Run())
}

func TestBadID(t *testing.T) {
	assert := assert.New(t)
	loc := New(&cfg)

	// Make sure that an id with no record returns an error as expected.
	_, idErr := loc.GetLocationByID(-999999)
	assert.Equal(mega_err.ERR_LOCATION_NOT_FOUND, idErr.Error())
}

func TestBadName(t *testing.T) {
	assert := assert.New(t)
	loc := New(&cfg)

	// Make sure that a name with no record returns an error as expected.
	_, nameErr := loc.GetLocationByName("DefinitelyNotARealName")
	assert.Equal(mega_err.ERR_LOCATION_NOT_FOUND, nameErr.Error())
}

func TestGetLocationByID(t *testing.T) {
	assert := assert.New(t)
	loc := New(&cfg)

	// Make sure our location by id works as expected
	byId, idErr := loc.GetLocationByID(137)
	assert.Nil(idErr)
	assert.Equal("3DC/Telecity Sofia", byId.Name)

	byId, idErr = loc.GetLocationByID(383)
	assert.Nil(idErr)
	assert.Equal("NextDC B2", byId.Name)
}

func TestGetLocationByName(t *testing.T) {
	assert := assert.New(t)
	loc := New(&cfg)

	// Make sure our location by name works as expected
	byName, nameErr := loc.GetLocationByName("3DC/Telecity Sofia")
	assert.Nil(nameErr)
	assert.Equal(137, byName.ID)

	byName, nameErr = loc.GetLocationByName("NextDC B2")
	assert.Nil(nameErr)
	assert.Equal(383, byName.ID)

	byName, nameErr = loc.GetLocationByName("Equinix SY3")
	assert.Nil(nameErr)
	assert.Equal(6, byName.ID)
}

func TestGetLocationByNameFuzzy(t *testing.T) {
	assert := assert.New(t)
	loc := New(&cfg)

	byFuzzy, fuzzyErr := loc.GetLocationByNameFuzzy("NextDC")
	assert.True(len(byFuzzy) > 0)
	assert.NoError(fuzzyErr)

	failFuzzy, failFuzzyErr := loc.GetLocationByNameFuzzy("definitely not a location name at all")
	assert.True(len(failFuzzy) == 0)
	assert.Error(errors.New(mega_err.ERR_NO_MATCHING_LOCATIONS), failFuzzyErr)
}

// first one should always be Australia
func TestGetCountries(t *testing.T) {
	assert := assert.New(t)
	loc := New(&cfg)

	countries, countriesErr := loc.GetCountries()
	assert.NoError(countriesErr)
	assert.Equal("Australia", countries[0].Name)
	assert.Equal("AUS", countries[0].Code)
	assert.Equal("AU", countries[0].Prefix)
	assert.Greater(countries[0].SiteCount, 0)
}

func TestMarketCodes(t *testing.T) {
	assert := assert.New(t)
	loc := New(&cfg)

	marketCodes, _ := loc.GetMarketCodes()
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
	loc := New(&cfg)

	locations, _ := loc.GetAllLocations()
	currentCount := len(locations)
	loc.FilterLocationsByMarketCode("AU", &locations)

	assert.Less(len(locations), currentCount)
	assert.Equal("AU", locations[0].Market)
}

func TestFilterLocationsByMcrAvailability(t *testing.T) {
	assert := assert.New(t)
	loc := New(&cfg)

	locations, _ := loc.GetAllLocations()
	currentCount := len(locations)
	loc.FilterLocationsByMcrAvailability(true, &locations)
	assert.Less(len(locations), currentCount)
}
