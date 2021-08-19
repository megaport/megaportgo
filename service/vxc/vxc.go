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
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/service/product"
	"github.com/megaport/megaportgo/types"
)

type VXC struct {
	*config.Config
	product *product.Product
}

func New(cfg *config.Config) *VXC {
	return &VXC{
		Config:  cfg,
		product: product.New(cfg),
	}
}

// BuyVXC purchases a generic VXC between two Megaport Ports. The productUID should be the Service Key for a port if
// it is in another account, otherwise it should be the port UID.
func (v *VXC) BuyVXC(portUID string, productUID string, name string, rateLimit int, aEndVLAN int, bEndVLAN int) (string, error) {
	buyOrder := []types.VXCOrder{
		{
			PortID: portUID,
			AssociatedVXCs: []types.VXCConfiguration{
				{
					Name:      name,
					RateLimit: rateLimit,
					AEnd: types.VXCOrderAEndConfiguration{
						VLAN: aEndVLAN,
					},
					BEnd: types.VXCOrderBEndConfiguration{
						ProductUID: productUID,
						VLAN:       bEndVLAN,
					},
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

// GetVXCDetails gets the details of a VXC.
func (v *VXC) GetVXCDetails(id string) (types.VXC, error) {
	url := "/v2/product/" + id
	response, err := v.Config.MakeAPICall("GET", url, nil)
	defer response.Body.Close()

	if err != nil {
		return types.VXC{}, err
	}

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return types.VXC{}, fileErr
	}

	vxcDetails := types.VXCResponse{}
	unmarshalErr := json.Unmarshal(body, &vxcDetails)

	if unmarshalErr != nil {
		return types.VXC{}, unmarshalErr
	}

	return vxcDetails.Data, nil
}

// GetVXCDetails deletes a VXC.
func (v *VXC) DeleteVXC(id string, deleteNow bool) (bool, error) {
	return v.product.DeleteProduct(id, deleteNow)
}

func (v *VXC) UpdateVXC(id string, name string, rateLimit int, aEndVLAN int, bEndVLAN int) (bool, error) {
	url := fmt.Sprintf("/v2/product/%s/%s", types.PRODUCT_VXC, id)
	var update interface{}

	if bEndVLAN == 0 {
		update = types.PartnerVXCUpdate{
			Name:      name,
			RateLimit: rateLimit,
			AEndVLAN:  aEndVLAN,
		}
	} else {
		update = types.VXCUpdate{
			Name:      name,
			RateLimit: rateLimit,
			AEndVLAN:  aEndVLAN,
			BEndVLAN:  &bEndVLAN,
		}
	}

	body, marshalErr := json.Marshal(update)

	if marshalErr != nil {
		return false, marshalErr
	}

	updateResponse, err := v.Config.MakeAPICall("PUT", url, []byte(body))
	isResErr, compiledResErr := v.Config.IsErrorResponse(updateResponse, &err, 200)

	if isResErr {
		return false, compiledResErr
	} else {
		return true, nil
	}
}

func (v *VXC) WaitForVXCProvisioning(vxcId string) (bool, error) {
	vxcInfo, _ := v.GetVXCDetails(vxcId)
	wait := 0

	// Go-Live
	v.Log.Info("Waiting for VXC status transition.")
	for strings.Compare(vxcInfo.ProvisioningStatus, "LIVE") != 0 && wait < 30 {
		time.Sleep(30 * time.Second)
		wait++
		vxcInfo, _ = v.GetVXCDetails(vxcId)

		if wait%5 == 0 {
			v.Log.Infoln("VXC is currently being provisioned. Status: ", vxcInfo.ProvisioningStatus)
		}
	}

	vxcInfo, _ = v.GetVXCDetails(vxcId)
	v.Log.Debugln("VXC waiting cycle complete. Status: ", vxcInfo.ProvisioningStatus)

	if vxcInfo.ProvisioningStatus == "LIVE" {
		return true, nil
	} else {
		if wait >= 30 {
			return false, errors.New(mega_err.ERR_VXC_PROVISION_TIMEOUT_EXCEED)
		} else {
			return false, errors.New(mega_err.ERR_VXC_NOT_LIVE)
		}
	}
}

func (v *VXC) WaitForVXCUpdated(id string, name string, rateLimit int, aEndVLAN int, bEndVLAN int) (bool, error) {
	wait := 0
	hasUpdated := false

	for !hasUpdated && wait < 30 {
		time.Sleep(30 * time.Second)
		wait++
		vxcDetails, _ := v.GetVXCDetails(id)

		if aEndVLAN == 0 {
			aEndVLAN = vxcDetails.AEndConfiguration.VLAN
		}

		if bEndVLAN == 0 {
			bEndVLAN = vxcDetails.BEndConfiguration.VLAN
		}

		if wait%5 == 0 {
			v.Log.Debugf("VXC Update in progress: Name %t; RateLimit %t; AEndVLAN %t; BEndVLAN %t\n",
				vxcDetails.Name == name,
				vxcDetails.RateLimit == rateLimit,
				vxcDetails.AEndConfiguration.VLAN == aEndVLAN,
				vxcDetails.BEndConfiguration.VLAN == bEndVLAN)
		}

		if vxcDetails.Name == name && vxcDetails.RateLimit == rateLimit && vxcDetails.AEndConfiguration.VLAN == aEndVLAN && vxcDetails.BEndConfiguration.VLAN == bEndVLAN {
			hasUpdated = true
		}
	}

	vxcDetails, _ := v.GetVXCDetails(id)
	v.Log.Debugf("VXC wait cyclecomplete: Name %t; RateLimit %t; AEndVLAN %t; BEndVLAN %t\n",
		vxcDetails.Name == name,
		vxcDetails.RateLimit == rateLimit,
		vxcDetails.AEndConfiguration.VLAN == aEndVLAN,
		vxcDetails.BEndConfiguration.VLAN == bEndVLAN)

	if wait >= 30 {
		return false, errors.New(mega_err.ERR_VXC_UPDATE_TIMEOUT_EXCEED)
	} else {
		return true, nil
	}
}
