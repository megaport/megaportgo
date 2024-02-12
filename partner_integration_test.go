package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PartnerIntegrationTestSuite IntegrationTestSuite

func TestPartnerIntegrationTestSuite(t *testing.T) {
	if *runIntegrationTests {
		suite.Run(t, new(PartnerIntegrationTestSuite))
	}
}

func (suite *PartnerIntegrationTestSuite) SetupSuite() {
	accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

	httpClient := NewHttpClient()

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	var err error

	megaportClient, err = New(httpClient, SetBaseURL(MEGAPORTURL), SetLogHandler(handler))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

func (suite *PartnerIntegrationTestSuite) SetupTest() {
	suite.client.Logger.Debug("logging in oauth")
	if accessKey == "" {
		suite.FailNow("MEGAPORT_ACCESS_KEY environment variable not set.")
	}

	if secretKey == "" {
		suite.FailNow("MEGAPORT_SECRET_KEY environment variable not set.")
	}

	ctx := context.Background()
	loginResp, loginErr := suite.client.AuthenticationService.LoginOauth(ctx, &LoginOauthRequest{
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if loginErr != nil {
		suite.client.Logger.Error("login error", slog.String("error", loginErr.Error()))
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

func (suite *PartnerIntegrationTestSuite) TestListPartnerMegaports() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	_, err := partnerSvc.ListPartnerMegaports(ctx)
	if err != nil {
		suite.FailNowf("could not list partners", "could not list partners %v", err)
	}
}

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