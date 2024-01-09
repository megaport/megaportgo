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

package vxc

// TODO: Add in tests for port-port using Service Keys.

import (
	"os"
	"testing"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/service/authentication"
	"github.com/megaport/megaportgo/service/location"
	"github.com/megaport/megaportgo/service/mcr"
	"github.com/megaport/megaportgo/service/port"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/assert"
)

const (
	MEGAPORTURL     = "https://api-staging.megaport.com/"
	TEST_LOCATION_A = "Global Switch Sydney"
	TEST_LOCATION_B = "Equinix SY3"
	TEST_LOCATION_C = "90558833-e14f-49cf-84ba-bce1c2c40f2d"
	MCR_LOCATION    = "AU"
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

	auth := authentication.New(&cfg)
	token, loginErr := auth.LoginOauth(clientID, clientSecret)

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

	logger.Info("Buying AWS Hosted Connection (B End).")
	vxcId, vxcErr := vxc.BuyVXC(
		aEnd,
		"Test VXC",
		500,
		types.VXCOrderAEndConfiguration{
			VLAN: shared.GenerateRandomVLAN(),
		},
		types.VXCOrderBEndConfiguration{
			VLAN:       shared.GenerateRandomVLAN(),
			ProductUID: bEnd,
		},
	)

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

func TestAWSVIFConnectionBuy(t *testing.T) {
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
	hostedVifId, hostedVifErr := vxc.BuyAWSVXC(
		portId,
		"Hosted AWS VIF Test Connection",
		500,
		types.VXCOrderAEndConfiguration{
			VLAN: shared.GenerateRandomVLAN(),
		},
		types.AWSVXCOrderBEndConfiguration{
			ProductUID: "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86",
			PartnerConfig: types.AWSVXCOrderBEndPartnerConfig{
				ConnectType:  "AWS",
				Type:         "private",
				ASN:          65105,
				AmazonASN:    65106,
				OwnerAccount: "684021030471",
				AuthKey:      "notarealauthkey",
				Prefixes:     "10.0.1.0/24",
			},
		},
	)

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

func TestAWSHostedConnectionBuy(t *testing.T) {
	vxc := New(&cfg)
	loc := location.New(&cfg)
	mcr := mcr.New(&cfg)

	testLocation := loc.GetRandom(MCR_LOCATION)

	logger.Info("Buying AWS Hosted Connection MCR (A End).")
	mcrId, mcrErr := mcr.BuyMCR(testLocation.ID, "AWS Hosted Conection Test MCR", 1, 1000, 0)
	logger.Infof("MCR Purchased: %s", mcrId)

	if !assert.NoError(t, mcrErr) && !assert.True(t, shared.IsGuid(mcrId)) {
		cfg.PurchaseError(mcrId, mcrErr)
		t.FailNow()
	}

	mcr.WaitForMcrProvisioning(mcrId)

	logger.Info("Buying AWS Hosted Connection (B End).")
	hostedConnectionId, hostedConnectionErr := vxc.BuyAWSVXC(
		mcrId,
		"Hosted Connection AWS Test Connection",
		500,
		types.VXCOrderAEndConfiguration{
			VLAN: shared.GenerateRandomVLAN(),
			PartnerConfig: types.VXCOrderAEndPartnerConfig{
				Interfaces: []types.PartnerConfigInterface{
					{
						IpAddresses: []string{"10.0.0.1/30"},
						IpRoutes: []types.IpRoute{
							{
								Prefix:      "10.0.0.1/32",
								Description: "Static route 1",
								NextHop:     "10.0.0.2",
							},
						},
						NatIpAddresses: []string{"10.0.0.1"},
						Bfd: types.BfdConfig{
							TxInterval: 300,
							RxInterval: 300,
							Multiplier: 3,
						},
						BgpConnections: []types.BgpConnectionConfig{
							{
								PeerAsn:        64512,
								LocalIpAddress: "10.0.0.1",
								PeerIpAddress:  "10.0.0.2",
								Password:       "notARealPAssword",
								Shutdown:       false,
								Description:    "BGP with MED and BFD enabled",
								MedIn:          100,
								MedOut:         100,
								BfdEnabled:     true,
							},
						},
					},
				},
			},
		},
		types.AWSVXCOrderBEndConfiguration{
			ProductUID: "b2e0b6b8-2943-4c44-8a07-9ec13060afb2",
			PartnerConfig: types.AWSVXCOrderBEndPartnerConfig{
				ConnectType:  "AWSHC",
				Type:         "private",
				OwnerAccount: "684021030471",
			},
		},
	)
	logger.Infof("AWS Hosted Connection ID: %s", hostedConnectionId)

	if !assert.NoError(t, hostedConnectionErr) && !assert.True(t, shared.IsGuid(hostedConnectionId)) {
		cfg.PurchaseError(hostedConnectionId, hostedConnectionErr)
		t.FailNow()
	}

	vxc.WaitForVXCProvisioning(hostedConnectionId)

	hostedVIFDeleteStatus, hostedVIFDeleteErr := vxc.DeleteVXC(hostedConnectionId, true)
	assert.NoError(t, hostedVIFDeleteErr)
	assert.True(t, hostedVIFDeleteStatus)

	mcrDeleteStatus, mcrDeleteErr := mcr.DeleteMCR(mcrId, true)
	assert.NoError(t, mcrDeleteErr)
	assert.True(t, mcrDeleteStatus)
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
	hostedVifId, hostedVifErr := vxc.BuyAWSVXC(
		portId,
		"Hosted AWS VIF Test Connection",
		500,
		types.VXCOrderAEndConfiguration{
			VLAN: 0,
		},
		types.AWSVXCOrderBEndConfiguration{
			ProductUID: "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86",
			PartnerConfig: types.AWSVXCOrderBEndPartnerConfig{
				ConnectType:  "AWS",
				Type:         "private",
				ASN:          65105,
				AmazonASN:    65106,
				OwnerAccount: "684021030471",
			},
		},
	)

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

	serviceKey := "1b2329a5-56dc-45d0-8a0d-87b706297777"
	peerings := []types.PartnerOrderAzurePeeringConfig{
		{
			Type:            "private",
			PeerASN:         "64555",
			PrimarySubnet:   "10.0.0.0/30",
			SecondarySubnet: "10.0.0.4/30",
			SharedKey:       "SharedKey1",
			VLAN:            100,
		},
		{
			Type:            "microsoft",
			PeerASN:         "64555",
			PrimarySubnet:   "192.88.99.0/30",
			SecondarySubnet: "192.88.99.4/30",
			Prefixes:        "192.88.99.64/26",
			SharedKey:       "SharedKey2",
			VLAN:            200,
		},
	}

	logger.Info("Buying Azure ExpressRoute VXC (B End).")

	// get partner port
	partnerPortId, partnerLookupErr := vxc.LookupPartnerPorts(serviceKey, 1000, PARTNER_AZURE, "")
	if partnerLookupErr != nil {
		t.FailNow()
	}

	// get partner config
	partnerConfig, partnerConfigErr := vxc.MarshallPartnerConfig(serviceKey, PARTNER_AZURE, peerings)
	if partnerConfigErr != nil {
		t.FailNow()
	}

	expressRouteId, buyErr := vxc.BuyPartnerVXC(
		portId,
		"Azure ExpressRoute Test VXC",
		1000,
		types.VXCOrderAEndConfiguration{
			VLAN: 0,
		},
		types.PartnerOrderBEndConfiguration{
			PartnerPortID: partnerPortId,
			PartnerConfig: partnerConfig,
		},
	)

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

	// get partner port
	partnerPortId, partnerLookupErr := vxc.LookupPartnerPorts(pairingKey, 1000, PARTNER_GOOGLE, "")
	if partnerLookupErr != nil {
		t.FailNow()
	}

	// get partner config
	partnerConfig, partnerConfigErr := vxc.MarshallPartnerConfig(pairingKey, PARTNER_GOOGLE, nil)
	if partnerConfigErr != nil {
		t.FailNow()
	}

	googleInterconnectId, buyErr := vxc.BuyPartnerVXC(
		portId,
		"Test Google Interconnect",
		1000,
		types.VXCOrderAEndConfiguration{
			VLAN: 0,
		},
		types.PartnerOrderBEndConfiguration{
			PartnerPortID: partnerPortId,
			PartnerConfig: partnerConfig,
		},
	)

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
	pairingKey := "7e51371e-72a3-40b5-b844-2e3efefaee59/australia-southeast1/2"
	logger.Info("Buying Google Interconnect VXC (B End).")

	// get partner port
	partnerPortId, partnerLookupErr := vxc.LookupPartnerPorts(pairingKey, 1000, PARTNER_GOOGLE, TEST_LOCATION_C)
	if partnerLookupErr != nil {
		t.FailNow()
	}

	// get partner config
	partnerConfig, partnerConfigErr := vxc.MarshallPartnerConfig(pairingKey, PARTNER_GOOGLE, nil)
	if partnerConfigErr != nil {
		t.FailNow()
	}

	googleInterconnectId, buyErr := vxc.BuyPartnerVXC(
		portId,
		"Test Google Interconnect",
		1000,
		types.VXCOrderAEndConfiguration{
			VLAN: 0,
		},
		types.PartnerOrderBEndConfiguration{
			PartnerPortID: partnerPortId,
			PartnerConfig: partnerConfig,
		},
	)

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
