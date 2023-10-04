// Copyright 2020 Megaport Pty Ltd
//
// Licensed under the Mozilla Public License, Version 2.0 (the
// "License"); you may not use this file except in compliance with
// the License. You may obtain a copy of the License at
//
//       https://mozilla.org/MPL/2.0/
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// The `authentication` package is used to manage login/logout of Megaport services and manage session details.
//
// This package is responsible for manipulating `types.Credential` objects, in line with the Megaport session
// lifecycle. This package also support TOTP authentication.
//
// For API Docs about authentication, please see: https://dev.megaport.com/#security.
package authentication

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
	"github.com/xlzd/gotp"
)

type Authentication struct {
	*config.Config

	bearerToken string
	tokenExpiry time.Time
}

func New(cfg *config.Config) *Authentication {
	return &Authentication{
		Config: cfg,
	}
}

// LoginOauth performs an OAuth-style logi using an API key and API
// secret key. It returns the bearer token or an error if the login
// was unsuccessful.
func (auth *Authentication) LoginOauth(accessKey, secretKey string) (string, error) {
	auth.Log.Debugln("Creating Session for:", accessKey)

	// Shortcut if we've already authenticated.
	if time.Now().Before(auth.tokenExpiry) {
		return auth.bearerToken, nil
	}

	// Encode the client ID and client secret to create Basic Authentication
	authHeader := base64.StdEncoding.EncodeToString([]byte(accessKey + ":" + secretKey))

	// Set the URL for the token endpoint
	var tokenURL string
	if auth.Config.Endpoint == "https://api.megaport.com/" {
		tokenURL = "https://auth-m2m.megaport.com/oauth2/token"
	} else if auth.Config.Endpoint == "https://api-staging.megaport.com/" {
		tokenURL = "https://oauth-m2m-staging.auth.ap-southeast-2.amazoncognito.com/oauth2/token"
	} else if auth.Config.Endpoint == "https://api-uat.megaport.com/" {
		tokenURL = "https://oauth-m2m-uat.auth.ap-southeast-2.amazoncognito.com/oauth2/token"
	} else if auth.Config.Endpoint == "https://api-uat2.megaport.com/" {
		tokenURL = "https://oauth-m2m-uat2.auth.ap-southeast-2.amazoncognito.com/oauth2/token"
	}

	// Create form data for the request body
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Create an HTTP request
	req, _ := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))

	// Set the request headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", "Basic "+authHeader)

	// Create an HTTP client and send the request
	auth.Log.Debugln("Login request to:", tokenURL)
	resp, resErr := auth.Client.Do(req)
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
	auth.tokenExpiry = time.Now().Add(time.Duration(authResponse.ExpiresIn) * time.Second)

	// Store the access token
	auth.bearerToken = authResponse.AccessToken
	auth.SessionToken = authResponse.AccessToken

	auth.Log.Debugln("session established")
	return auth.bearerToken, nil
}

// LoginUsername performs a login against the portal endpoint. This
// method is being deprecated in favour of the OAuth method. It returns
// the bearer token or an error if authentication fails.
func (auth *Authentication) LoginUsername(username, password, oneTimePassword string) (string, error) {
	auth.Log.Debugln("Creating Session for:", username)

	loginURL := auth.Config.Endpoint + "/v2/login"
	data := url.Values{}

	data.Add("username", username)
	data.Add("password", password)

	if oneTimePassword != "" {
		otpVal, otpErr := generateOneTimePassword(oneTimePassword)

		if otpErr != nil {
			return "", otpErr
		}

		data.Add("oneTimePassword", otpVal)
	}

	loginRequest, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRequest.Header.Set("Accept", "application/json")
	loginRequest.Header.Set("User-Agent", "Go-Megaport-Library/0.1")

	response, resErr := auth.Client.Do(loginRequest)

	if resErr != nil {
		return "", resErr
	}
	defer response.Body.Close()

	isError, compiledError := auth.Config.IsErrorResponse(response, &resErr, 200)
	if isError {
		return "", compiledError
	}

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return "", fileErr
	}

	authResponse := types.LoginResponse{}
	parseErr := json.Unmarshal([]byte(body), &authResponse)
	if parseErr != nil {
		return "", parseErr
	}
	oauth := authResponse.Data.OAuthToken

	// Calculate the token expiration time
	auth.tokenExpiry = time.Now().Add(time.Duration(oauth.ExpiresIn) * time.Second)

	// Store the access token
	auth.bearerToken = oauth.AccessToken
	auth.SessionToken = oauth.AccessToken

	auth.Log.Debugln("session created")
	return auth.bearerToken, nil
}

// generateOneTimePassword Generates a OTP using a Google Authenticator-compatible OTP Key. The field `one_time_password_key` must be set in
// your Megaport credentials file.
func generateOneTimePassword(otp string) (string, error) {
	if otp == "" {
		return "", errors.New(mega_err.ERR_NO_OTP_KEY_DEFINED)
	}

	return gotp.NewDefaultTOTP(otp).Now(), nil
}
