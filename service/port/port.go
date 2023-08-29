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

// buyPort orders a Port.
func (p *Port) buyPort(name string, term int, portSpeed int, locationId int, market string, isLAG bool, lagCount int, isPrivate bool, diversityZone string) (string, error) {
	var buyOrder []types.PortOrder
	var portConfig types.PortOrderConfig

	if term != 1 && term != 12 && term != 24 && term != 36 {
		return "", errors.New(mega_err.ERR_TERM_NOT_VALID)
	}

	switch diversityZone {
	case "red":
		portConfig.DiversityZone = "red"
	case "blue":
		portConfig.DiversityZone = "blue"
	case "any":
		// Continue with no zone configured
	default:
		return "", errors.New(mega_err.ERR_ZONE_NOT_VALID)
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
				Config:                portConfig,
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
				Config:                portConfig,
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

// BuyPort orders a Port or LAG with unspecified zone. Used for compatability with older versions of megaportgo.
func (p *Port) BuyPort(name string, term int, portSpeed int, locationId int, market string, isLAG bool, lagCount int, isPrivate bool) (string, error) {
	return p.buyPort(name, term, portSpeed, locationId, market, isLAG, lagCount, isPrivate, "any")
}

// BuySinglePort orders a single Port with unspecified zone.
func (p *Port) BuySinglePort(name string, term int, portSpeed int, locationId int, market string, isPrivate bool) (string, error) {
	return p.buyPort(name, term, portSpeed, locationId, market, false, 0, isPrivate, "any")
}

// BuyZonedSinglePort orders a single Port in the requested zone.
func (p *Port) BuyZonedSinglePort(name string, term int, portSpeed int, locationId int, market string, isPrivate bool, diversityZone string) (string, error) {
	return p.buyPort(name, term, portSpeed, locationId, market, false, 0, isPrivate, diversityZone)
}

// BuyLAGPort orders a LAG Port/s with unspecified zone.
func (p *Port) BuyLAGPort(name string, term int, portSpeed int, locationId int, market string, lagCount int, isPrivate bool) (string, error) {
	return p.buyPort(name, term, portSpeed, locationId, market, true, lagCount, isPrivate, "any")
}

// BuyZonedLAGPort orders a LAG Port/s in the requested zone.
func (p *Port) BuyZonedLAGPort(name string, term int, portSpeed int, locationId int, market string, lagCount int, isPrivate bool, diversityZone string) (string, error) {
	return p.buyPort(name, term, portSpeed, locationId, market, true, lagCount, isPrivate, diversityZone)
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
	portInfo, _ := p.GetPortDetails(portId)
	wait := 0

	p.Log.Debugln("Waiting for port status transition.")
	for portInfo.ProvisioningStatus != "CONFIGURED" && portInfo.ProvisioningStatus != "LIVE" && wait < 30 {
		time.Sleep(30 * time.Second)
		wait++
		portInfo, _ = p.GetPortDetails(portId)

		if wait%5 == 0 {
			p.Log.Debugln("Port is currently being provisioned. Status: ", portInfo.ProvisioningStatus)
		}
	}

	portInfo, _ = p.GetPortDetails(portId)
	p.Log.Debugln("Port waiting cycle complete. Status:", portInfo.ProvisioningStatus)

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
