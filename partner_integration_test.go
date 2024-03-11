package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// PartnerIntegrationTestSuite is the integration test suite for the Partner service
type PartnerIntegrationTestSuite IntegrationTestSuite

func TestPartnerIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(PartnerIntegrationTestSuite))
	}
}

func (suite *PartnerIntegrationTestSuite) SetupSuite() {
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

// TestListPartnerMegaports tests the ListPartnerMegaports method.
func (suite *PartnerIntegrationTestSuite) TestListPartnerMegaports() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	_, err := partnerSvc.ListPartnerMegaports(ctx)
	if err != nil {
		suite.FailNowf("could not list partners", "could not list partners %v", err)
	}
}

// TestFilterPartnerMegaportByCompanyName tests the FilterPartnerMegaportByCompanyName method.
func (suite *PartnerIntegrationTestSuite) TestFilterPartnerMegaportByCompanyName() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	if err != nil {
		suite.FailNowf("could not list partners", "could not list partners %v", err)
	}
	filtered, err := partnerSvc.FilterPartnerMegaportByCompanyName(ctx, partners, "AWS", true)
	if err != nil {
		suite.FailNowf("could not filter partners", "could not filter partners %v", err)
	}
	for _, partner := range filtered {
		suite.Equal(partner.CompanyName, "AWS")
	}
}

// TestFilterPartnerMegaportByLocationId tests the FilterPartnerMegaportByLocationId method.
func (suite *PartnerIntegrationTestSuite) TestFilterPartnerMegaportByLocationId() {
	partnerSvc := suite.client.PartnerService
	locSvc := suite.client.LocationService
	ctx := context.Background()

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	if err != nil {
		suite.FailNowf("could not list partners", "could not list partners %v", err)
	}
	location, err := locSvc.GetLocationByName(ctx, "Equinix SY3")
	if err != nil {
		suite.FailNowf("could not get location", "could not get location %v", err)
	}
	filtered, err := partnerSvc.FilterPartnerMegaportByLocationId(ctx, partners, location.ID)
	if err != nil {
		suite.FailNowf("could not filter partners", "could not filter partners %v", err)
	}
	for _, partner := range filtered {
		suite.Equal(partner.LocationId, location.ID)
	}
}

// TestFilterPartnerMegaportByConnectType tests the FilterPartnerMegaportByConnectType method.
func (suite *PartnerIntegrationTestSuite) TestFilterPartnerMegaportByConnectType() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	if err != nil {
		suite.FailNowf("could not list partners", "could not list partners %v", err)
	}
	filtered, err := partnerSvc.FilterPartnerMegaportByConnectType(ctx, partners, CONNECT_TYPE_AWS_VIF, true)
	if err != nil {
		suite.FailNowf("could not filter partners", "could not filter partners %v", err)
	}
	for _, partner := range filtered {
		suite.Equal(partner.ConnectType, CONNECT_TYPE_AWS_VIF)
	}
	filtered2, err := partnerSvc.FilterPartnerMegaportByConnectType(ctx, partners, CONNECT_TYPE_AWS_HOSTED_CONNECTION, true)
	if err != nil {
		suite.FailNowf("could not filter partners", "could not filter partners %v", err)
	}
	for _, partner := range filtered2 {
		suite.Equal(partner.ConnectType, CONNECT_TYPE_AWS_HOSTED_CONNECTION)
	}
}

// TestFilterPartnerMegaportByProductName tests the FilterPartnerMegaportByProductName method.
func (suite *PartnerIntegrationTestSuite) TestFilterPartnerMegaportByProductName() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	if err != nil {
		suite.FailNowf("could not list partners", "could not list partners %v", err)
	}
	productName := partners[0].ProductName
	filtered, err := partnerSvc.FilterPartnerMegaportByProductName(ctx, partners, productName, true)
	if err != nil {
		suite.FailNowf("could not filter partners", "could not filter partners %v", err)
	}
	for _, partner := range filtered {
		suite.Equal(partner.ProductName, productName)
	}
}

// TestFilterPartnerMegaportByDiversityZone tests the FilterPartnerMegaportByDiversityZone method.
func (suite *PartnerIntegrationTestSuite) TestFilterPartnerMegaportByDiversityZone() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	if err != nil {
		suite.FailNowf("could not list partners", "could not list partners %v", err)
	}
	filtered, err := partnerSvc.FilterPartnerMegaportByDiversityZone(ctx, partners, "red", true)
	if err != nil {
		suite.FailNowf("could not filter partners", "could not filter partners %v", err)
	}
	for _, partner := range filtered {
		suite.Equal(partner.DiversityZone, "red")
	}
}
