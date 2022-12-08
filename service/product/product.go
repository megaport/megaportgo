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

// The `product` package contains shared code for the management of MCR, Ports, and VXCs.
package product

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
)

type Product struct {
	*config.Config
}

func New(cfg *config.Config) *Product {
	return &Product{
		Config: cfg,
	}
}

// ExecuteOrder executes an order against the Megaport API.
func (p *Product) ExecuteOrder(requestBody *[]byte) (*[]byte, error) {
	url := "/v2/networkdesign/buy"

	response, resErr := p.Config.MakeAPICall("POST", url, *requestBody)

	if response != nil {
		defer response.Body.Close()
	}

	isError, parsedError := p.Config.IsErrorResponse(response, &resErr, 200)

	if isError {
		return nil, parsedError
	}

	body, fileErr := ioutil.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	return &body, nil
}

// DeleteProduct is responsible for either scheduling a product for deletion "CANCEL" or deleting a product immediately
// "CANCEL_NOW".
func (p *Product) DeleteProduct(id string, deleteNow bool) (bool, error) {
	var action string

	if deleteNow {
		action = "CANCEL_NOW"
	} else {
		action = "CANCEL"
	}

	url := "/v2/product/" + id + "/action/" + action
	response, err := p.Config.MakeAPICall("POST", url, nil)
	defer response.Body.Close()

	isError, errorMessage := p.Config.IsErrorResponse(response, &err, 200)

	if isError {
		return false, errorMessage
	} else {
		return true, nil
	}
}

// RestoreProduct will re-enable a Product if a product has been scheduled for deletion.
func (p *Product) RestoreProduct(id string) (bool, error) {
	url := "/v2/product/" + id + "/action/UN_CANCEL"
	response, err := p.Config.MakeAPICall("POST", url, nil)
	defer response.Body.Close()

	isError, errorMessage := p.Config.IsErrorResponse(response, &err, 200)

	if isError {
		return false, errorMessage
	} else {
		return true, nil
	}
}

// ModifyProduct modifies a product. The available fields to modify are Name, Cost Centre, and Marketplace Visibility.
func (p *Product) ModifyProduct(productId string, productType string, name string, costCentre string, marketplaceVisibility bool) (bool, error) {
	if productType == types.PRODUCT_MEGAPORT || productType == types.PRODUCT_MCR {
		update := types.ProductUpdate{
			Name:                 name,
			CostCentre:           costCentre,
			MarketplaceVisbility: marketplaceVisibility,
		}
		url := fmt.Sprintf("/v2/product/%s/%s", productType, productId)

		body, marshalErr := json.Marshal(update)

		if marshalErr != nil {
			return false, marshalErr
		}

		updateResponse, err := p.Config.MakeAPICall("PUT", url, []byte(body))
		isResErr, compiledResErr := p.Config.IsErrorResponse(updateResponse, &err, 200)

		if isResErr {
			return false, compiledResErr
		} else {
			return true, nil
		}
	} else {
		return false, errors.New(mega_err.ERR_WRONG_PRODUCT_MODIFY)
	}

}

func (p *Product) ManageProductLock(productId string, shouldLock bool) (bool, error) {
	verb := "POST"

	if !shouldLock {
		verb = "DELETE"
	}
	url := fmt.Sprintf("/v2/product/%s/lock", productId)
	lockResponse, err := p.Config.MakeAPICall(verb, url, nil)
	isResErr, compiledResErr := p.Config.IsErrorResponse(lockResponse, &err, 200)
	if isResErr {
		return false, compiledResErr
	} else {
		return true, nil
	}
}

// GetMCRPrefixFilterLists returns prefix filter lists for the specified MCR2.
func (p *Product) GetMCRPrefixFilterLists(id string) ([]types.PrefixFilterList, error) {
	url := "/v2/product/mcr2/" + id + "/prefixLists?"

	response, err := p.Config.MakeAPICall("GET", url, nil)
	isError, errorMessage := p.Config.IsErrorResponse(response, &err, 200)

	if isError {
		return nil, errorMessage
	}
	defer response.Body.Close()

	body, fileErr := ioutil.ReadAll(response.Body)
	if fileErr != nil {
		return []types.PrefixFilterList{}, fileErr
	}

	prefixFilterList := types.MCRPrefixFilterListResponse{}
	unmarshalErr := json.Unmarshal(body, &prefixFilterList)

	if unmarshalErr != nil {
		return []types.PrefixFilterList{}, unmarshalErr
	}

	return prefixFilterList.Data, nil
}

// CreateMCRPrefixFilterList will create an MCR2 product prefix filter list.
func (p *Product) CreateMCRPrefixFilterList(id string, prefixFilterList types.MCRPrefixFilterList) (bool, error) {
	url := "/v2/product/mcr2/" + id + "/prefixList"

	body, marshalErr := json.Marshal(prefixFilterList)
	if marshalErr != nil {
		return false, marshalErr
	}

	response, err := p.Config.MakeAPICall("POST", url, []byte(body))
	if response != nil {
		defer response.Body.Close()
	}

	isError, errorMessage := p.Config.IsErrorResponse(response, &err, 200)
	if isError {
		return false, errorMessage
	} else {
		return true, nil
	}
}
