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
	"encoding/json"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestCreateUser(t *testing.T) {
	if os.Getenv("CREATE_USER") == "TRUE" {
		megaportUrl, exists := os.LookupEnv("MEGAPORT_URL")

		if !exists {
			megaportUrl = "https://api-staging.megaport.com/"
		}

		credentials, credErr := CreateUser("Go", "Testing", megaportUrl)
		assert.NoError(t, credErr)
		credentials, loginErr := Login(true)
		assert.NoError(t, loginErr)
		assert.True(t, shared.IsGuid(credentials.Section("authentication").GetAttribute("session_token")))
		isSetup, setupErr := setCompanyNameAndMarket(credentials.Username)
		assert.NoError(t, setupErr)
		assert.True(t, isSetup)
	}
}

// TestLogin tests the Login() func.
func TestLogin(t *testing.T) {
	credentials, loginErr := Login(true)
	assert.NoError(t, loginErr)

	if loginErr != nil {
		log.Fatalf("LoginError: %s", loginErr.Error())
	}

	// Username is an email address
	assert.NotNil(t, shared.IsEmail(credentials.Username))
	// Password is not empty
	assert.NotEmpty(t, credentials.Password)
	// SessionToken is a valid guid
	assert.NotNil(t, shared.IsGuid(credentials.Section("authentication").GetAttribute("session_token")))

	// Validate that the correct values are restored from file
	restoredCredential, restoreErr := Login(false)
	assert.NoError(t, restoreErr)
	assert.Equal(t, credentials.Username, restoredCredential.Username)
	assert.Equal(t, credentials.Password, restoredCredential.Password)
	assert.Equal(t, credentials.Section("authentication").GetAttribute("session_token"),
		restoredCredential.Section("authentication").GetAttribute("session_token"))
}

// TestLogin tests the Login() func.
func TestLogout(t *testing.T) {
	logoutErr := Logout()
	assert.Nil(t, logoutErr)
}

func setCompanyNameAndMarket(contactEmail string) (bool, error) {
	company := types.CompanyEnablement{TradingName: "Go Testing Company"}
	market := types.Market{
		Currency:               "AUD",
		Language:               "en",
		CompanyLegalIdentifier: "ABN987654",
		CompanyLegalName:       "Go Testing Company",
		BillingContactName:     "Go Testing",
		BillingContactPhone:    "0730000000",
		BillingContactEmail:    contactEmail,
		AddressLine1:           "Level 3, 825 Ann St,  QLD 4006",
		City:                   "Fortitude Valley",
		State:                  "QLD",
		Postcode:               "4006",
		Country:                "AU",
		FirstPartyID:           808,
	}

	companyJSON, companyMarshalErr := json.Marshal(company)

	if companyMarshalErr != nil {
		return false, companyMarshalErr
	}

	companyResponse, companyErr := shared.MakeAPICall("POST", "/v2/social/company", companyJSON)
	defer companyResponse.Body.Close()

	isCompanyError, parsedCompanyErr := shared.IsErrorResponse(companyResponse, &companyErr, 200)

	if isCompanyError {
		return false, parsedCompanyErr
	}

	marketJSON, marketMarshalErr := json.Marshal(market)

	if marketMarshalErr != nil {
		return false, marketMarshalErr
	}

	marketResponse, marketErr := shared.MakeAPICall("POST", "/v2/market/", marketJSON)
	defer marketResponse.Body.Close()

	isMarketError, parsedMarketErr := shared.IsErrorResponse(marketResponse, &marketErr, 201)

	if isMarketError {
		return false, parsedMarketErr
	}

	return true, nil
}
