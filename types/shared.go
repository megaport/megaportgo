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

package types

type PortInterface struct {
	Demarcation  string `json:"demarcation"`
	Description  string `json:"description"`
	ID           int    `json:"id"`
	LOATemplate  string `json:"loa_template"`
	Media        string `json:"media"`
	Name         string `json:"name"`
	PortSpeed    int    `json:"port_speed"`
	ResourceName string `json:"resource_name"`
	ResourceType string `json:"resource_type"`
	Up           int    `json:"up"`
}

const APPLICATION_SHORT_NAME = "mpt"

const PRODUCT_MEGAPORT = "megaport"
const PRODUCT_VXC = "vxc"
const PRODUCT_MCR = "mcr2"
const PRODUCT_MVE = "mve"
const PRODUCT_IX = "ix"

const STATUS_DECOMMISSIONED string = "DECOMMISSIONED"
const STATUS_CANCELLED string = "CANCELLED"
const SINGLE_PORT string = "Single"
const LAG_PORT string = "LAG"
const CONNECT_TYPE_AWS_VIF string = "AWS"
const CONNECT_TYPE_AWS_HOSTED_CONNECTION string = "AWSHC"

const MODIFY_NAME string = "NAME"
const MODIFY_COST_CENTRE = "COST_CENTRE"
const MODIFY_MARKETPLACE_VISIBILITY string = "MARKETPLACE_VISIBILITY"
const MODIFY_RATE_LIMIT = "RATE_LIMIT"
const MODIFY_A_END_VLAN = "A_VLAN"
const MODIFY_B_END_VLAN = "B_VLAN"
