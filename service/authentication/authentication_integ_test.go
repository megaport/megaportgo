// +build integration

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

package authentication

import (
	"fmt"
	"os"
	"testing"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/shared"
	"github.com/stretchr/testify/assert"
)

var username string
var password string
var otp string

var logger *config.DefaultLogger
var cfg config.Config

const (
	MEGAPORTURL = "https://api-staging.megaport.com/"
)

func TestMain(m *testing.M) {
	logger = config.NewDefaultLogger()
	logger.SetLevel(config.DebugLevel)

	username = os.Getenv("MEGAPORT_USERNAME")
	password = os.Getenv("MEGAPORT_PASSWORD")
	otp = os.Getenv("MEGAPORT_MFA_OTP_KEY")
	logLevel := os.Getenv("LOG_LEVEL")

	fmt.Println(logLevel)
	if logLevel != "" {
		logger.SetLevel(config.StringToLogLevel(logLevel))
	}

	if username == "" {
		logger.Error("MEGAPORT_USERNAME environment variable not set.")
		os.Exit(1)
	}

	if password == "" {
		logger.Error("MEGAPORT_PASSWORD environment variable not set.")
		os.Exit(1)
	}

	cfg = config.Config{
		Log:      logger,
		Endpoint: MEGAPORTURL,
	}

	os.Exit(m.Run())
}

// TestLogin tests that the package can login and get a session token from the API
func TestLogin(t *testing.T) {

	auth := New(&cfg, username, password, otp)
	token, loginErr := auth.Login()

	assert.NoError(t, loginErr)

	if loginErr != nil {
		logger.Errorf("LoginError: %s", loginErr.Error())
	}

	// Session Token is not empty
	assert.NotEmpty(t, token)
	// SessionToken is a valid guid
	assert.NotNil(t, shared.IsGuid(token))
}
