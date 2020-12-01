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

type Location struct {
	Name             string                 `json:"name"`
	Country          string                 `json:"country"`
	LiveDate         int                    `json:"liveDate"`
	SiteCode         string                 `json:"siteCode"`
	NetworkRegion    string                 `json:"networkRegion"`
	Address          map[string]string      `json:"address"`
	Campus           string                 `json:"campus"`
	Latitude         float64                `json:"latitude"`
	Longitude        float64                `json:"longitude"`
	Products         map[string]interface{} `json:"products"`
	Market           string                 `json:"market"`
	Metro            string                 `json:"metro"`
	VRouterAvailable bool                   `json:"vRouterAvailable"`
	ID               int                    `json:"id"`
	Status           string                 `json:"status"`
}

type Country struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Prefix    string `json:"prefix"`
	SiteCount int    `json:"siteCount"`
}
