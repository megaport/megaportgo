package megaport

import (
	"context"
	"errors"

	"github.com/megaport/megaportgo/mega_err"
)

func (suite *IntegrationTestSuite) TestBadID() {
	ctx := context.Background()
	// Make sure that an id with no record returns an error as expected.
	_, idErr := megaportClient.LocationService.GetLocationByID(ctx, -999999)
	suite.Equal(mega_err.ERR_LOCATION_NOT_FOUND, idErr.Error())
}

func (suite *IntegrationTestSuite) TestBadName() {
	ctx := context.Background()

	// Make sure that a name with no record returns an error as expected.
	_, nameErr := megaportClient.LocationService.GetLocationByName(ctx, "DefinitelyNotARealName")
	suite.Equal(mega_err.ERR_LOCATION_NOT_FOUND, nameErr.Error())
}

func (suite *IntegrationTestSuite) TestGetLocationByID() {
	ctx := context.Background()

	// Make sure our location by id works as expected
	byId, idErr := megaportClient.LocationService.GetLocationByID(ctx, 137)
	suite.Nil(idErr)
	suite.Equal("3DC/Telecity Sofia", byId.Name)

	byId, idErr = megaportClient.LocationService.GetLocationByID(ctx, 383)
	suite.Nil(idErr)
	suite.Equal("NextDC B2", byId.Name)
}

func (suite *IntegrationTestSuite) TestGetLocationByName() {
	ctx := context.Background()

	// Make sure our location by name works as expected
	byName, nameErr := megaportClient.LocationService.GetLocationByName(ctx, "3DC/Telecity Sofia")
	suite.Nil(nameErr)
	suite.Equal(137, byName.ID)

	byName, nameErr = megaportClient.LocationService.GetLocationByName(ctx, "NextDC B2")
	suite.Nil(nameErr)
	suite.Equal(383, byName.ID)

	byName, nameErr = megaportClient.LocationService.GetLocationByName(ctx, "Equinix SY3")
	suite.Nil(nameErr)
	suite.Equal(6, byName.ID)
}

func (suite *IntegrationTestSuite) TestGetLocationByNameFuzzy() {
	ctx := context.Background()

	byFuzzy, fuzzyErr := megaportClient.LocationService.GetLocationByNameFuzzy(ctx, "NextDC")
	suite.True(len(byFuzzy) > 0)
	suite.NoError(fuzzyErr)

	failFuzzy, failFuzzyErr := megaportClient.LocationService.GetLocationByNameFuzzy(ctx, "definitely not a location name at all")
	suite.True(len(failFuzzy) == 0)
	suite.Error(errors.New(mega_err.ERR_NO_MATCHING_LOCATIONS), failFuzzyErr)
}

// first one should always be Australia
func (suite *IntegrationTestSuite) TestListCountries() {
	ctx := context.Background()

	countries, countriesErr := megaportClient.LocationService.ListCountries(ctx)
	suite.NoError(countriesErr)
	suite.Equal("Australia", countries[0].Name)
	suite.Equal("AUS", countries[0].Code)
	suite.Equal("AU", countries[0].Prefix)
	suite.Greater(countries[0].SiteCount, 0)
}

func (suite *IntegrationTestSuite) TestMarketCodes() {
	ctx := context.Background()

	marketCodes, _ := megaportClient.LocationService.ListMarketCodes(ctx)
	found := false

	for i := 0; i < len(marketCodes); i++ {
		if marketCodes[i] == "AU" {
			found = true
		}
	}

	suite.True(found)
}

func (suite *IntegrationTestSuite) TestFilterLocationsByMarketCode() {
	ctx := context.Background()

	locations, err := megaportClient.LocationService.ListLocations(ctx)
	suite.NoError(err)
	currentCount := len(locations)
	filtered, filterErr := megaportClient.LocationService.FilterLocationsByMarketCode(ctx, "AU", locations)
	suite.NoError(filterErr)

	suite.Less(len(filtered), currentCount)
	suite.Equal("AU", filtered[0].Market)
}
