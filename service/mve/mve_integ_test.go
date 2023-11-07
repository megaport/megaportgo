//go:build integration
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

package mve

import (
	"os"
	"testing"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/service/authentication"
	"github.com/megaport/megaportgo/service/location"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/assert"
)

const (
	TEST_MVE_TEST_LOCATION_MARKET = "AU"
	MEGAPORTURL                   = "https://api-staging.megaport.com/"
)

var logger *config.DefaultLogger
var cfg config.Config

func TestMain(m *testing.M) {
	logger = config.NewDefaultLogger()
	logger.SetLevel(config.DebugLevel)

	clientID := os.Getenv("MEGAPORT_ACCESS_KEY")
	clientSecret := os.Getenv("MEGAPORT_SECRET_KEY")
	logLevel := os.Getenv("LOG_LEVEL")

	if logLevel != "" {
		logger.SetLevel(config.StringToLogLevel(logLevel))
	}

	if clientID == "" {
		logger.Error("MEGAPORT_ACCESS_KEY environment variable not set.")
		os.Exit(1)
	}

	if clientSecret == "" {
		logger.Error("MEGAPORT_SECRET_KEY environment variable not set.")
		os.Exit(1)
	}

	cfg = config.Config{
		Client:   config.NewHttpClient(),
		Log:      logger,
		Endpoint: MEGAPORTURL,
	}

	auth := authentication.New(&cfg)
	if _, err := auth.LoginOauth(clientID, clientSecret); err != nil {
		logger.Errorf("LoginError: %s", err)
	}

	os.Exit(m.Run())
}

func readSSHPubKey() string {
	key, err := os.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa.pub")
	if err != nil {
		panic(err)
	}
	return string(key)
}

func TestC8KVAutoLifecycle(t *testing.T) {
	assert := assert.New(t)
	mve := New(&cfg)
	location := location.New(&cfg)

	logger.Debug("Buying MVE")
	testLocation := location.GetRandom(TEST_MVE_TEST_LOCATION_MARKET)
	logger.Debugf("Test location determined, Location: %s", testLocation.Name)

	mveConfig := map[string]interface{}{
		"vendor":       "Cisco",
		"productSize":  "SMALL",
		"imageId":      int(42),
		"sshPublicKey": readSSHPubKey(),
	}

	mveUid, err := mve.BuyMVE(testLocation.ID, "MVE Test", 12, mveConfig, nil)

	if !assert.NoError(err) && assert.False(shared.IsGuid(mveUid)) {
		mve.Config.PurchaseError(mveUid, err)
	}

	logger.Debugf("MVE Purchased: %s", mveUid)
	logger.Debug("Wating for MVE to be provisioned")
	mve.WaitForMVEProvisioning(mveUid)

	// Testing MVE Delete
	logger.Info("Deleting MVE now.")
	hardDeleteStatus, deleteErr := mve.DeleteMVE(mveUid)
	assert.True(hardDeleteStatus)
	assert.NoError(deleteErr)

	mveDeleteInfo, _ := mve.GetMVEDetails(mveUid)
	assert.EqualValues(types.STATUS_DECOMMISSIONED, mveDeleteInfo.ProvisioningStatus)
	logger.Debugf("Status is: %s", mveDeleteInfo.ProvisioningStatus)
}
