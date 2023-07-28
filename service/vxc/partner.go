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

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
)

const PARTNER_AZURE string = "AZURE"
const PARTNER_GOOGLE string = "GOOGLE"
const PARTNER_AWS string = "AWS"
const PARTNER_OCI string = "ORACLE"

// LookupPartnerPorts is used to find available partner ports. This is Step 1 of the purchase process for most partner
// ports as outlined at https://dev.megaport.com/#cloud-partner-api-orders.
func (v *VXC) LookupPartnerPorts(key string, portSpeed int, partner string, requestedProductID string) (string, error) {
	lookupUrl := "/v2/secure/" + strings.ToLower(partner) + "/" + key
	response, resErr := v.Config.MakeAPICall("GET", lookupUrl, nil)
	defer response.Body.Close()
	isErr, compiledErr := v.Config.IsErrorResponse(response, &resErr, 200)

	if isErr {
		return "", compiledErr
	}

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return "", fileErr
	}

	lookupResponse := types.PartnerLookupResponse{}
	parseErr := json.Unmarshal([]byte(body), &lookupResponse)

	if parseErr != nil {
		return "", parseErr
	}

	for i := 0; i < len(lookupResponse.Data.Megaports); i++ {
		if lookupResponse.Data.Megaports[i].VXC == 0 && lookupResponse.Data.Megaports[i].PortSpeed >= portSpeed { // nil is 0
			// We only need the first available one that has enough speed capacity.
			if requestedProductID == "" {
				return lookupResponse.Data.Megaports[i].ProductUID, nil
				// Try to match Product ID if provided
			} else if lookupResponse.Data.Megaports[i].ProductUID == requestedProductID {
				return lookupResponse.Data.Megaports[i].ProductUID, nil
			}
		}
	}

	return "", errors.New(mega_err.ERR_NO_AVAILABLE_VXC_PORTS)
}

// BuyAWSVXC buys an AWS VXC.
func (v *VXC) BuyPartnerVXC(
	portUID string,
	vxcName string,
	rateLimit int,
	aEndConfiguration types.VXCOrderAEndConfiguration,
	bEndConfiguration types.PartnerOrderBEndConfiguration,
) (string, error) {

	buyOrder := []types.PartnerOrder{
		{
			PortID: portUID,
			AssociatedVXCs: []types.PartnerOrderContents{
				{
					Name:      vxcName,
					RateLimit: rateLimit,
					AEnd:      aEndConfiguration,
					BEnd:      bEndConfiguration,
				},
			},
		},
	}

	requestBody, _ := json.Marshal(buyOrder)

	responseBody, responseErr := v.product.ExecuteOrder(&requestBody)

	if responseErr != nil {
		return "", responseErr
	}

	orderInfo := types.VXCOrderResponse{}
	err := json.Unmarshal(*responseBody, &orderInfo)

	if err != nil {
		return "", err
	}

	return orderInfo.Data[0].TechnicalServiceUID, nil
}

// BuyPartnerVXC performs Step 2 of the partner port purchase process. These are for partners that require some kind
// of partner pairing key (e.g. GCP, Azure).
func (v *VXC) MarshallPartnerConfig(
	key string,
	partner string,
	attributes interface{},
) (interface{}, error) {

	var partnerConfig interface{} = nil

	if partner == PARTNER_AZURE {
		azurePeerings := attributes.([]types.PartnerOrderAzurePeeringConfig)
		partnerConfig = types.PartnerOrderAzurePartnerConfig{
			ConnectType: partner,
			ServiceKey:  key,
			Peers:       azurePeerings,
		}
	} else if partner == PARTNER_GOOGLE {
		partnerConfig = types.PartnerOrderGooglePartnerConfig{
			ConnectType: partner,
			PairingKey:  key,
		}
	} else if partner == PARTNER_AWS {
		// Marshal/unmarshal via JSON so we can reuse struct field mappings
		partnerConfigJson, err := json.Marshal(attributes)
		if err != nil {
			return nil, err
		}
		newPartnerConfig := types.AWSVXCOrderBEndPartnerConfig{}
		if err := json.Unmarshal(partnerConfigJson, &newPartnerConfig); err != nil {
			return nil, err
		}
		partnerConfig = newPartnerConfig
	} else if partner == PARTNER_OCI {
		partnerConfig = types.PartnerOrderOciPartnerConfig{
			ConnectType: partner,
			VirtualCircutId:  key,
		}
	} else {
		return "", errors.New(mega_err.ERR_INVALID_PARTNER)
	}

	return partnerConfig, nil
}
