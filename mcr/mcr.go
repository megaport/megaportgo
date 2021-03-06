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

// the `mcr` package is designed to allow you to orchestrate the creation of Megaport Cloud Routers. It provides
// complete lifecycle management of an MCR.
package mcr

import (
	"encoding/json"
	"errors"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/product"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"strings"
	"time"
)

// BuyMCR purchases an MCR.
func BuyMCR(locationID int, name string, portSpeed int, mcrASN int) (string, error) {
	orderConfig := types.MCROrderConfig{}

	if mcrASN != 0 {
		orderConfig.ASN = mcrASN
	}

	if portSpeed != 1000 && portSpeed != 2500 && portSpeed != 5000 && portSpeed != 10000 {
		return "", errors.New(mega_err.ERR_MCR_INVALID_PORT_SPEED)
	}

	order := []types.MCROrder{
		{
			LocationID: locationID,
			Name:       name,
			Type:       "MCR2",
			PortSpeed:  portSpeed,
			Config:     orderConfig,
		},
	}

	requestBody, marshalErr := json.Marshal(order)

	if marshalErr != nil {
		return "", marshalErr
	}

	body, resErr := product.ExecuteOrder(&requestBody)

	if resErr != nil {
		return "", resErr
	}

	orderInfo := types.MCROrderResponse{}
	unmarshalErr := json.Unmarshal(*body, &orderInfo)

	if unmarshalErr != nil {
		return "", unmarshalErr
	}

	return orderInfo.Data[0].TechnicalServiceUID, nil
}

// BuyMCR get the details of an MCR.
func GetMCRDetails(id string) (types.MCR, error) {
	url := "/v2/product/" + id
	response, err := shared.MakeAPICall("GET", url, nil)
	defer response.Body.Close()

	isError, parsedError := shared.IsErrorResponse(response, &err, 200)

	if isError {
		return types.MCR{}, parsedError
	}

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return types.MCR{}, fileErr
	}

	portDetails := types.MCRResponse{}
	unmarshalErr := json.Unmarshal(body, &portDetails)

	if unmarshalErr != nil {
		return types.MCR{}, unmarshalErr
	}

	return portDetails.Data, nil
}

// ModifyMCR modifies an MCR.
func ModifyMCR(mcrId string, name string, costCentre string, marketplaceVisibility bool) (bool, error) {
	return product.ModifyProduct(mcrId, types.PRODUCT_MCR, name, costCentre, marketplaceVisibility)
}

// ModifyMCR deletes an MCR.
func DeleteMCR(id string, deleteNow bool) (bool, error) {
	return product.DeleteProduct(id, deleteNow)
}

// ModifyMCR un-deletes an MCR.
func RestoreMCR(id string) (bool, error) {
	return product.RestoreProduct(id)
}

// DebugWaitMCRLive will should be used for testing only.
func WaitForMcrProvisioning(mcrId string) (bool, error) {
	mcrInfo, _ := GetMCRDetails(mcrId)
	wait := 0

	// Go-Live
	log.Debug().Msg("Waiting for MCR to transition to 'LIVE'.")
	for strings.Compare(mcrInfo.ProvisioningStatus, "LIVE") != 0 && wait < 30 {
		time.Sleep(30 * time.Second)
		wait++
		mcrInfo, _ = GetMCRDetails(mcrId)

		if wait%5 == 0 {
			log.Debug().Str("Status", mcrInfo.ProvisioningStatus).Msg("MCR is currently being provisioned.")
		}
	}

	mcrInfo, _ = GetMCRDetails(mcrId)
	log.Debug().Str("Status", mcrInfo.ProvisioningStatus).Msg("MCR waiting cycle complete.")

	if mcrInfo.ProvisioningStatus == "LIVE" {
		return true, nil
	} else {
		if wait >= 30 {
			return false, errors.New(mega_err.ERR_MCR_PROVISION_TIMEOUT_EXCEED)
		} else {
			return false, errors.New(mega_err.ERR_MCR_NOT_LIVE)
		}
	}
}
