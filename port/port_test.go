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

package port

import (
	"errors"
	"fmt"
	"github.com/megaport/megaportgo/location"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

const TEST_LOCATION_A = "Interactive 437 Williamstown"

// TestSinglePort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func TestSinglePort(t *testing.T) {
	portId, portErr := testCreatePort(types.SINGLE_PORT)

	if assert.NoError(t, portErr) && assert.True(t, shared.IsGuid(portId)) {
		WaitForPortProvisioning(portId)
		testModifyPort(portId, types.SINGLE_PORT, t)
		testLockPort(portId, t)
		testCancelPort(portId, types.SINGLE_PORT, t)
		testDeletePort(portId, types.SINGLE_PORT, t)
	} else {
		shared.PurchaseError(portId, portErr)
	}
}

// TestLAGPort tests the creation of a LAG Port, then passes the id to PortScript to finalise lifecycle testing.
func TestLAGPort(t *testing.T) {
	portId, portErr := testCreatePort(types.LAG_PORT)

	if assert.NoError(t, portErr) && assert.True(t, shared.IsGuid(portId)) {
		WaitForPortProvisioning(portId)
		testModifyPort(portId, types.LAG_PORT, t)
		testCancelPort(portId, types.LAG_PORT, t)
		testDeletePort(portId, types.LAG_PORT, t)
	} else {
		shared.PurchaseError(portId, portErr)
	}
}

func testCreatePort(portType string) (string, error) {
	log.Printf("Buying %s Port.", portType)
	var portId string
	var portErr error
	testLocation, _ := location.GetLocationByName(TEST_LOCATION_A)

	if portType == types.LAG_PORT {
		portId, portErr = BuyLAGPort("Buy Port (LAG) Test", 1, 10000, testLocation.ID, "AU", 4, true)
	} else {
		portId, portErr = BuySinglePort("Buy Port (Single) Test", 1, 10000, testLocation.ID, "AU", true)
	}

	log.Printf("Port Purchased: %s", portId)
	return portId, portErr
}

func testModifyPort(portId string, portType string, t *testing.T) {
	portInfo, _ := GetPortDetails(portId)
	log.Printf("Modifying %s Port.", portType)
	newPortName := fmt.Sprintf("Buy Port (%s) [Modified]", portType)
	_, modifyErr := ModifyPort(portId, newPortName, "", portInfo.MarketplaceVisibility)
	assert.NoError(t, modifyErr)
	portInfo, _ = GetPortDetails(portId)
	assert.EqualValues(t, newPortName, portInfo.Name)
}

// PortScript tests the remaining lifecycle for a Port (not dependant on port-type), Go-Live, Modification,
// and Soft/Hard Deletes.
func testCancelPort(portId string, portType string, t *testing.T) {
	// Soft Delete
	log.Printf("Scheduling %s Port for deletion (30 days).", portType)
	softDeleteStatus, deleteErr := DeletePort(portId, false)
	assert.True(t, softDeleteStatus)
	assert.NoError(t, deleteErr)
	portInfo, _ := GetPortDetails(portId)
	assert.EqualValues(t, types.STATUS_CANCELLED, portInfo.ProvisioningStatus)
	log.Printf("Status is: '%s'", portInfo.ProvisioningStatus)
	restoreStatus, restoreErr := RestorePort(portId)
	assert.True(t, restoreStatus)
	assert.NoError(t, restoreErr)
}

func testDeletePort(portId string, portType string, t *testing.T) {
	// Hard Delete
	log.Printf("Deleting %s Port now.", portType)
	hardDeleteStatus, deleteErr := DeletePort(portId, true)
	assert.True(t, hardDeleteStatus)
	assert.NoError(t, deleteErr)
	portInfo, _ := GetPortDetails(portId)
	assert.EqualValues(t, types.STATUS_DECOMMISSIONED, portInfo.ProvisioningStatus)
	log.Printf("Status is: '%s'", portInfo.ProvisioningStatus)
}

func testLockPort(portId string, t *testing.T) {
	log.Printf("Locking Port now.")
	lockStatus, lockErr := LockPort(portId)
	assert.True(t, lockStatus)
	assert.NoError(t, lockErr)
	portInfo, _ := GetPortDetails(portId)
	assert.EqualValues(t, true, portInfo.Locked)
	log.Printf("Test lock of an already locked port.")
	lockStatus, lockErr = LockPort(portId)
	assert.True(t, lockStatus)
	assert.Error(t, errors.New(mega_err.ERR_PORT_ALREADY_LOCKED), lockErr)

	log.Printf("Unlocking Port now.")
	unlockStatus, unlockErr := UnlockPort(portId)
	assert.True(t, unlockStatus)
	assert.NoError(t, unlockErr)
	log.Printf("Test unlocking of a port that doesn't have a lock.")
	unlockStatus, unlockErr = UnlockPort(portId)
	assert.True(t, unlockStatus)
	assert.Error(t, errors.New(mega_err.ERR_PORT_NOT_LOCKED), unlockErr)
}
