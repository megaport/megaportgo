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
package mve

import (
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/service/product"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
)

// MVE provides an interface for configuring and ordering MVE instances.
type MVE struct {
	*config.Config
	product *product.Product
}

// New returns a new MVE ready for configuration.
func New(cfg *config.Config) *MVE {
	return &MVE{cfg, product.New(cfg)}
}

// BuyMVE purchases an MVE.
func (m *MVE) BuyMVE(locationID int, name string, term int, config map[string]interface{}, vnics []*types.MVENetworkInterface) (string, error) {
	// Create a default vNIC if none specified.
	if len(vnics) == 0 {
		vnics = []*types.MVENetworkInterface{{Description: "Data Plane"}}
	}

	if term != 1 && term != 12 && term != 24 && term != 36 {
		return "", errors.New(mega_err.ERR_TERM_NOT_VALID)
	}

	order := []*types.MVEOrderConfig{{
		LocationID:        locationID,
		Name:              name,
		Term:              term,
		ProductType:       strings.ToUpper(types.PRODUCT_MVE),
		NetworkInterfaces: vnics,
		VendorConfig:      config,
	}}

	requestBody, err := json.Marshal(order)
	if err != nil {
		return "", err
	}

	body, err := m.product.ExecuteOrder(&requestBody)
	if err != nil {
		return "", err
	}

	orderInfo := types.MVEOrderResponse{}
	if err := json.Unmarshal(*body, &orderInfo); err != nil {
		return "", err
	}

	return orderInfo.Data[0].TechnicalServiceUID, nil
}

// GetMVEDetails returns the details of a configured MVE.
func (m *MVE) GetMVEDetails(uid string) (*types.MVE, error) {
	url := "/v2/product/" + uid
	res, err := m.Config.MakeAPICall("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	mveDetails := types.MVEResponse{}
	if err := json.Unmarshal(body, &mveDetails); err != nil {
		return nil, err
	}

	return &mveDetails.Data, nil
}

func (m *MVE) ModifyMVE(uid string, name string) (bool, error) {
	return m.product.ModifyProduct(uid, types.PRODUCT_MVE, name, "", false)
}

func (m *MVE) DeleteMVE(uid string) (bool, error) {
	return m.product.DeleteProduct(uid, true)
}

func (m *MVE) WaitForMVEProvisioning(uid string) (bool, error) {
	// Try for ~5mins.
	for i := 0; i < 30; i++ {
		details, err := m.GetMVEDetails(uid)
		if err != nil {
			return false, err
		}

		if slices.Contains(shared.SERVICE_STATE_READY, details.ProvisioningStatus) {
			return true, nil
		}

		// Wrong status, wait a bit and try again.
		m.Log.Debugf("MVE status is %q - waiting", details.ProvisioningStatus)
		time.Sleep(10 * time.Second)
	}

	return false, errors.New(mega_err.ERR_MVE_PROVISION_TIMEOUT_EXCEED)
}
