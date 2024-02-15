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

type AuthIntegrationTestSuite IntegrationTestSuite

func TestAuthIntegrationTestSuite(t *testing.T) {
	if *runIntegrationTests {
		suite.Run(t, new(AuthIntegrationTestSuite))
	}
}

func (suite *AuthIntegrationTestSuite) SetupSuite() {
	accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

	httpClient := &http.Client{}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	var err error

	megaportClient, err = New(httpClient, SetBaseURL(MEGAPORTURL), SetLogHandler(handler))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

func (suite *AuthIntegrationTestSuite) TestLogin() {
	megaportClient.Logger.Debug("logging in")
	if accessKey == "" {
		megaportClient.Logger.Error("MEGAPORT_ACCESS_KEY environment variable not set.")
		os.Exit(1)
	}

	if secretKey == "" {
		megaportClient.Logger.Error("MEGAPORT_SECRET_KEY environment variable not set.")
		os.Exit(1)
	}

	ctx := context.Background()
	loginResp, loginErr := megaportClient.AuthenticationService.Login(ctx, &LoginRequest{
		AccessKey: accessKey,
		SecretKey: secretKey,
	})
	if loginErr != nil {
		megaportClient.Logger.Error("login error", "error", loginErr.Error())
	}
	suite.NoError(loginErr)

	// Session Token is not empty
	suite.NotEmpty(loginResp.Token)
	// SessionToken is a valid guid
	suite.NotNil(IsGuid(loginResp.Token))
}