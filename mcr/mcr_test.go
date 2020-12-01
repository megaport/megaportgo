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
	"github.com/megaport/megaportgo/location"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/megaport/megaportgo/vxc"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

const TEST_MCR_TEST_LOCATION_MARKET = "AU"

func TestMCRLifecycle(t *testing.T) {
	assert := assert.New(t)

	mcrId, portErr := testCreateMCR()

	if assert.NoError(portErr) && assert.True(shared.IsGuid(mcrId)) {
		WaitForMcrProvisioning(mcrId)
		testModifyMCR(mcrId, t)
		testCancelMCR(mcrId, t)
		testDeleteMCR(mcrId, t)
	} else {
		shared.PurchaseError(mcrId, portErr)
	}
}

func TestMCRConnectionAdd(t *testing.T) {
	assert := assert.New(t)
	testLocation := location.GetRandom(TEST_MCR_TEST_LOCATION_MARKET)
	log.Info().Str("Location", testLocation.Name).Msg("Test location determined.")
	log.Print("Buying MCR.")
	mcrId, mcrErr := BuyMCR(testLocation.ID, "MCR and AWS Interconnectivity", 1000, 0)
	log.Printf("MCR Purchased: '%s'.", mcrId)

	if assert.NoError(mcrErr) && assert.True(shared.IsGuid(mcrId)) {
		WaitForMcrProvisioning(mcrId)

		log.Printf("Buying AWS VIF Connection (B End).")
		vifOneId, vifOneErr := vxc.BuyAWSHostedVIF(mcrId, "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86", "MCR and AWS Connection 1", 500, shared.GenerateRandomVLAN(), types.CONNECT_TYPE_AWS_VIF, "private", 65105, 65105, "684021030471", "notarealauthkey", "10.0.1.0/24", "", "")
		vifTwoId, vifTwoErr := vxc.BuyAWSHostedVIF(mcrId, "87860c28-81ef-4e79-8cc7-cfc5a4c4bc86", "MCR and AWS Connection 2", 500, shared.GenerateRandomVLAN(), types.CONNECT_TYPE_AWS_VIF, "private", 65105, 65105, "684021030471", "notarealauthkey", "10.0.2.0/24", "", "")
		log.Printf("AWS VIF Connection 1: '%s.'", vifOneId)
		log.Printf("AWS VIF Connection 2: '%s.'", vifTwoId)

		if assert.NoError(vifOneErr) && assert.NoError(vifTwoErr) && assert.True(shared.IsGuid(vifOneId) && assert.True(shared.IsGuid(vifTwoId))) {
			vxc.WaitForVXCProvisioning(vifOneId)
			vxc.WaitForVXCProvisioning(vifTwoId)

			vifOneDeleteStatus, vifOneDeleteErr := vxc.DeleteVXC(vifOneId, true)
			assert.NoError(vifOneDeleteErr)
			assert.True(vifOneDeleteStatus)

			vifTwoDeleteStatus, vifTwoDeleteErr := vxc.DeleteVXC(vifTwoId, true)
			assert.NoError(vifTwoDeleteErr)
			assert.True(vifTwoDeleteStatus)

			mcrDeleteStatus, mcrDeleteErr := DeleteMCR(mcrId, true)
			assert.NoError(mcrDeleteErr)
			assert.True(mcrDeleteStatus)
		} else {
			shared.PurchaseError(vifOneId, vifOneErr)
			shared.PurchaseError(vifTwoId, vifTwoErr)
		}
	} else {
		shared.PurchaseError(mcrId, mcrErr)
	}
}

func TestPortSpeedValidation(t *testing.T) {
	assert := assert.New(t)
	testLocation, _ := location.GetLocationByName("Global Switch Sydney")
	_, buyErr := BuyMCR(testLocation.ID, "Test MCR", 500, 0)
	assert.EqualError(buyErr, mega_err.ERR_MCR_INVALID_PORT_SPEED)
}

func testCreateMCR() (string, error) {
	log.Printf("Buying MCR Port.")
	testLocation := location.GetRandom(TEST_MCR_TEST_LOCATION_MARKET)
	log.Info().Str("Location", testLocation.Name).Msg("Test location determined.")
	mcrID, mcrErr := BuyMCR(testLocation.ID, "Buy MCR", 1000, 0)

	log.Printf("MCR Purchased: %s", mcrID)
	return mcrID, mcrErr
}

func testModifyMCR(mcrID string, t *testing.T) {
	assert := assert.New(t)
	mcrInfo, _ := GetMCRDetails(mcrID)
	log.Printf("Modifying MCR.")
	newMCRName := fmt.Sprintf("Buy MCR [Modified]")
	_, modifyErr := ModifyMCR(mcrID, newMCRName, "", mcrInfo.MarketplaceVisibility)
	assert.NoError(modifyErr)
	mcrInfo, _ = GetMCRDetails(mcrID)
	assert.EqualValues(newMCRName, mcrInfo.Name)
}

// PortScript tests the remaining lifecycle for a Port (not dependant on port-type), Go-Live, Modification,
// and Soft/Hard Deletes.
func testCancelMCR(mcrId string, t *testing.T) {
	assert := assert.New(t)
	// Soft Delete
	log.Printf("Scheduling MCR for deletion (30 days).")
	softDeleteStatus, deleteErr := DeleteMCR(mcrId, false)
	assert.True(softDeleteStatus)
	assert.NoError(deleteErr)
	mcrInfo, _ := GetMCRDetails(mcrId)
	assert.EqualValues(types.STATUS_CANCELLED, mcrInfo.ProvisioningStatus)
	log.Printf("Status is: '%s'", mcrInfo.ProvisioningStatus)
	restoreStatus, restoreErr := RestoreMCR(mcrId)
	assert.True(restoreStatus)
	assert.NoError(restoreErr)
}

func testDeleteMCR(mcrId string, t *testing.T) {
	assert := assert.New(t)
	// Hard Delete
	log.Printf("Deleting MCR now.")
	hardDeleteStatus, deleteErr := DeleteMCR(mcrId, true)
	assert.True(hardDeleteStatus)
	assert.NoError(deleteErr)
	mcrInfo, _ := GetMCRDetails(mcrId)
	assert.EqualValues(types.STATUS_DECOMMISSIONED, mcrInfo.ProvisioningStatus)
	log.Printf("Status is: '%s'", mcrInfo.ProvisioningStatus)
}
