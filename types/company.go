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

type Market struct {
	Currency               string `json:"currencyEnum"`
	Language               string `json:"language"`
	CompanyLegalIdentifier string `json:"companyLegalIdentifier"`
	CompanyLegalName       string `json:"companyLegalName"`
	BillingContactName     string `json:"billingContactName"`
	BillingContactPhone    string `json:"billingContactPhone"`
	BillingContactEmail    string `json:"billingContactEmail"`
	AddressLine1           string `json:"address1"`
	AddressLine2           string `json:"address2"`
	City                   string `json:"city"`
	State                  string `json:"state"`
	Postcode               string `json:"postcode"`
	Country                string `json:"country"`
	PONumber               string `json:"yourPoNumber"`
	TaxNumber              string `json:"taxNumber"`
	FirstPartyID           int    `json:"firstPartyId"`
}

type CompanyEnablement struct {
	TradingName string `json:"tradingName"`
}
