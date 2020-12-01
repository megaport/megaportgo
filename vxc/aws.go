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
	"github.com/megaport/megaportgo/product"
	"github.com/megaport/megaportgo/types"
)

// BuyAWSHostedVIF buys an AWS Hosted VIF.
func BuyAWSHostedVIF(portUID string, productUID string, name string, rateLimit int, aEndVLAN int, connectType string, vifType string, asn int, amazonASN int, ownerAccount string, authKey string, prefixes string, customerIPAddress string, amazonIPAddress string) (string, error) {
	buyOrder := []types.AWSHostedVIFOrder{
		types.AWSHostedVIFOrder{
			PortID: portUID,
			AssociatedVXCs: []types.AWSHostedVIFOrderConfiguration{
				{
					Name:      name,
					RateLimit: rateLimit,
					AEnd: types.VXCOrderAEndConfiguration{
						VLAN: aEndVLAN,
					},
					BEnd: types.AWSHostedVIFOrderBEndConfiguration{
						ProductUID: productUID,
						PartnerConfig: types.AWSHostedVIFOrderPartnerConfig{
							ConnectType:       connectType,
							Type:              vifType,
							ASN:               asn,
							AmazonASN:         amazonASN,
							OwnerAccount:      ownerAccount,
							AuthKey:           authKey,
							Prefixes:          prefixes,
							CustomerIPAddress: customerIPAddress,
							AmazonIPAddress:   amazonIPAddress,
						},
					},
				},
			},
		},
	}

	requestBody, _ := json.Marshal(buyOrder)
	responseBody, responseErr := product.ExecuteOrder(&requestBody)

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
