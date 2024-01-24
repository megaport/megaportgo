package megaport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"time"

	"github.com/megaport/megaportgo/types"
)

type AuthenticationService interface {
	LoginOauth(ctx context.Context, accessKey, secretKey string) (string, error)
}

type AuthenticationServiceOp struct {
	*Client

	bearerToken string
	tokenExpiry time.Time
}

func NewAuthenticationServiceOp(c *Client) *AuthenticationServiceOp {
	return &AuthenticationServiceOp{
		Client: c,
	}
}

// LoginOauth performs an OAuth-style logi using an API key and API
// secret key. It returns the bearer token or an error if the login
// was unsuccessful.
func (svc *AuthenticationServiceOp) LoginOauth(ctx context.Context, accessKey, secretKey string) (string, error) {
	svc.Logger.Debug("creating session", "access_key", accessKey)

	// Shortcut if we've already authenticated.
	if time.Now().Before(svc.tokenExpiry) {
		return svc.bearerToken, nil
	}

	// Encode the client ID and client secret to create Basic Authentication
	authHeader := base64.StdEncoding.EncodeToString([]byte(accessKey + ":" + secretKey))

	// Set the URL for the token endpoint
	var tokenURL string

	if svc.Client.BaseURL.Host == "api.megaport.com" {
		tokenURL = "https://auth-m2m.megaport.com/oauth2/token"
	} else if svc.Client.BaseURL.Host == "api-staging.megaport.com" {
		tokenURL = "https://oauth-m2m-staging.auth.ap-southeast-2.amazoncognito.com/oauth2/token"
	} else if svc.Client.BaseURL.Host == "api-uat.megaport.com" {
		tokenURL = "https://oauth-m2m-uat.auth.ap-southeast-2.amazoncognito.com/oauth2/token"
	} else if svc.Client.BaseURL.Host == "api-uat2.megaport.com" {
		tokenURL = "https://oauth-m2m-uat2.auth.ap-southeast-2.amazoncognito.com/oauth2/token"
	}

	// Create form data for the request body
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Create an HTTP request
	req, err := svc.Client.NewRequest(ctx, "POST", tokenURL, nil)
	if err != nil {
		return "", err
	}

	req.URL.RawQuery = data.Encode()

	// Set the request headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+authHeader)

	svc.Logger.Debug(req.Header.Get("Authorization"))
	svc.Logger.Debug(req.Header.Get("Content-Type"))

	// Create an HTTP client and send the request
	svc.Logger.Debug("login request", "token_url", tokenURL)
	resp, resErr := svc.Client.Do(ctx, req, nil)
	if resErr != nil {
		return "", resErr
	}

	defer resp.Body.Close()

	// Read the response body
	body, fileErr := io.ReadAll(resp.Body)
	if fileErr != nil {
		return "", fileErr
	}

	// Parse the response JSON to extract the access token and expiration time
	authResponse := types.AccessTokenResponse{}
	if parseErr := json.Unmarshal(body, &authResponse); parseErr != nil {
		return "", parseErr
	}

	if authResponse.Error != "" {
		return "", errors.New("authentication error: " + authResponse.Error)
	}

	// Calculate the token expiration time
	svc.tokenExpiry = time.Now().Add(time.Duration(authResponse.ExpiresIn) * time.Second)

	// Store the access token
	svc.bearerToken = authResponse.AccessToken
	svc.SessionToken = authResponse.AccessToken

	svc.Logger.Debug("session established")

	return svc.bearerToken, nil
}
