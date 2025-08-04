package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// LocationIntegrationTestSuite is the integration test suite for the Location service
type LocationIntegrationTestSuite IntegrationTestSuite

func TestLocationIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(LocationIntegrationTestSuite))
	}
}

func (suite *LocationIntegrationTestSuite) SetupSuite() {
	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	megaportClient, err := New(nil, WithBaseURL(MEGAPORTURL), WithLogHandler(handler), WithCredentials(accessKey, secretKey))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	_, err = megaportClient.Authorize(ctx)
	if err != nil {
		suite.FailNowf("", "could not authorize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

// TestBadIDV3 tests the GetLocationByIDV3 method with an invalid ID.
func (suite *LocationIntegrationTestSuite) TestBadIDV3() {
	ctx := context.Background()
	// Make sure that an id with no record returns an error as expected.
	_, idErr := suite.client.LocationService.GetLocationByIDV3(ctx, -999999)
	suite.Equal(ErrLocationNotFound, idErr)
}

// TestBadNameV3 tests the GetLocationByNameV3 method with an invalid name.
func (suite *LocationIntegrationTestSuite) TestBadNameV3() {
	ctx := context.Background()

	// Make sure that a name with no record returns an error as expected.
	_, nameErr := suite.client.LocationService.GetLocationByNameV3(ctx, "DefinitelyNotARealName")
	suite.Equal(ErrLocationNotFound, nameErr)
}

// TestGetLocationByIDV3 tests the GetLocationByIDV3 method.
func (suite *LocationIntegrationTestSuite) TestGetLocationByIDV3() {
	ctx := context.Background()

	// Make sure our location by id works as expected
	byId, idErr := suite.client.LocationService.GetLocationByIDV3(ctx, 137)
	suite.Nil(idErr)
	suite.Equal("3DC/Telecity Sofia", byId.Name)

	byId, idErr = suite.client.LocationService.GetLocationByIDV3(ctx, 383)
	suite.Nil(idErr)
	suite.Equal("NextDC B2", byId.Name)
}

// TestGetLocationByNameV3 tests the GetLocationByNameV3 method.
func (suite *LocationIntegrationTestSuite) TestGetLocationByNameV3() {
	ctx := context.Background()

	// Make sure our location by name works as expected
	byName, nameErr := suite.client.LocationService.GetLocationByNameV3(ctx, "3DC/Telecity Sofia")
	suite.Nil(nameErr)
	suite.Equal(137, byName.ID)

	byName, nameErr = suite.client.LocationService.GetLocationByNameV3(ctx, "NextDC B2")
	suite.Nil(nameErr)
	suite.Equal(383, byName.ID)

	byName, nameErr = suite.client.LocationService.GetLocationByNameV3(ctx, "Equinix SY3")
	suite.Nil(nameErr)
	suite.Equal(6, byName.ID)
}

// TestGetLocationByNameFuzzyV3 tests the GetLocationByNameFuzzyV3 method.
func (suite *LocationIntegrationTestSuite) TestGetLocationByNameFuzzyV3() {
	ctx := context.Background()

	byFuzzy, fuzzyErr := suite.client.LocationService.GetLocationByNameFuzzyV3(ctx, "NextDC")
	suite.True(len(byFuzzy) > 0)
	suite.NoError(fuzzyErr)

	failFuzzy, failFuzzyErr := suite.client.LocationService.GetLocationByNameFuzzyV3(ctx, "definitely not a location name at all")
	suite.True(len(failFuzzy) == 0)
	suite.Error(ErrNoMatchingLocations, failFuzzyErr)
}

// TestListCountries tests the ListCountries method
// first one should always be Australia
func (suite *LocationIntegrationTestSuite) TestListCountries() {
	ctx := context.Background()

	countries, countriesErr := suite.client.LocationService.ListCountries(ctx)
	suite.NoError(countriesErr)
	suite.Equal("Australia", countries[0].Name)
	suite.Equal("AUS", countries[0].Code)
	suite.Equal("AU", countries[0].Prefix)
	suite.Greater(countries[0].SiteCount, 0)
}

// TestListMarketCodes tests the ListMarketCodes method
func (suite *LocationIntegrationTestSuite) TestMarketCodes() {
	ctx := context.Background()

	marketCodes, _ := suite.client.LocationService.ListMarketCodes(ctx)
	found := false

	for i := 0; i < len(marketCodes); i++ {
		if marketCodes[i] == "AU" {
			found = true
		}
	}

	suite.True(found)
}

// TestFilterLocationsByMarketCodeV3 tests the FilterLocationsByMarketCodeV3 method
func (suite *LocationIntegrationTestSuite) TestFilterLocationsByMarketCodeV3() {
	ctx := context.Background()

	locations, err := suite.client.LocationService.ListLocationsV3(ctx)
	suite.NoError(err)
	currentCount := len(locations)
	filtered, filterErr := suite.client.LocationService.FilterLocationsByMarketCodeV3(ctx, "AU", locations)
	suite.NoError(filterErr)

	suite.Less(len(filtered), currentCount)
	suite.Equal("AU", filtered[0].Market)
}

// TestFilterLocationsByMcrAvailabilityV3 tests the FilterLocationsByMcrAvailabilityV3 method
func (suite *LocationIntegrationTestSuite) TestFilterLocationsByMcrAvailabilityV3() {
	ctx := context.Background()

	locations, err := suite.client.LocationService.ListLocationsV3(ctx)
	suite.NoError(err)

	// Filter for MCR-enabled locations
	mcrEnabledLocations := suite.client.LocationService.FilterLocationsByMcrAvailabilityV3(ctx, true, locations)
	suite.Greater(len(mcrEnabledLocations), 0)

	// Verify all returned locations have MCR support
	for _, location := range mcrEnabledLocations {
		suite.True(location.HasMCRSupport())
		suite.Greater(len(location.GetMCRSpeeds()), 0)
	}

	// Filter for non-MCR locations
	nonMcrLocations := suite.client.LocationService.FilterLocationsByMcrAvailabilityV3(ctx, false, locations)

	// Verify all returned locations don't have MCR support
	for _, location := range nonMcrLocations {
		suite.False(location.HasMCRSupport())
	}
}

// TestListLocationsV3 tests the ListLocationsV3 method and validates v3-specific fields
func (suite *LocationIntegrationTestSuite) TestListLocationsV3() {
	ctx := context.Background()

	locations, err := suite.client.LocationService.ListLocationsV3(ctx)
	suite.NoError(err)
	suite.Greater(len(locations), 0)

	// Test v3-specific features on a sample location
	sampleLocation := locations[0]

	// Test basic fields
	suite.NotEmpty(sampleLocation.ID)
	suite.NotEmpty(sampleLocation.Name)
	suite.NotEmpty(sampleLocation.Market)
	suite.NotEmpty(sampleLocation.Metro)
	suite.NotEmpty(sampleLocation.Status)

	// Test address structure
	suite.NotEmpty(sampleLocation.Address.Country)
	suite.NotEmpty(sampleLocation.Address.City)

	// Test data center information
	suite.NotEmpty(sampleLocation.DataCentre.Name)
	suite.Greater(sampleLocation.DataCentre.ID, 0)

	// Test helper methods
	suite.NotEmpty(sampleLocation.GetDataCenterName())
	suite.Greater(sampleLocation.GetDataCenterID(), 0)
	suite.NotEmpty(sampleLocation.GetCountry())
}

// TestLocationV3HelperMethods tests the v3-specific helper methods
func (suite *LocationIntegrationTestSuite) TestLocationV3HelperMethods() {
	ctx := context.Background()

	locations, err := suite.client.LocationService.ListLocationsV3(ctx)
	suite.NoError(err)
	suite.Greater(len(locations), 0)

	// Find a location with MCR support for testing
	var mcrLocation *LocationV3
	for _, location := range locations {
		if location.HasMCRSupport() {
			mcrLocation = location
			break
		}
	}

	if mcrLocation != nil {
		// Test MCR-related methods
		suite.True(mcrLocation.HasMCRSupport())
		mcrSpeeds := mcrLocation.GetMCRSpeeds()
		suite.Greater(len(mcrSpeeds), 0)

		// Test Megaport speeds
		megaportSpeeds := mcrLocation.GetMegaportSpeeds()
		suite.Greater(len(megaportSpeeds), 0)

		// Test data center methods
		suite.NotEmpty(mcrLocation.GetDataCenterName())
		suite.Greater(mcrLocation.GetDataCenterID(), 0)
		suite.NotEmpty(mcrLocation.GetCountry())
	}

	// Find a location with MVE support for testing
	var mveLocation *LocationV3
	for _, location := range locations {
		if location.HasMVESupport() {
			mveLocation = location
			break
		}
	}

	if mveLocation != nil {
		suite.True(mveLocation.HasMVESupport())
		// MVE max CPU cores might be nil, so just test that the method doesn't panic
		suite.NotPanics(func() {
			mveLocation.GetMVEMaxCpuCores()
		})
	}

	// Test cross-connect support
	for _, location := range locations[:10] { // Test first 10 locations to avoid too many calls
		suite.NotPanics(func() {
			location.HasCrossConnectSupport()
			location.GetCrossConnectType()
		})
	}
}
