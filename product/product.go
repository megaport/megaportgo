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
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
	"io/ioutil"
)

// ExecuteOrder executes an order against the Megaport API.
func ExecuteOrder(requestBody *[]byte) (*[]byte, error) {
	url := "/v2/networkdesign/buy"
	response, resErr := shared.MakeAPICall("POST", url, *requestBody)
	defer response.Body.Close()

	isError, parsedError := shared.IsErrorResponse(response, &resErr, 200)

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
func DeleteProduct(id string, deleteNow bool) (bool, error) {
	var action string

	if deleteNow {
		action = "CANCEL_NOW"
	} else {
		action = "CANCEL"
	}

	url := "/v2/product/" + id + "/action/" + action
	response, err := shared.MakeAPICall("POST", url, nil)
	defer response.Body.Close()

	isError, errorMessage := shared.IsErrorResponse(response, &err, 200)

	if isError {
		return false, errorMessage
	} else {
		return true, nil
	}
}

// RestoreProduct will re-enable a Product if a product has been scheduled for deletion.
func RestoreProduct(id string) (bool, error) {
	url := "/v2/product/" + id + "/action/UN_CANCEL"
	response, err := shared.MakeAPICall("POST", url, nil)
	defer response.Body.Close()

	isError, errorMessage := shared.IsErrorResponse(response, &err, 200)

	if isError {
		return false, errorMessage
	} else {
		return true, nil
	}
}

// ModifyProduct modifies a product. The available fields to modify are Name, Cost Centre, and Marketplace Visibility.
func ModifyProduct(productId string, productType string, name string, costCentre string, marketplaceVisibility bool) (bool, error) {
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

		updateResponse, err := shared.MakeAPICall("PUT", url, []byte(body))
		isResErr, compiledResErr := shared.IsErrorResponse(updateResponse, &err, 200)

		if isResErr {
			return false, compiledResErr
		} else {
			return true, nil
		}
	} else {
		return false, errors.New(mega_err.ERR_WRONG_PRODUCT_MODIFY)
	}

}

func ManageProductLock(productId string, shouldLock bool) (bool, error) {
	verb := "POST"

	if !shouldLock {
		verb = "DELETE"
	}
	url := fmt.Sprintf("/v2/product/%s/lock", productId)
	lockResponse, err := shared.MakeAPICall(verb, url, nil)
	isResErr, compiledResErr := shared.IsErrorResponse(lockResponse, &err, 200)
	if isResErr {
		return false, compiledResErr
	} else {
		return true, nil
	}
}
