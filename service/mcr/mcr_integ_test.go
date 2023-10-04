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

package mcr

import (
	"fmt"
	"os"
	"testing"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/service/authentication"
	"github.com/megaport/megaportgo/service/location"
	"github.com/megaport/megaportgo/service/vxc"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/assert"
)

const (
	TEST_MCR_TEST_LOCATION_MARKET = "AU"
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
		Log:      logger,
		Endpoint: MEGAPORTURL,
	}

	auth := authentication.New(&cfg, clientID, clientSecret)
	token, loginErr := auth.Login()

	if loginErr != nil {
		logger.Errorf("LoginError: %s", loginErr.Error())
	}

	cfg.SessionToken = token
	os.Exit(m.Run())
}

func TestMCRLifecycle(t *testing.T) {
	assert := assert.New(t)
	mcr := New(&cfg)
	location := location.New(&cfg)

	logger.Debug("Buying MCR Port.")
	testLocation := location.GetRandom(TEST_MCR_TEST_LOCATION_MARKET)

	logger.Debugf("Test location determined, Location: %s", testLocation.Name)
	mcrId, portErr := mcr.BuyMCR(testLocation.ID, "Buy MCR", 1000, 0)

	if !assert.NoError(portErr) && assert.False(shared.IsGuid(mcrId)) {
		mcr.Config.PurchaseError(mcrId, portErr)
	}

	logger.Debugf("MCR Purchased: %s", mcrId)
	logger.Debug("Wating for MCR to be provisioned")
	mcr.WaitForMcrProvisioning(mcrId)

	// Testing MCR Modify
	mcrInfo, _ := mcr.GetMCRDetails(mcrId)

	logger.Debug("Modifying MCR.")
	newMCRName := fmt.Sprintf("Buy MCR [Modified]")

	_, modifyErr := mcr.ModifyMCR(mcrId, newMCRName, "", mcrInfo.MarketplaceVisibility)
	assert.NoError(modifyErr)

	mcrInfo, _ = mcr.GetMCRDetails(mcrId)
	assert.EqualValues(newMCRName, mcrInfo.Name)

	// Testing MCR Cancel
	logger.Info("Scheduling MCR for deletion (30 days).")

	// This is a soft Delete
	softDeleteStatus, deleteErr := mcr.DeleteMCR(mcrId, false)
	assert.True(softDeleteStatus)
	assert.NoError(deleteErr)

	mcrCancelInfo, _ := mcr.GetMCRDetails(mcrId)
	assert.EqualValues(types.STATUS_CANCELLED, mcrCancelInfo.ProvisioningStatus)
	logger.Debugf("Status is: %s", mcrCancelInfo.ProvisioningStatus)

	restoreStatus, restoreErr := mcr.RestoreMCR(mcrId)
	assert.True(restoreStatus)
	assert.NoError(restoreErr)

	// Testing MCR Delete
	logger.Info("Deleting MCR now.")

	// This is a Hard Delete
	hardDeleteStatus, deleteErr := mcr.DeleteMCR(mcrId, true)
	assert.True(hardDeleteStatus)
	assert.NoError(deleteErr)

	mcrDeleteInfo, _ := mcr.GetMCRDetails(mcrId)
	assert.EqualValues(types.STATUS_DECOMMISSIONED, mcrDeleteInfo.ProvisioningStatus)
	logger.Debugf("Status is: %s", mcrDeleteInfo.ProvisioningStatus)
}

func TestMCRConnectionAdd(t *testing.T) {
	assert := assert.New(t)

	mcr := New(&cfg)
	location := location.New(&cfg)
	vxc := vxc.New(&cfg)

	testLocation := location.GetRandom(TEST_MCR_TEST_LOCATION_MARKET)

	logger.Infof("Test location determined, Location: %s", testLocation.Name)
	logger.Debug("Buying MCR")

	mcrId, mcrErr := mcr.BuyMCR(testLocation.ID, "MCR and AWS Interconnectivity", 1000, 0)

	logger.Infof("MCR Purchased: %s", mcrId)

	if assert.NoError(mcrErr) && assert.True(shared.IsGuid(mcrId)) {
		mcr.WaitForMcrProvisioning(mcrId)

		logger.Info("Buying A")
		logger.Info("Buying AWS VIF Connection (B End).")
		vifOneId, vifOneErr := vxc.BuyAWSVXC(
			mcrId,
			"MCR and AWS Connection 1",
			500,
			types.VXCOrderAEndConfiguration{
				VLAN: shared.GenerateRandomVLAN(),
			},
			types.AWSVXCOrderBEndConfiguration{
				ProductUID: "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86",
				PartnerConfig: types.AWSVXCOrderBEndPartnerConfig{
					ConnectType:  types.CONNECT_TYPE_AWS_VIF,
					Type:         "private",
					ASN:          65105,
					AmazonASN:    65106,
					OwnerAccount: "684021030471",
					AuthKey:      "notarealauthkey",
					Prefixes:     "10.0.1.0/24",
				},
			},
		)

		vifTwoId, vifTwoErr := vxc.BuyAWSVXC(
			mcrId,
			"MCR and AWS Connection 2",
			500,
			types.VXCOrderAEndConfiguration{
				VLAN: shared.GenerateRandomVLAN(),
			},
			types.AWSVXCOrderBEndConfiguration{
				ProductUID: "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86",
				PartnerConfig: types.AWSVXCOrderBEndPartnerConfig{
					ConnectType:  types.CONNECT_TYPE_AWS_VIF,
					Type:         "private",
					ASN:          65105,
					AmazonASN:    65106,
					OwnerAccount: "684021030471",
					AuthKey:      "notarealauthkey",
					Prefixes:     "10.0.2.0/24",
				},
			},
		)

		logger.Infof("AWS VIF Connection 1: %s", vifOneId)
		logger.Infof("AWS VIF Connection 2: %s", vifTwoId)

		if assert.NoError(vifOneErr) && assert.NoError(vifTwoErr) && assert.True(shared.IsGuid(vifOneId) && assert.True(shared.IsGuid(vifTwoId))) {
			vxc.WaitForVXCProvisioning(vifOneId)
			vxc.WaitForVXCProvisioning(vifTwoId)

			vifOneDeleteStatus, vifOneDeleteErr := vxc.DeleteVXC(vifOneId, true)
			assert.NoError(vifOneDeleteErr)
			assert.True(vifOneDeleteStatus)

			vifTwoDeleteStatus, vifTwoDeleteErr := vxc.DeleteVXC(vifTwoId, true)
			assert.NoError(vifTwoDeleteErr)
			assert.True(vifTwoDeleteStatus)

			mcrDeleteStatus, mcrDeleteErr := mcr.DeleteMCR(mcrId, true)
			assert.NoError(mcrDeleteErr)
			assert.True(mcrDeleteStatus)
		} else {
			mcr.Config.PurchaseError(vifOneId, vifOneErr)
			mcr.Config.PurchaseError(vifTwoId, vifTwoErr)
		}
	} else {
		mcr.Config.PurchaseError(mcrId, mcrErr)
	}
}

func TestPortSpeedValidation(t *testing.T) {
	assert := assert.New(t)
	mcr := New(&cfg)
	location := location.New(&cfg)

	testLocation, _ := location.GetLocationByName("Global Switch Sydney")
	_, buyErr := mcr.BuyMCR(testLocation.ID, "Test MCR", 500, 0)
	assert.EqualError(buyErr, mega_err.ERR_MCR_INVALID_PORT_SPEED)
}

func TestCreatePrefixFilterList(t *testing.T) {
	assert := assert.New(t)
	mcr := New(&cfg)
	location := location.New(&cfg)

	logger.Info("Buying MCR Port.")
	testLocation := location.GetRandom(TEST_MCR_TEST_LOCATION_MARKET)

	logger.Infof("Test location determined, Location: %s", testLocation.Name)
	mcrId, portErr := mcr.BuyMCR(testLocation.ID, "Buy MCR", 1000, 0)

	if !assert.NoError(portErr) && assert.False(shared.IsGuid(mcrId)) {
		mcr.Config.PurchaseError(mcrId, portErr)
	}

	logger.Infof("MCR Purchased: %s", mcrId)
	logger.Info("Waiting for MCR to be provisioned")

	mcr.WaitForMcrProvisioning(mcrId)

	logger.Info("Creating prefix filter list")

	prefixFilterEntries := []types.MCRPrefixListEntry{
		{
			Action: "permit",
			Prefix: "10.0.1.0/24",
			Ge:     24,
			Le:     24,
		},
		{
			Action: "deny",
			Prefix: "10.0.2.0/24",
			Ge:     24,
			Le:     24,
		},
	}

	validatedPrefixFilterList := types.MCRPrefixFilterList{
		Description:   "Test Prefix Filter List",
		AddressFamily: "IPv4",
		Entries:       prefixFilterEntries,
	}

	_, prefixErr := mcr.CreatePrefixFilterList(mcrId, validatedPrefixFilterList)

	if prefixErr != nil {
		logger.Infof("%s", prefixErr.Error())
	}

	logger.Info("Deleting MCR now.")
	hardDeleteStatus, deleteErr := mcr.DeleteMCR(mcrId, true)
	assert.True(hardDeleteStatus)
	assert.NoError(deleteErr)

	mcrDeleteInfo, _ := mcr.GetMCRDetails(mcrId)
	assert.EqualValues(types.STATUS_DECOMMISSIONED, mcrDeleteInfo.ProvisioningStatus)
	logger.Debugf("Status is: %s", mcrDeleteInfo.ProvisioningStatus)
}
