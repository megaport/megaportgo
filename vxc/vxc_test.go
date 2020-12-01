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
	"testing"

	"github.com/megaport/megaportgo/location"
	"github.com/megaport/megaportgo/port"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

const TEST_LOCATION_A = "Global Switch Sydney"
const TEST_LOCATION_B = "Equinix SY3"

func TestVXCBuy(t *testing.T) {
	assert := assert.New(t)
	fuzzySearch, locationErr := location.GetLocationByNameFuzzy(TEST_LOCATION_A)
	testLocation := fuzzySearch[0]

	assert.NoError(locationErr)
	log.Info().Msg("Buying Port (A End).")
	aEnd, aErr := port.BuySinglePort("VXC Port A", 1, 1000, testLocation.ID, "AU", true)
	log.Info().Msgf("Port Purchased: '%s'.", aEnd)

	log.Info().Msg("Buying Port (B End).")
	bEnd, bErr := port.BuySinglePort("VXC Port B", 1, 1000, testLocation.ID, "AU", true)
	log.Info().Msgf("Port Purchased: '%s'.", bEnd)

	if assert.NoError(aErr) && assert.True(shared.IsGuid(aEnd)) && assert.NoError(bErr) && assert.True(shared.IsGuid(bEnd)) {
		port.WaitForPortProvisioning(aEnd)
		port.WaitForPortProvisioning(bEnd)

		log.Info().Msgf("Buying VXC.")

		vxcId, vxcErr := BuyVXC(aEnd, bEnd, "Test VXC", 500, shared.GenerateRandomVLAN(), shared.GenerateRandomVLAN())
		log.Info().Msgf("VXC Purchased: '%s'.", vxcId)

		if assert.NoError(vxcErr) && assert.True(shared.IsGuid(vxcId)) {
			WaitForVXCProvisioning(vxcId)

			newAVLAN := shared.GenerateRandomVLAN()
			newBVLAN := shared.GenerateRandomVLAN()
			updateStatus, updateErr := UpdateVXC(vxcId, "VXC Update Name Test", 1000, newAVLAN, newBVLAN)
			assert.True(updateStatus)
			assert.NoError(updateErr)
			WaitForVXCUpdated(vxcId, "VXC Update Name Test", 1000, newAVLAN, newBVLAN)
			vxcInfo, _ := GetVXCDetails(vxcId)
			assert.EqualValues("VXC Update Name Test", vxcInfo.Name)
			assert.EqualValues(1000, vxcInfo.RateLimit)
			assert.EqualValues(newAVLAN, vxcInfo.AEndConfiguration.VLAN)
			assert.EqualValues(newBVLAN, vxcInfo.BEndConfiguration.VLAN)

			vxcDeleteStatus, vxcDeleteErr := DeleteVXC(vxcId, true)
			assert.NoError(vxcDeleteErr)
			assert.True(vxcDeleteStatus, nil)

			aDeleteStatus, aDeleteErr := port.DeletePort(aEnd, true)
			assert.NoError(aDeleteErr)
			assert.True(aDeleteStatus)

			bDeleteStatus, bDeleteErr := port.DeletePort(bEnd, true)
			assert.NoError(bDeleteErr)
			assert.True(bDeleteStatus)
		} else {
			shared.PurchaseError(vxcId, vxcErr)
		}
	} else {
		shared.PurchaseError(aEnd, aErr)
		shared.PurchaseError(bEnd, bErr)
	}
}

func TestAWSConnectionBuy(t *testing.T) {
	testLocation, _ := location.GetLocationByName(TEST_LOCATION_B)
	log.Info().Msg("Buying AWS VIF Port (A End).")
	portId, portErr := port.BuySinglePort("AWS VIF Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	log.Info().Msgf("Port Purchased: '%s'.", portId)

	if assert.NoError(t, portErr) && assert.True(t, shared.IsGuid(portId)) {
		port.WaitForPortProvisioning(portId)

		log.Info().Msgf("Buying AWS VIF Connection (B End).")
		hostedVifId, hostedVifErr := BuyAWSHostedVIF(portId, "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86", "Hosted AWS VIF Test Connection", 500, shared.GenerateRandomVLAN(), types.CONNECT_TYPE_AWS_VIF, "private", 65105, 65105, "684021030471", "notarealauthkey", "10.0.1.0/24", "", "")
		log.Info().Msgf("AWS VIF Connection ID: '%s.'", hostedVifId)
		if assert.NoError(t, hostedVifErr) && assert.True(t, shared.IsGuid(hostedVifId)) {
			WaitForVXCProvisioning(hostedVifId)

			hostedVIFDeleteStatus, hostedVIFDeleteErr := DeleteVXC(hostedVifId, true)
			assert.NoError(t, hostedVIFDeleteErr)
			assert.True(t, hostedVIFDeleteStatus)

			portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
			assert.NoError(t, portDeleteErr)
			assert.True(t, portDeleteStatus)
		} else {
			shared.PurchaseError(hostedVifId, hostedVifErr)
		}
	} else {
		shared.PurchaseError(portId, portErr)
	}
}

func TestAWSConnectionBuyDefaults(t *testing.T) {
	testLocation, _ := location.GetLocationByName(TEST_LOCATION_B)
	log.Info().Msg("Buying AWS VIF Port (A End).")
	portId, portErr := port.BuySinglePort("AWS VIF Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	log.Info().Msgf("Port Purchased: '%s'.", portId)

	if assert.NoError(t, portErr) && assert.True(t, shared.IsGuid(portId)) {
		port.WaitForPortProvisioning(portId)

		log.Info().Msgf("Buying AWS VIF Connection (B End).")
		hostedVifId, hostedVifErr := BuyAWSHostedVIF(portId, "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86", "Hosted AWS VIF Test Connection", 500, 0, types.CONNECT_TYPE_AWS_VIF, "private", 65105, 65106, "684021030471", "", "", "", "")
		log.Info().Msgf("AWS VIF Connection ID: '%s.'", hostedVifId)
		if assert.NoError(t, hostedVifErr) && assert.True(t, shared.IsGuid(hostedVifId)) {
			WaitForVXCProvisioning(hostedVifId)

			hostedVIFDeleteStatus, hostedVIFDeleteErr := DeleteVXC(hostedVifId, true)
			assert.NoError(t, hostedVIFDeleteErr)
			assert.True(t, hostedVIFDeleteStatus)

			portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
			assert.NoError(t, portDeleteErr)
			assert.True(t, portDeleteStatus)
		} else {
			shared.PurchaseError(hostedVifId, hostedVifErr)
		}
	} else {
		shared.PurchaseError(portId, portErr)
	}
}

func TestBuyAzureExpressRoute(t *testing.T) {
	fuzzySearch, _ := location.GetLocationByNameFuzzy(TEST_LOCATION_A)
	testLocation := fuzzySearch[0]
	log.Info().Msg("Buying Azure ExpressRoute Port (A End).")
	portId, portErr := port.BuySinglePort("Azure ExpressRoute Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	log.Info().Msgf("Port Purchased: '%s'.", portId)

	if assert.NoError(t, portErr) && assert.True(t, shared.IsGuid(portId)) {
		port.WaitForPortProvisioning(portId)
		serviceKey := "9d025691-38dc-48f3-9f95-fbb42e1a9f92"
		log.Info().Msgf("Buying Azure ExpressRoute VXC (B End).")
		expressRouteId, buyErr := BuyAzureExpressRoute(portId, "Test Express Route", 1000, 0, serviceKey)

		if buyErr != nil {
			shared.PurchaseError(expressRouteId, buyErr)
		}

		if assert.NoError(t, buyErr) && assert.True(t, shared.IsGuid(expressRouteId)) {
			log.Info().Msgf("Express Route ID: '%s'", expressRouteId)
			WaitForVXCProvisioning(expressRouteId)

			expressRouteDeleteStatus, expressRouteDeleteErr := DeleteVXC(expressRouteId, true)
			assert.NoError(t, expressRouteDeleteErr)
			assert.True(t, expressRouteDeleteStatus)

			portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
			assert.NoError(t, portDeleteErr)
			assert.True(t, portDeleteStatus)
		} else {
			shared.PurchaseError(expressRouteId, buyErr)
		}
	}
}

func TestBuyGoogleInterconnect(t *testing.T) {
	testLocation, _ := location.GetLocationByName(TEST_LOCATION_B)
	log.Info().Msg("Buying Google Interconnect Port (A End).")
	portId, portErr := port.BuySinglePort("Google Interconnect Test Port", 1, 1000, int(testLocation.ID), "AU", true)
	log.Info().Msgf("Port Purchased: '%s'.", portId)

	if assert.NoError(t, portErr) && assert.True(t, shared.IsGuid(portId)) {
		port.WaitForPortProvisioning(portId)
		pairingKey := "7e51371e-72a3-40b5-b844-2e3efefaee59/us-central1/2"
		log.Info().Msgf("Buying Google Interconnect VXC (B End).")
		googleInterconnectId, buyErr := BuyGoogleInterconnect(portId, "Test Google Interconnect", 1000, 0, pairingKey)

		if buyErr != nil {
			shared.PurchaseError(googleInterconnectId, buyErr)
		}

		if assert.NoError(t, buyErr) && assert.True(t, shared.IsGuid(googleInterconnectId)) {
			log.Info().Msgf("Google Interconnect ID: '%s'", googleInterconnectId)
			WaitForVXCProvisioning(googleInterconnectId)

			googleInterconnectDeleteStatus, googleInterconnectDeleteErr := DeleteVXC(googleInterconnectId, true)
			assert.NoError(t, googleInterconnectDeleteErr)
			assert.True(t, googleInterconnectDeleteStatus)

			portDeleteStatus, portDeleteErr := port.DeletePort(portId, true)
			assert.NoError(t, portDeleteErr)
			assert.True(t, portDeleteStatus)
		} else {
			shared.PurchaseError(googleInterconnectId, buyErr)
		}
	}
}
