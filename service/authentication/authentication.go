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
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
	"github.com/xlzd/gotp"
)

type Authentication struct {
	*config.Config

	username        string
	password        string
	oneTimePassword string
}

func New(cfg *config.Config, username string, password string, oneTimePassword string) *Authentication {
	return &Authentication{
		Config:          cfg,
		username:        username,
		password:        password,
		oneTimePassword: oneTimePassword,
	}
}

// Login is a wrapper around `GetSessionToken` which loads a username and password from file first before initiating
// the retrieval of a Session Token.
func (auth *Authentication) Login() (string, error) {
	auth.Log.Debugln("Creating Session for:", auth.username)
	token, err := auth.getSessionToken()

	if err != nil {
		auth.Log.Debugln("Unable to get Session token: ", err)
		return "", err
	}

	auth.Log.Debugln("Session Established")
	return token, nil
}

func (a *Authentication) getSessionToken() (string, error) {
	loginURL := a.Config.Endpoint + "/v2/login"
	data := url.Values{}
	client := &http.Client{}

	data.Add("username", a.username)
	data.Add("password", a.password)

	if a.oneTimePassword != "" {
		oneTimePassword, otpErr := generateOneTimePassword(a.oneTimePassword)

		if otpErr != nil {
			return "", otpErr
		}

		data.Add("oneTimePassword", oneTimePassword)
	}

	loginRequest, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRequest.Header.Set("Accept", "application/json")
	loginRequest.Header.Set("User-Agent", "Go-Megaport-Library/0.1")

	response, resErr := client.Do(loginRequest)

	if resErr != nil {
		return "", resErr
	}

	defer response.Body.Close()

	isError, compiledError := a.Config.IsErrorResponse(response, &resErr, 200)

	if isError {
		return "", compiledError
	}

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return "", fileErr
	}

	authResponse := types.GenericResponse{}

	parseErr := json.Unmarshal([]byte(body), &authResponse)

	if parseErr != nil {
		return "", parseErr
	}

	return authResponse.Data["session"].(string), nil
}

// generateOneTimePassword Generates a OTP using a Google Authenticator-compatible OTP Key. The field `one_time_password_key` must be set in
// your Megaport credentials file.
func generateOneTimePassword(otp string) (string, error) {
	if otp == "" {
		return "", errors.New(mega_err.ERR_NO_OTP_KEY_DEFINED)
	}

	return gotp.NewDefaultTOTP(otp).Now(), nil
}
