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

type PartnerMegaport struct {
	ConnectType  string `json:"connectType"`
	ProductUID   string `json:"productUid"`
	ProductName  string `json:"title"`
	CompanyUID   string `json:"companyUid"`
	CompanyName  string `json:"companyName"`
	LocationId   int    `json:"locationId"`
	Speed        int    `json:"speed"`
	Rank         int    `json:"rank"`
	VXCPermitted bool   `json:"vxcPermitted"`
}
