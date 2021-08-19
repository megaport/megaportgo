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

package vxc

// TODO: Add in tests for port-port using Service Keys.

import (
	"os"
	"testing"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/service/authentication"
	"github.com/megaport/megaportgo/service/location"
	"github.com/megaport/megaportgo/service/port"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/assert"
)

const (
	MEGAPORTURL     = "https://api-staging.megaport.com/"
	TEST_LOCATION_A = "Global Switch Sydney"
	TEST_LOCATION_B = "Equinix SY3"
	TEST_LOCATION_C = "42464fbb-4d38-4f82-9061-294e1d84ed9f"
)

var logger *config.DefaultLogger
var cfg config.Config

func TestMain(m *testing.M) {
	logger = config.NewDefaultLogger()
	logger.SetLevel(config.DebugLevel)

	username := os.Getenv("MEGAPORT_USERNAME")
	password := os.Getenv("MEGAPORT_PASSWORD")
	otp := os.Getenv("MEGAPORT_MFA_OTP_KEY")
	logLevel := os.Getenv("LOG_LEVEL")

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

	auth := authentication.New(&cfg, username, password, otp)
	token, loginErr := auth.Login()

	if loginErr != nil {
		logger.Errorf("LoginError: %s", loginErr.Error())
	}

	cfg.SessionToken = token
	os.Exit(m.Run())
}

func TestVXCBuy(t *testing.T) {
	assert := assert.New(t)
	vxc := New(&cfg)
	loc := location.New(&cfg)
	port := port.New(&cfg)

	fuzzySearch, locationErr := loc.GetLocationByNameFuzzy(TEST_LOCATION_A)
	testLocation := fuzzySearch[0]

	assert.NoError(locationErr)

	logger.Info("Buying Port (A End).")
	aEnd, aErr := port.BuySinglePort("VXC Port A", 1, 1000, testLocation.ID, "AU", true)
	logger.Infof("Port Purchased: %s", aEnd)

	if !assert.NoError(aErr) && !assert.True(shared.IsGuid(aEnd)) {
		cfg.PurchaseError(aEnd, aErr)
		t.FailNow()
	}

	logger.Info("Buying Port (B End).")
	bEnd, bErr := port.BuySinglePort("VXC Port B", 1, 1000, testLocation.ID, "AU", true)
	logger.Infof("Port Purchased: %s", bEnd)

	if !assert.NoError(bErr) && !assert.True(shared.IsGuid(bEnd)) {
		cfg.PurchaseError(bEnd, bErr)
		t.FailNow()
	}

	port.WaitForPortProvisioning(aEnd)
	port.WaitForPortProvisioning(bEnd)

	logger.Info("Buying VXC.")

	vxcId, vxcErr := vxc.BuyVXC(aEnd, bEnd, "Test VXC", 500, shared.GenerateRandomVLAN(), shared.GenerateRandomVLAN())
	logger.Infof("VXC Purchased: %s", vxcId)

	if !assert.NoError(vxcErr) && !assert.True(shared.IsGuid(vxcId)) {
		cfg.PurchaseError(vxcId, vxcErr)
		t.FailNow()
	}

	vxc.WaitForVXCProvisioning(vxcId)

	newAVLAN := shared.GenerateRandomVLAN()
	newBVLAN := shared.GenerateRandomVLAN()
	updateStatus, updateErr := vxc.UpdateVXC(vxcId, "VXC Update Name Test", 1000, newAVLAN, newBVLAN)
	assert.True(updateStatus)
	assert.NoError(updateErr)
	vxc.WaitForVXCUpdated(vxcId, "VXC Update Name Test", 1000, newAVLAN, newBVLAN)
	vxcInfo, _ := vxc.GetVXCDetails(vxcId)
	assert.EqualValues("VXC Update Name Test", vxcInfo.Name)
	assert.EqualValues(1000, vxcInfo.RateLimit)
	assert.EqualValues(newAVLAN, vxcInfo.AEndConfiguration.VLAN)
	assert.EqualValues(newBVLAN, vxcInfo.BEndConfiguration.VLAN)

	vxcDeleteStatus, vxcDeleteErr := vxc.DeleteVXC(vxcId, true)
	assert.NoError(vxcDeleteErr)
	assert.True(vxcDeleteStatus, nil)

	aDeleteStatus, aDeleteErr := port.DeletePort(aEnd, true)
	assert.NoError(aDeleteErr)
	assert.True(aDeleteStatus)

	bDeleteStatus, bDeleteErr := port.DeletePort(bEnd, true)
	assert.NoError(bDeleteErr)
	assert.True(bDeleteStatus)
}

func TestAWSConnectionBuy(t *testing.T) {
	vxc := New(&cfg)
	loc := location.New(&cfg)
	port := port.New(&cfg)

	testLocation, _ := loc.GetLocationByName(TEST_LOCATION_B)

	logger.Info("Buying AWS VIF Port (A End).")
	portId, portErr := port.BuySinglePort("AWS VIF Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	logger.Infof("Port Purchased: %s", portId)

	if !assert.NoError(t, portErr) && !assert.True(t, shared.IsGuid(portId)) {
		cfg.PurchaseError(portId, portErr)
		t.FailNow()
	}

	port.WaitForPortProvisioning(portId)

	logger.Info("Buying AWS VIF Connection (B End).")
	hostedVifId, hostedVifErr := vxc.BuyAWSHostedVIF(portId, "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86", "Hosted AWS VIF Test Connection", 500, shared.GenerateRandomVLAN(), types.CONNECT_TYPE_AWS_VIF, "private", 65105, 65105, "684021030471", "notarealauthkey", "10.0.1.0/24", "", "")
	logger.Infof("AWS VIF Connection ID: %s", hostedVifId)

	if !assert.NoError(t, hostedVifErr) && !assert.True(t, shared.IsGuid(hostedVifId)) {
		cfg.PurchaseError(hostedVifId, hostedVifErr)
		t.FailNow()
	}

	vxc.WaitForVXCProvisioning(hostedVifId)

	hostedVIFDeleteStatus, hostedVIFDeleteErr := vxc.DeleteVXC(hostedVifId, true)
	assert.NoError(t, hostedVIFDeleteErr)
	assert.True(t, hostedVIFDeleteStatus)

	portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
	assert.NoError(t, portDeleteErr)
	assert.True(t, portDeleteStatus)
}

func TestAWSConnectionBuyDefaults(t *testing.T) {
	vxc := New(&cfg)
	loc := location.New(&cfg)
	port := port.New(&cfg)

	testLocation, _ := loc.GetLocationByName(TEST_LOCATION_B)
	logger.Info("Buying AWS VIF Port (A End).")
	portId, portErr := port.BuySinglePort("AWS VIF Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	logger.Infof("Port Purchased: %s", portId)

	if !assert.NoError(t, portErr) && !assert.True(t, shared.IsGuid(portId)) {
		cfg.PurchaseError(portId, portErr)
		t.FailNow()
	}
	port.WaitForPortProvisioning(portId)

	logger.Info("Buying AWS VIF Connection (B End).")
	hostedVifId, hostedVifErr := vxc.BuyAWSHostedVIF(portId, "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86", "Hosted AWS VIF Test Connection", 500, 0, types.CONNECT_TYPE_AWS_VIF, "private", 65105, 65106, "684021030471", "", "", "", "")
	logger.Infof("AWS VIF Connection ID: %s", hostedVifId)

	if !assert.NoError(t, hostedVifErr) && !assert.True(t, shared.IsGuid(hostedVifId)) {
		cfg.PurchaseError(hostedVifId, hostedVifErr)
		t.FailNow()
	}

	vxc.WaitForVXCProvisioning(hostedVifId)

	hostedVIFDeleteStatus, hostedVIFDeleteErr := vxc.DeleteVXC(hostedVifId, true)
	assert.NoError(t, hostedVIFDeleteErr)
	assert.True(t, hostedVIFDeleteStatus)

	portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
	assert.NoError(t, portDeleteErr)
	assert.True(t, portDeleteStatus)

}

func TestBuyAzureExpressRoute(t *testing.T) {
	vxc := New(&cfg)
	loc := location.New(&cfg)
	port := port.New(&cfg)

	fuzzySearch, _ := loc.GetLocationByNameFuzzy(TEST_LOCATION_A)
	testLocation := fuzzySearch[0]
	logger.Info("Buying Azure ExpressRoute Port (A End).")
	portId, portErr := port.BuySinglePort("Azure ExpressRoute Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	logger.Infof("Port Purchased: %s", portId)

	if !assert.NoError(t, portErr) && !assert.True(t, shared.IsGuid(portId)) {
		t.FailNow()
	}
	port.WaitForPortProvisioning(portId)
	serviceKey := "9d025691-38dc-48f3-9f95-fbb42e1a9f92"
	peers := map[string]bool{
		"private":   true,
		"public":    false,
		"microsoft": false,
	}

	logger.Info("Buying Azure ExpressRoute VXC (B End).")
	expressRouteId, buyErr := vxc.BuyAzureExpressRoute(portId, "Test Express Route", 1000, 0, serviceKey, peers)

	if buyErr != nil {
		cfg.PurchaseError(expressRouteId, buyErr)
	}

	if !assert.NoError(t, buyErr) && !assert.True(t, shared.IsGuid(expressRouteId)) {
		cfg.PurchaseError(expressRouteId, buyErr)
		t.FailNow()
	}

	logger.Infof("Express Route ID: %s", expressRouteId)
	vxc.WaitForVXCProvisioning(expressRouteId)

	expressRouteDeleteStatus, expressRouteDeleteErr := vxc.DeleteVXC(expressRouteId, true)
	assert.NoError(t, expressRouteDeleteErr)
	assert.True(t, expressRouteDeleteStatus)

	portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
	assert.NoError(t, portDeleteErr)
	assert.True(t, portDeleteStatus)

}

func TestBuyGoogleInterconnect(t *testing.T) {
	vxc := New(&cfg)
	loc := location.New(&cfg)
	port := port.New(&cfg)

	testLocation, _ := loc.GetLocationByName(TEST_LOCATION_B)
	logger.Info("Buying Google Interconnect Port (A End).")
	portId, portErr := port.BuySinglePort("Google Interconnect Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	logger.Infof("Port Purchased: %s", portId)

	if !assert.NoError(t, portErr) && !assert.True(t, shared.IsGuid(portId)) {
		t.FailNow()
	}
	port.WaitForPortProvisioning(portId)

	pairingKey := "7e51371e-72a3-40b5-b844-2e3efefaee59/us-central1/2"
	logger.Info("Buying Google Interconnect VXC (B End).")
	googleInterconnectId, buyErr := vxc.BuyGoogleInterconnect(portId, "Test Google Interconnect", 1000, 0, pairingKey)

	if buyErr != nil {
		cfg.PurchaseError(googleInterconnectId, buyErr)
	}

	if !assert.NoError(t, buyErr) && !assert.True(t, shared.IsGuid(googleInterconnectId)) {
		cfg.PurchaseError(googleInterconnectId, buyErr)
		t.FailNow()
	}

	logger.Infof("Google Interconnect ID: %s", googleInterconnectId)
	vxc.WaitForVXCProvisioning(googleInterconnectId)

	googleInterconnectDeleteStatus, googleInterconnectDeleteErr := vxc.DeleteVXC(googleInterconnectId, true)
	assert.NoError(t, googleInterconnectDeleteErr)
	assert.True(t, googleInterconnectDeleteStatus)

	portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
	assert.NoError(t, portDeleteErr)
	assert.True(t, portDeleteStatus)

}

func TestBuyGoogleInterconnectLocation(t *testing.T) {
	vxc := New(&cfg)
	loc := location.New(&cfg)
	port := port.New(&cfg)

	testLocation, _ := loc.GetLocationByName(TEST_LOCATION_B)
	logger.Info("Buying Google Interconnect Port (A End).")
	portId, portErr := port.BuySinglePort("Google Interconnect Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	logger.Infof("Port Purchased: %s", portId)

	if !assert.NoError(t, portErr) && !assert.True(t, shared.IsGuid(portId)) {
		t.FailNow()
	}
	port.WaitForPortProvisioning(portId)
	pairingKey := "7e51371e-72a3-40b5-b844-2e3efefaee59/us-central1/2"
	logger.Info("Buying Google Interconnect VXC (B End).")
	googleInterconnectId, buyErr := vxc.BuyGoogleInterconnectLocation(portId, "Test Google Interconnect", 1000, 0, pairingKey, TEST_LOCATION_C)

	if buyErr != nil {
		cfg.PurchaseError(googleInterconnectId, buyErr)
	}

	if !assert.NoError(t, buyErr) && !assert.True(t, shared.IsGuid(googleInterconnectId)) {
		cfg.PurchaseError(googleInterconnectId, buyErr)
		t.FailNow()
	}

	logger.Infof("Google Interconnect ID: %s", googleInterconnectId)
	vxc.WaitForVXCProvisioning(googleInterconnectId)

	googleInterconnectDeleteStatus, googleInterconnectDeleteErr := vxc.DeleteVXC(googleInterconnectId, true)
	assert.NoError(t, googleInterconnectDeleteErr)
	assert.True(t, googleInterconnectDeleteStatus)

	portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
	assert.NoError(t, portDeleteErr)
	assert.True(t, portDeleteStatus)
}
