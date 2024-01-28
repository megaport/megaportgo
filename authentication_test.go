package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/megaport/megaportgo/shared"
	"github.com/stretchr/testify/suite"
)

var accessKey string
var secretKey string

var megaportClient *Client

var programLevel = new(slog.LevelVar)

const (
	MEGAPORTURL = "https://api-staging.megaport.com/"
)

type IntegrationTestSuite TestSuite

func TestIntegrationTestSuite(t *testing.T) {
	if os.Getenv("CI") != "true" {
		suite.Run(t, new(IntegrationTestSuite))
	}
}

func (suite *IntegrationTestSuite) SetupSuite() {
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
}
func (suite *IntegrationTestSuite) TestLoginOauth() {
	megaportClient.Logger.Debug("testing login oauth")
	if accessKey == "" {
		megaportClient.Logger.Error("MEGAPORT_ACCESS_KEY environment variable not set.")
		os.Exit(1)
	}

	if secretKey == "" {
		megaportClient.Logger.Error("MEGAPORT_SECRET_KEY environment variable not set.")
		os.Exit(1)
	}

	ctx := context.Background()
	token, loginErr := megaportClient.AuthenticationService.LoginOauth(ctx, accessKey, secretKey)
	if loginErr != nil {
		megaportClient.Logger.Error("login error", "error", loginErr.Error())
	}
	suite.NoError(loginErr)

	// Session Token is not empty
	suite.NotEmpty(token)
	// SessionToken is a valid guid
	suite.NotNil(shared.IsGuid(token))

	megaportClient.Logger.Info("", "token", token)
	megaportClient.SessionToken = token
}
