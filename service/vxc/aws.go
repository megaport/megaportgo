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

// The `vxc` package houses functions that mangage VXCs, including those between two Megaports (generic) and those
// between vendors on the Megaport Marketplace.
package vxc

import (
	"encoding/json"

	"github.com/megaport/megaportgo/types"
)

// BuyAWSVXC buys an AWS VXC.
func (v *VXC) BuyAWSVXC(
	portUID string,
	vxcName string,
	rateLimit int,
	aEndConfiguration types.AWSVXCOrderAEndConfiguration,
	bEndConfiguration types.AWSVXCOrderBEndConfiguration,
) (string, error) {

	buyOrder := []types.AWSVXCOrder{
		types.AWSVXCOrder{
			PortID: portUID,
			AssociatedVXCs: []types.AWSVXCOrderConfiguration{
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
