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

// The `partner` package is used to find Partner Megaports that can be used as the B-End for VXCs.
package partner

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
)

type Partner struct {
	*config.Config
}

func New(cfg *config.Config) *Partner {
	return &Partner{
		Config: cfg,
	}
}

// GetAllPartnerMegaports gets a list of all partner megaports in the Megaport Marketplace.
func (p *Partner) GetAllPartnerMegaports() ([]types.PartnerMegaport, error) {
	partnerMegaportUrl := "/v2/dropdowns/partner/megaports"

	response, resErr := p.Config.MakeAPICall("GET", partnerMegaportUrl, nil)
	isResErr, parsedResErr := p.Config.IsErrorResponse(response, &resErr, 200)

	if isResErr {
		return nil, parsedResErr
	}
	defer response.Body.Close()

	body, fileErr := ioutil.ReadAll(response.Body)

	if fileErr != nil {
		return []types.PartnerMegaport{}, fileErr
	}

	partnerMegaportResponse := types.PartnerMegaportResponse{}
	unmarshalErr := json.Unmarshal(body, &partnerMegaportResponse)

	if unmarshalErr != nil {
		return []types.PartnerMegaport{}, unmarshalErr
	}

	return partnerMegaportResponse.Data, nil
}

func (p *Partner) FilterPartnerMegaportByProductName(partnerMegaports *[]types.PartnerMegaport, productName string, exactMatch bool) error {
	existingMegaports := *partnerMegaports
	var filteredMegaports []types.PartnerMegaport

	for i := 0; i < len(existingMegaports); i++ {
		match := false

		if productName != "" {
			if exactMatch { // Exact Match
				if productName == existingMegaports[i].ProductName {
					match = true
				}
			} else {
				if fuzzy.Match(productName, existingMegaports[i].ProductName) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match && existingMegaports[i].VXCPermitted {
			filteredMegaports = append(filteredMegaports, existingMegaports[i])
		}
	}

	*partnerMegaports = filteredMegaports

	if len(*partnerMegaports) == 0 {
		return errors.New(mega_err.ERR_PARTNER_PORT_NO_RESULTS)
	} else {
		return nil
	}
}

func (p *Partner) FilterPartnerMegaportByConnectType(partnerMegaports *[]types.PartnerMegaport, connectType string, exactMatch bool) error {
	existingMegaports := *partnerMegaports
	var filteredMegaports []types.PartnerMegaport

	for i := 0; i < len(existingMegaports); i++ {
		match := false

		if connectType != "" {
			if exactMatch { // Exact Match
				if connectType == existingMegaports[i].ConnectType {
					match = true
				}
			} else {
				if fuzzy.Match(connectType, existingMegaports[i].ConnectType) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match && existingMegaports[i].VXCPermitted {
			filteredMegaports = append(filteredMegaports, existingMegaports[i])
		}
	}

	*partnerMegaports = filteredMegaports

	if len(*partnerMegaports) == 0 {
		return errors.New(mega_err.ERR_PARTNER_PORT_NO_RESULTS)
	} else {
		return nil
	}
}

func (p *Partner) FilterPartnerMegaportByCompanyName(partnerMegaports *[]types.PartnerMegaport, companyName string, exactMatch bool) error {
	existingMegaports := *partnerMegaports
	var filteredMegaports []types.PartnerMegaport

	for i := 0; i < len(existingMegaports); i++ {
		match := false

		if companyName != "" {
			if exactMatch { // Exact Match
				if companyName == existingMegaports[i].CompanyName {
					match = true
				}
			} else {
				if fuzzy.Match(companyName, existingMegaports[i].CompanyName) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match && existingMegaports[i].VXCPermitted {
			filteredMegaports = append(filteredMegaports, existingMegaports[i])
		}
	}

	*partnerMegaports = filteredMegaports

	if len(*partnerMegaports) == 0 {
		return errors.New(mega_err.ERR_PARTNER_PORT_NO_RESULTS)
	} else {
		return nil
	}
}

func (p *Partner) FilterPartnerMegaportByLocationId(partnerMegaports *[]types.PartnerMegaport, locationId int) error {
	existingMegaports := *partnerMegaports
	var filteredMegaports []types.PartnerMegaport

	for i := 0; i < len(existingMegaports); i++ {
		if locationId >= 0 {
			if locationId == existingMegaports[i].LocationId && existingMegaports[i].VXCPermitted {
				filteredMegaports = append(filteredMegaports, existingMegaports[i])
			}
		} else {
			filteredMegaports = append(filteredMegaports, existingMegaports[i])
		}
	}

	*partnerMegaports = filteredMegaports

	if len(*partnerMegaports) == 0 {
		return errors.New(mega_err.ERR_PARTNER_PORT_NO_RESULTS)
	} else {
		return nil
	}
}
