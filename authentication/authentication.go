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
	"errors"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/engi-fyi/go-credentials/credential"
	"github.com/engi-fyi/go-credentials/factory"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// CreateUser creates a user for use within the Megaport Ecosystem. Is is recommended you do not use this function for
// anything except testing, and only in the staging environment.
func CreateUser(firstName string, lastName string, megaportUrl string) (*credential.Credential, error) {
	credFactory, factErr := factory.New(types.APPLICATION_SHORT_NAME)
	postfix := generateRandomString(10)
	createUserUrl := megaportUrl + "/v2/social/registration"
	username := "golib+" + postfix + "@sink.megaport.com"
	password := generateRandomStringWithCharset(20, CHARSET+"!@+=")

	if factErr != nil {
		return nil, factErr
	}

	myCredentials, credErr := credential.New(credFactory, username, password)
	myCredentials.Section("user_details").SetAttribute("first_name", firstName)
	myCredentials.Section("user_details").SetAttribute("second_name", lastName)
	myCredentials.SetAttribute("megaport_url", megaportUrl)

	if credErr != nil {
		return nil, credErr
	}

	data := url.Values{}
	client := &http.Client{}

	data.Add("firstName", firstName)
	data.Add("lastName", lastName)
	data.Add("email", myCredentials.Username)
	data.Add("password", myCredentials.Password)

	loginRequest, _ := http.NewRequest("POST", createUserUrl, strings.NewReader(data.Encode()))
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRequest.Header.Set("Accept", "application/json")
	loginRequest.Header.Set("User-Agent", "Go-Megaport-Library/0.1")

	response, resErr := client.Do(loginRequest)

	if resErr != nil {
		return nil, resErr
	}

	defer response.Body.Close()

	isError, compiledError := shared.IsErrorResponse(response, &resErr, 201)

	if isError {
		return nil, compiledError
	}

	saveErr := myCredentials.Save()

	if saveErr != nil {
		return nil, saveErr
	}

	return myCredentials, nil
}

// Login is a wrapper around `GetSessionToken` which loads a username and password from file first before initiating
// the retrieval of a Session Token.
func Login(forceNew bool) (*credential.Credential, error) {
	credFactory, factErr := factory.New(types.APPLICATION_SHORT_NAME)

	if factErr != nil {
		return nil, factErr
	}
	credentials, credErr := credential.Load(credFactory)

	if credErr != nil {
		return nil, credErr
	}

	if credentials.Section("authentication").GetAttribute("session_token") == "" || forceNew {
		sessionErr := shared.GetSessionToken(credentials)

		if sessionErr != nil {
			return nil, sessionErr
		}
	}

	credentials.Save()
	return credentials, nil
}

// Logout deletes the stored Session Token and clears the `MEGAPORT_SESSION_TOKEN` environment variable.
func Logout() error {
	credFactory, factoryErr := factory.New(types.APPLICATION_SHORT_NAME)

	if factoryErr != nil {
		return factoryErr
	}

	credentials, credErr := credential.Load(credFactory)

	if credErr != nil {
		return credErr
	}

	credentials.Section("authentication").DeleteAttribute("session_token")

	if credentials.Section("authentication").GetAttribute("session_token") != "" {
		return errors.New(mega_err.ERR_SESSION_TOKEN_STILL_EXIST)
	}
	saveErr := credentials.Save()

	if saveErr != nil {
		return saveErr
	}

	return nil
}

func generateRandomStringWithCharset(length int, charset string) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateRandomString(length int) string {
	return generateRandomStringWithCharset(length, CHARSET)
}
