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

// The `shared` package is houses functions that are used throughout the entire Megaport Go Library. They are not meant
// for general use outside of the Library.
package shared

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
	"github.com/engi-fyi/go-credentials/credential"
	"github.com/engi-fyi/go-credentials/factory"
	"github.com/xlzd/gotp"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// MakeAPICall is a wrapper for HTTP calls against the Megaport API. It simply requires HTTP Verb, a URL, and the body
// of the request encoded as a byte array. It also sets generic headers such as content types, user agent. It also
// ensures the authentication tokesns are in the request headers.
func MakeAPICall(verb string, url string, body []byte) (*http.Response, error) {
	credFactory, credErr := factory.New(types.APPLICATION_SHORT_NAME)

	if credErr != nil {
		return nil, credErr
	}

	credentials, credErr := credential.Load(credFactory)

	if credErr != nil {
		return nil, credErr
	}

	if credentials.Section("authentication").GetAttribute("session_token") == "" {
		tokenErr := GetSessionToken(credentials)

		if tokenErr != nil {
			return nil, tokenErr
		}
	}

	fullUrl := credentials.GetAttribute("megaport_url") + url

	if fullUrl == "" {
		return nil, errors.New(mega_err.ERR_MEGAPORT_URL_NOT_SET)
	}

	client := &http.Client{}
	var request *http.Request
	var reqErr error

	if body == nil {
		request, reqErr = http.NewRequest(verb, fullUrl, nil)
	} else {
		request, reqErr = http.NewRequest(verb, fullUrl, bytes.NewBuffer(body))
	}

	if reqErr != nil {
		return nil, reqErr
	}

	request.Header.Set("X-Auth-Token", credentials.Section("authentication").GetAttribute("session_token"))
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Go-Megaport-Library/0.1")

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, resErr := client.Do(request)

	if resErr != nil {
		return nil, resErr
	} else {
		return response, nil
	}
}

// GetSessionToken connects to the Megaport API, passes the saved username, password, and a generated OTP
// (if otp key is set). The Megaport API returns a session token which is then saved in the session token file. The
// Session Token is also stored in the environment variable `MEGAPORT_SESSION_TOKEN`.
func GetSessionToken(credentials *credential.Credential) error {
	megaportUrl := credentials.GetAttribute("megaport_url")
	loginURL := megaportUrl + "/v2/login"
	data := url.Values{}
	client := &http.Client{}

	data.Add("username", credentials.Username)
	data.Add("password", credentials.Password)

	if megaportUrl == "" {
		return errors.New(mega_err.ERR_MEGAPORT_URL_NOT_SET)
	}

	if credentials.Section("otp").GetAttribute("key") != "" {

		oneTimePassword, otpErr := GenerateOneTimePassword(credentials)

		if otpErr != nil {
			return otpErr
		}

		data.Add("oneTimePassword", oneTimePassword)
	}

	loginRequest, _ := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRequest.Header.Set("Accept", "application/json")
	loginRequest.Header.Set("User-Agent", "Go-Megaport-Library/0.1")

	response, resErr := client.Do(loginRequest)

	if resErr != nil {
		return resErr
	}

	defer response.Body.Close()

	isError, compiledError := IsErrorResponse(response, &resErr, 200)

	if isError {
		return compiledError
	}

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return fileErr
	}

	authResponse := types.GenericResponse{}

	parseErr := json.Unmarshal([]byte(body), &authResponse)

	if parseErr != nil {
		return parseErr
	}

	credentials.Section("authentication").SetAttribute("session_token", authResponse.Data["session"].(string))

	return nil
}

// Generates a OTP using a Google Authenticator-compatible OTP Key. The field `one_time_password_key` must be set in
// your Megaport credentials file.
func GenerateOneTimePassword(credentials *credential.Credential) (string, error) {
	if credentials.Section("otp").GetAttribute("key") == "" {
		return "", errors.New(mega_err.ERR_NO_OTP_KEY_DEFINED)
	}

	return gotp.NewDefaultTOTP(credentials.Section("otp").GetAttribute("key")).Now(), nil
}

// IsErrorResponse returns an error report if an error response is detected from the API.
func IsErrorResponse(response *http.Response, responseErr *error, expectedReturnCode int) (bool, error) {
	if *responseErr != nil {
		return true, *responseErr
	}

	if response.StatusCode != expectedReturnCode {
		body, fileErr := ioutil.ReadAll(response.Body)

		if fileErr != nil {
			return false, fileErr
		}

		errResponse := types.ErrorResponse{}
		errParseErr := json.Unmarshal([]byte(body), &errResponse)

		if errParseErr != nil {
			errorReport := fmt.Sprintf(mega_err.ERR_PARSING_ERR_RESPONSE, response.StatusCode, errParseErr.Error(), string(body))
			return true, errors.New(errorReport)
		}

		return true, errors.New(errResponse.Message + ": " + errResponse.Data)
	}

	return false, nil
}

// PurchaseError prints out details about a failed purchase.
func PurchaseError(productID string, err error) {
	if !IsGuid(productID) {
		log.Printf("Returned product ID is empty.")
	}

	if err != nil {
		log.Printf("Error purchasing Product: %s", err)
	}
}

// GenerateRandomNumber generates a random number between an upper and lower bound.
func GenerateRandomNumber(lowerBound int, upperBound int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(upperBound) + lowerBound
}
