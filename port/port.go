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

// the `port` package is designed to allow you to orchestrate the creation of Megaport Ports. It provides
// complete lifecycle management of Ports.
package port

import (
	"encoding/json"
	"errors"
	"github.com/megaport/megaportgo/mega_err"
	"io/ioutil"
	"time"

	"github.com/megaport/megaportgo/product"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"github.com/rs/zerolog/log"
)

const MODIFY_NAME string = "NAME"
const MODIFY_COST_CENTRE = "COST_CENTRE"
const MODIFY_MARKETPLACE_VISIBILITY string = "MARKETPLACE_VISIBILITY"

// BuyPort orders a Port.
func BuyPort(name string, term int, portSpeed int, locationId int, market string, isLAG bool, lagCount int, isPrivate bool) (string, error) {
	var buyOrder []types.PortOrder

	if term != 1 && term != 12 && term != 24 && term != 36 {
		return "", errors.New(mega_err.ERR_TERM_NOT_VALID)
	}

	if isLAG {
		buyOrder = []types.PortOrder{
			types.PortOrder{
				Name:                  name,
				Term:                  term,
				ProductType:           "MEGAPORT",
				PortSpeed:             portSpeed,
				LocationID:            locationId,
				CreateDate:            shared.GetCurrentTimestamp(),
				Virtual:               false,
				Market:                market,
				LagPortCount:          lagCount,
				MarketplaceVisibility: !isPrivate,
			},
		}
	} else {
		buyOrder = []types.PortOrder{
			types.PortOrder{
				Name:                  name,
				Term:                  term,
				ProductType:           "MEGAPORT",
				PortSpeed:             portSpeed,
				LocationID:            locationId,
				CreateDate:            shared.GetCurrentTimestamp(),
				Virtual:               false,
				Market:                market,
				MarketplaceVisibility: !isPrivate,
			},
		}
	}

	requestBody, _ := json.Marshal(buyOrder)
	responseBody, responseErr := product.ExecuteOrder(&requestBody)

	if responseErr != nil {
		return "", responseErr
	}

	orderInfo := types.PortOrderResponse{}
	unmarshalErr := json.Unmarshal(*responseBody, &orderInfo)

	if unmarshalErr != nil {
		return "", unmarshalErr
	}

	return orderInfo.Data[0].TechnicalServiceUID, nil
}

// BuyPort orders a single Port. Same as BuyPort, with isLag set to false.
func BuySinglePort(name string, term int, portSpeed int, locationId int, market string, isPrivate bool) (string, error) {
	return BuyPort(name, term, portSpeed, locationId, market, false, 0, isPrivate)
}

// BuyPort orders a LAG Port. Same as BuyPort, with isLag set to true.
func BuyLAGPort(name string, term int, portSpeed int, locationId int, market string, lagCount int, isPrivate bool) (string, error) {
	return BuyPort(name, term, portSpeed, locationId, market, true, lagCount, isPrivate)
}

func GetPortDetails(id string) (types.Port, error) {
	url := "/v2/product/" + id
	response, err := shared.MakeAPICall("GET", url, nil)
	defer response.Body.Close()

	isError, parsedError := shared.IsErrorResponse(response, &err, 200)

	if isError {
		return types.Port{}, parsedError
	}

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return types.Port{}, fileErr
	}

	portDetails := types.PortResponse{}
	unmarshalErr := json.Unmarshal(body, &portDetails)

	if unmarshalErr != nil {
		return types.Port{}, unmarshalErr
	}

	return portDetails.Data, nil
}

func ModifyPort(portId string, name string, costCentre string, marketplaceVisibility bool) (bool, error) {
	return product.ModifyProduct(portId, types.PRODUCT_MEGAPORT, name, costCentre, marketplaceVisibility)
}

func DeletePort(id string, deleteNow bool) (bool, error) {
	return product.DeleteProduct(id, deleteNow)
}

func RestorePort(id string) (bool, error) {
	return product.RestoreProduct(id)
}

// TODO: Tests for locking.
func LockPort(id string) (bool, error) {
	portInfo, _ := GetPortDetails(id)
	if !portInfo.Locked {
		return product.ManageProductLock(id, true)
	} else {
		return true, errors.New(mega_err.ERR_PORT_ALREADY_LOCKED)
	}
}

func UnlockPort(id string) (bool, error) {
	portInfo, _ := GetPortDetails(id)
	if portInfo.Locked {
		return product.ManageProductLock(id, false)
	} else {
		return true, errors.New(mega_err.ERR_PORT_NOT_LOCKED)
	}
}

func WaitForPortProvisioning(portId string) (bool, error) {
	portInfo, _ := GetPortDetails(portId)
	wait := 0

	log.Debug().Msg("Waiting for port status transition.")
	for portInfo.ProvisioningStatus != "CONFIGURED" && portInfo.ProvisioningStatus  != "LIVE" && wait < 30 {
		time.Sleep(30 * time.Second)
		wait++
		portInfo, _ = GetPortDetails(portId)

		if wait%5 == 0 {
			log.Debug().Str("Status", portInfo.ProvisioningStatus).Msg("Port is currently being provisioned.")
		}
	}

	portInfo, _ = GetPortDetails(portId)
	log.Debug().Str("Status", portInfo.ProvisioningStatus).Msg("Port waiting cycle complete.")

	if portInfo.ProvisioningStatus == "CONFIGURED" || portInfo.ProvisioningStatus == "LIVE" {
		return true, nil
	} else {
		if wait >= 30 {
			return false, errors.New(mega_err.ERR_PORT_PROVISION_TIMEOUT_EXCEED)
		} else {
			return false, errors.New(mega_err.ERR_PORT_NOT_LIVE)
		}
	}
}
