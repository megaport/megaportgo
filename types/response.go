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

type GenericResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    map[string]interface{} `json:"data"`
}

type ErrorResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    string `json:"data"`
}

type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    struct {
		Permissions map[string][]string `json:"permissions"`
		OAuthToken  struct {
			AccessToken string `json:"accessToken"`
			ExpiresIn   int
		}
	}
}

type LocationResponse struct {
	Message string     `json:"message"`
	Terms   string     `json:"terms"`
	Data    []Location `json:"data"`
}

type PortOrderResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []PortOrderConfirmation `json:"data"`
}

type PortResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    Port   `json:"data"`
}

type VXCOrderResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    []VXCOrderConfirmation `json:"data"`
}

type VXCResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    VXC    `json:"data"`
}

type PartnerMegaportResponse struct {
	Message string            `json:"message"`
	Terms   string            `json:"terms"`
	Data    []PartnerMegaport `json:"data"`
}

type MCROrderResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    []MCROrderConfirmation `json:"data"`
}

type MCRResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    MCR    `json:"data"`
}

type MVEOrderResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    []MVEOrderConfirmation `json:"data"`
}

type MVEResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    MVE    `json:"data"`
}

type PrefixFilterList struct {
	Id            int    `json:"id"`
	Description   string `json:"description"`
	AddressFamily string `json:"addressFamily"`
}

type MCRPrefixFilterListResponse struct {
	Message string             `json:"message"`
	Terms   string             `json:"terms"`
	Data    []PrefixFilterList `json:"data"`
}

type PartnerLookupResponse struct {
	Message string        `json:"message"`
	Data    PartnerLookup `json:"data"`
	Terms   string        `json:"terms"`
}

type CountryResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    []CountryInnerResponse `json:"data"`
}

type CountryInnerResponse struct {
	Countries     []Country `json:"countries"`
	NetworkRegion string    `json:"networkRegion"`
}
