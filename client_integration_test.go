package megaport

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	client *Client
}

type ClientIntegrationTestSuite IntegrationTestSuite

func TestClientIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(ClientIntegrationTestSuite))
	}
}

func (suite *ClientIntegrationTestSuite) SetupSuite() {
	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")

	httpClient := &http.Client{}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	var err error

	megaportClient, err := New(httpClient, WithBaseURL(MEGAPORTURL), WithLogHandler(handler), WithCredentials(accessKey, secretKey))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	// NOTE: unlike other test suites we don't do the authorization here because we want to test that functionality

	suite.client = megaportClient
}

func (suite *ClientIntegrationTestSuite) TestLogin() {
	ctx := context.Background()
	resp, err := suite.client.Authorize(ctx)
	if err != nil {
		suite.client.Logger.Error("login error", "error", err.Error())
	}
	suite.NoError(err)

	// Session Token is not empty
	suite.NotEmpty(resp.AccessToken)
	suite.NotZero(resp.Expiration)

	// Internal is same as returned
	suite.Equal(resp.AccessToken, suite.client.accessToken)
	suite.Equal(resp.Expiration, suite.client.tokenExpiry)
}
