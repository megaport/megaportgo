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
	"io/ioutil"
	"slices"
	"time"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/service/product"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
)

const MODIFY_NAME string = "NAME"
const MODIFY_COST_CENTRE = "COST_CENTRE"
const MODIFY_MARKETPLACE_VISIBILITY string = "MARKETPLACE_VISIBILITY"

type Port struct {
	*config.Config
	product *product.Product
}

func New(cfg *config.Config) *Port {
	return &Port{
		Config:  cfg,
		product: product.New(cfg),
	}
}

// BuyPort orders a Port.
func (p *Port) BuyPort(name string, term int, portSpeed int, locationId int, market string, isLAG bool, lagCount int, isPrivate bool) (string, error) {
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
	responseBody, responseErr := p.product.ExecuteOrder(&requestBody)

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
func (p *Port) BuySinglePort(name string, term int, portSpeed int, locationId int, market string, isPrivate bool) (string, error) {
	return p.BuyPort(name, term, portSpeed, locationId, market, false, 0, isPrivate)
}

// BuyPort orders a LAG Port. Same as BuyPort, with isLag set to true.
func (p *Port) BuyLAGPort(name string, term int, portSpeed int, locationId int, market string, lagCount int, isPrivate bool) (string, error) {
	return p.BuyPort(name, term, portSpeed, locationId, market, true, lagCount, isPrivate)
}

func (p *Port) GetPortDetails(id string) (types.Port, error) {
	url := "/v2/product/" + id
	response, err := p.Config.MakeAPICall("GET", url, nil)
	defer response.Body.Close()

	isError, parsedError := p.Config.IsErrorResponse(response, &err, 200)

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

type ParsedProductsResponse struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    []interface{} `json:"data"`
}

func (p *Port) GetPorts() ([]types.Port, error) {
	url := "/v2/products"
	response, err := p.Config.MakeAPICall("GET", url, nil)
	defer response.Body.Close()

	isError, parsedError := p.Config.IsErrorResponse(response, &err, 200)

	if isError {
		return []types.Port{}, parsedError
	}

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return []types.Port{}, fileErr
	}

	parsed := ParsedProductsResponse{}

	unmarshalErr := json.Unmarshal(body, &parsed)

	if unmarshalErr != nil {
		return []types.Port{}, unmarshalErr
	}

	var ports []types.Port

	for _, unmarshaledData := range parsed.Data {
		// The products query response will likely contain non-port objects.  As a result
		// we need to initially Unmarshal as ParsedProductsResponse so that we may iterate
		// over the entries in Data then re-Marshal those entries so that we may Unmarshal
		// them as Port (and `continue` where that doesn't work).  We could write a custom
		// deserializer to avoid this but that is a lot of work for a performance
		// optimization which is likely irrelevant in practice.
		// Unfortunately I know of no better (maintainable) method of making this work.
		remarshaled, err := json.Marshal(unmarshaledData)
		if err != nil {
			p.Log.Debugln("Could not remarshal %v as port.", err.Error())
			continue
		}
		port := types.Port{}
		unmarshalErr = json.Unmarshal(remarshaled, &port)
		if unmarshalErr != nil {
			p.Log.Debugln("Could not unmarshal %v as port.", unmarshalErr.Error())
			continue
		}
		ports = append(ports, port)
	}

	return ports, nil
}

func (p *Port) ModifyPort(portId string, name string, costCentre string, marketplaceVisibility bool) (bool, error) {
	return p.product.ModifyProduct(portId, types.PRODUCT_MEGAPORT, name, costCentre, marketplaceVisibility)
}

func (p *Port) DeletePort(id string, deleteNow bool) (bool, error) {
	return p.product.DeleteProduct(id, deleteNow)
}

func (p *Port) RestorePort(id string) (bool, error) {
	return p.product.RestoreProduct(id)
}

// TODO: Tests for locking.
func (p *Port) LockPort(id string) (bool, error) {
	portInfo, _ := p.GetPortDetails(id)
	if !portInfo.Locked {
		return p.product.ManageProductLock(id, true)
	} else {
		return true, errors.New(mega_err.ERR_PORT_ALREADY_LOCKED)
	}
}

func (p *Port) UnlockPort(id string) (bool, error) {
	portInfo, _ := p.GetPortDetails(id)
	if portInfo.Locked {
		return p.product.ManageProductLock(id, false)
	} else {
		return true, errors.New(mega_err.ERR_PORT_NOT_LOCKED)
	}
}

func (p *Port) WaitForPortProvisioning(portId string) (bool, error) {
	// Try for ~5mins.
	for i := 0; i < 30; i++ {
		details, err := p.GetPortDetails(portId)
		if err != nil {
			return false, err
		}

		if slices.Contains(shared.SERVICE_STATE_READY, details.ProvisioningStatus) {
			return true, nil
		}

		// Wrong status, wait a bit and try again.
		p.Log.Debugf("Port status is currently %q - waiting", details.ProvisioningStatus)
		time.Sleep(10 * time.Second)
	}

	return false, errors.New(mega_err.ERR_PORT_PROVISION_TIMEOUT_EXCEED)
}
