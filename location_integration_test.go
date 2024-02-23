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
	if *runIntegrationTests {
		suite.Run(t, new(LocationIntegrationTestSuite))
	}
}

func (suite *LocationIntegrationTestSuite) SetupSuite() {
	accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	var err error

	megaportClient, err = New(nil, SetBaseURL(MEGAPORTURL), SetLogHandler(handler))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

func (suite *LocationIntegrationTestSuite) SetupTest() {
	suite.client.Logger.Debug("logging in")
	if accessKey == "" {
		suite.client.Logger.Error("MEGAPORT_ACCESS_KEY environment variable not set.")
		os.Exit(1)
	}

	if secretKey == "" {
		suite.client.Logger.Error("MEGAPORT_SECRET_KEY environment variable not set.")
		os.Exit(1)
	}

	ctx := context.Background()
	loginResp, loginErr := suite.client.AuthenticationService.Login(ctx, &LoginRequest{
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if loginErr != nil {
		suite.client.Logger.Error("login error", "error", loginErr.Error())
		suite.FailNowf("login error", "login error %v", loginErr)
	}

	// Session Token is not empty
	if !suite.NotEmpty(loginResp.Token) {
		suite.FailNow("empty token")
	}

	// SessionToken is a valid guid
	if !suite.NotNil(IsGuid(loginResp.Token)) {
		suite.FailNowf("invalid guid for token", "invalid guid for token %v", loginResp.Token)
	}

	suite.client.SessionToken = loginResp.Token
}

// TestBadID tests the GetLocationByID method with an invalid ID.
func (suite *LocationIntegrationTestSuite) TestBadID() {
	ctx := context.Background()
	// Make sure that an id with no record returns an error as expected.
	_, idErr := suite.client.LocationService.GetLocationByID(ctx, -999999)
	suite.Equal(ErrLocationNotFound, idErr)
}

// TestBadName tests the GetLocationByName method with an invalid name.
func (suite *LocationIntegrationTestSuite) TestBadName() {
	ctx := context.Background()

	// Make sure that a name with no record returns an error as expected.
	_, nameErr := suite.client.LocationService.GetLocationByName(ctx, "DefinitelyNotARealName")
	suite.Equal(ErrLocationNotFound, nameErr)
}

// TestGetLocationByID tests the GetLocationByID method.
func (suite *LocationIntegrationTestSuite) TestGetLocationByID() {
	ctx := context.Background()

	// Make sure our location by id works as expected
	byId, idErr := suite.client.LocationService.GetLocationByID(ctx, 137)
	suite.Nil(idErr)
	suite.Equal("3DC/Telecity Sofia", byId.Name)

	byId, idErr = suite.client.LocationService.GetLocationByID(ctx, 383)
	suite.Nil(idErr)
	suite.Equal("NextDC B2", byId.Name)
}

// TestGetLocationByName tests the GetLocationByName method.
func (suite *LocationIntegrationTestSuite) TestGetLocationByName() {
	ctx := context.Background()

	// Make sure our location by name works as expected
	byName, nameErr := suite.client.LocationService.GetLocationByName(ctx, "3DC/Telecity Sofia")
	suite.Nil(nameErr)
	suite.Equal(137, byName.ID)

	byName, nameErr = suite.client.LocationService.GetLocationByName(ctx, "NextDC B2")
	suite.Nil(nameErr)
	suite.Equal(383, byName.ID)

	byName, nameErr = suite.client.LocationService.GetLocationByName(ctx, "Equinix SY3")
	suite.Nil(nameErr)
	suite.Equal(6, byName.ID)
}

// TestGetLocationByNameFuzzy tests the GetLocationByNameFuzzy method.
func (suite *LocationIntegrationTestSuite) TestGetLocationByNameFuzzy() {
	ctx := context.Background()

	byFuzzy, fuzzyErr := suite.client.LocationService.GetLocationByNameFuzzy(ctx, "NextDC")
	suite.True(len(byFuzzy) > 0)
	suite.NoError(fuzzyErr)

	failFuzzy, failFuzzyErr := suite.client.LocationService.GetLocationByNameFuzzy(ctx, "definitely not a location name at all")
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

// TestFilterLocationsByMarketCode tests the FilterLocationsByMarketCode method
func (suite *IntegrationTestSuite) TestFilterLocationsByMarketCode() {
	ctx := context.Background()

	locations, err := suite.client.LocationService.ListLocations(ctx)
	suite.NoError(err)
	currentCount := len(locations)
	filtered, filterErr := suite.client.LocationService.FilterLocationsByMarketCode(ctx, "AU", locations)
	suite.NoError(filterErr)

	suite.Less(len(filtered), currentCount)
	suite.Equal("AU", filtered[0].Market)
}
