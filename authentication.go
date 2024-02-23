package megaport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/url"
	"time"
)

// AuthenticationService is an interface for interfacing with the Authentication endpoints of the Megaport API.
type AuthenticationService interface {
	// Login performs an OAuth-style login using an API key and API secret key. It returns the bearer token or an error if the login was unsuccessful.
	Login(ctx context.Context, req *LoginRequest) (*LoginOauthResponse, error)
}

// AuthenticationServiceOp handles communication with Authentication methods of the Megaport API.
type AuthenticationServiceOp struct {
	*Client
}

// NewAuthenticationService creates a new instance of the Authentication Service.
func NewAuthenticationService(c *Client) *AuthenticationServiceOp {
	return &AuthenticationServiceOp{
		Client: c,
	}
}

// LoginRequest is the request structure for the Login method containing the API Access Key and API Secret Key.
type LoginRequest struct {
	AccessKey string
	SecretKey string
}

// LoginOauthResponse is the response structure for the Login method containing the bearer token.
type LoginOauthResponse struct {
	Token string
}

// AccessTokenResponse is the response structure for the Login method containing the access token and expiration time.
type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error"`
}

// LoginResponse is the response structure for the Login method containing the access token and expiration time, as well as the available permissions.
type LoginResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    struct {
		Permissions map[string][]string `json:"permissions"`
		OAuthToken  struct {
			AccessToken string `json:"accessToken"`
			ExpiresIn   int
		}
	}
}

// Login performs an OAuth-style login using an API key and API secret key. It returns the bearer token or an error if the login was unsuccessful.
func (svc *AuthenticationServiceOp) Login(ctx context.Context, req *LoginRequest) (*LoginOauthResponse, error) {
	svc.Logger.DebugContext(ctx, "creating session", slog.String("access_key", req.AccessKey))

	// Shortcut if we've already authenticated.
	if time.Now().Before(svc.TokenExpiry) {
		return &LoginOauthResponse{
			Token: svc.SessionToken,
		}, nil
	}

	// Encode the client ID and client secret to create Basic Authentication
	authHeader := base64.StdEncoding.EncodeToString([]byte(req.AccessKey + ":" + req.SecretKey))

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
	clientReq, err := svc.Client.NewRequest(ctx, "POST", tokenURL, nil)
	if err != nil {
		return nil, err
	}

	clientReq.URL.RawQuery = data.Encode()

	// Set the request headers
	clientReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientReq.Header.Set("Authorization", "Basic "+authHeader)

	// Create an HTTP client and send the request
	svc.Logger.DebugContext(ctx, "login request", slog.String("token_url", tokenURL), slog.String("authorization_header", clientReq.Header.Get("Authorization")), slog.String("content_type", clientReq.Header.Get("Content_Type")))
	resp, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return nil, resErr
	}

	defer resp.Body.Close()

	// Read the response body
	body, fileErr := io.ReadAll(resp.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	// Parse the response JSON to extract the access token and expiration time
	authResponse := AccessTokenResponse{}
	if parseErr := json.Unmarshal(body, &authResponse); parseErr != nil {
		return nil, parseErr
	}

	if authResponse.Error != "" {
		return nil, errors.New("authentication error: " + authResponse.Error)
	}

	// Store the access token
	svc.TokenExpiry = time.Now().Add(time.Duration(authResponse.ExpiresIn) * time.Second)
	// Calculate the token expiration time
	svc.SessionToken = authResponse.AccessToken

	svc.Logger.DebugContext(ctx, "session established")

	return &LoginOauthResponse{
		Token: authResponse.AccessToken,
	}, nil
}
