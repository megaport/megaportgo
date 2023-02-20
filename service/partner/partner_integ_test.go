//go:build integration
// +build integration

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

package partner

import (
	"os"
	"testing"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/service/authentication"
	"github.com/megaport/megaportgo/service/location"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/assert"
)

const (
	MEGAPORTURL = "https://api-staging.megaport.com/"
)

var logger *config.DefaultLogger
var cfg config.Config

func TestMain(m *testing.M) {
	logger = config.NewDefaultLogger()
	logger.SetLevel(config.DebugLevel)

	username := os.Getenv("MEGAPORT_USERNAME")
	password := os.Getenv("MEGAPORT_PASSWORD")
	otp := os.Getenv("MEGAPORT_MFA_OTP_KEY")
	logLevel := os.Getenv("LOG_LEVEL")

	if logLevel != "" {
		logger.SetLevel(config.StringToLogLevel(logLevel))
	}

	if username == "" {
		logger.Error("MEGAPORT_USERNAME environment variable not set.")
		os.Exit(1)
	}

	if password == "" {
		logger.Error("MEGAPORT_PASSWORD environment variable not set.")
		os.Exit(1)
	}

	cfg = config.Config{
		Log:      logger,
		Endpoint: MEGAPORTURL,
	}

	auth := authentication.New(&cfg, username, password, otp)
	token, loginErr := auth.Login()

	if loginErr != nil {
		logger.Errorf("LoginError: %s", loginErr.Error())
	}

	cfg.SessionToken = token
	os.Exit(m.Run())
}

func TestGetAllPartnerMegaports(t *testing.T) {
	// assert := assert.New(t)
	partner := New(&cfg)

	// Make sure that an id with no record returns an error as expected.
	partner.GetAllPartnerMegaports()
	// TODO: figure out a condition for this.
}

func TestFilterPartnerMegaportByCompanyName(t *testing.T) {
	assert := assert.New(t)
	partner := New(&cfg)

	partnerMegaports, _ := partner.GetAllPartnerMegaports()
	partner.FilterPartnerMegaportByCompanyName(&partnerMegaports, "AWS", true)

	for i := 0; i < len(partnerMegaports); i++ {
		assert.Equal(partnerMegaports[i].CompanyName, "AWS")
	}
}

func TestFilterPartnerMegaportByLocationId(t *testing.T) {
	assert := assert.New(t)
	partner := New(&cfg)
	loc := location.New(&cfg)

	partnerMegaports, _ := partner.GetAllPartnerMegaports()
	location, _ := loc.GetLocationByName("Equinix SY3")
	partner.FilterPartnerMegaportByLocationId(&partnerMegaports, location.ID)
	assert.Greater(len(partnerMegaports), 0)

	for i := 0; i < len(partnerMegaports); i++ {
		assert.Equal(partnerMegaports[i].LocationId, location.ID)
	}
}

func TestFilterPartnerMegaportByConnectType(t *testing.T) {
	assert := assert.New(t)
	partner := New(&cfg)

	partnerMegaports, _ := partner.GetAllPartnerMegaports()
	partner.FilterPartnerMegaportByConnectType(&partnerMegaports, types.CONNECT_TYPE_AWS_VIF, true)
	assert.Greater(len(partnerMegaports), 0)

	for i := 0; i < len(partnerMegaports); i++ {
		assert.Equal(partnerMegaports[i].ConnectType, types.CONNECT_TYPE_AWS_VIF)
	}

	partnerMegaports, _ = partner.GetAllPartnerMegaports()
	partner.FilterPartnerMegaportByConnectType(&partnerMegaports, types.CONNECT_TYPE_AWS_VIF, false)
	assert.Greater(len(partnerMegaports), 0)

	for i := 0; i < len(partnerMegaports); i++ {
		assert.Contains(partnerMegaports[i].ConnectType, types.CONNECT_TYPE_AWS_VIF)
	}

	partnerMegaports, _ = partner.GetAllPartnerMegaports()
	partner.FilterPartnerMegaportByConnectType(&partnerMegaports, types.CONNECT_TYPE_AWS_HOSTED_CONNECTION, true)
	assert.Greater(len(partnerMegaports), 0)

	for i := 0; i < len(partnerMegaports); i++ {
		assert.Equal(partnerMegaports[i].ConnectType, types.CONNECT_TYPE_AWS_HOSTED_CONNECTION)
	}
}

func TestFilterPartnerMegaportByProductName(t *testing.T) {
	partner := New(&cfg)

	partnerMegaports, _ := partner.GetAllPartnerMegaports()
	productName := "Asia Pacific (Sydney) (ap-southeast-2)"
	partner.FilterPartnerMegaportByProductName(&partnerMegaports, productName, true)

	for i := 0; i < len(partnerMegaports); i++ {
		logger.Infof("Item found. ProductName: %s", partnerMegaports[i].ProductName)
	}
}

func TestFilterPartnerMegaportByDiversityZone(t *testing.T) {
	assert := assert.New(t)
	partner := New(&cfg)

	partnerMegaports, _ := partner.GetAllPartnerMegaports()
	partner.FilterPartnerMegaportByDiversityZone(&partnerMegaports, "red", true)

	for i := 0; i < len(partnerMegaports); i++ {
		assert.Equal(partnerMegaports[i].DiversityZone, "red")
	}
}

/*
func TestGetPartnerMegaportByFilter(t *testing.T) {
	// 52dfc422-9041-4a16-b040-f03c795a3e01
	GetPartnerMegaportByFilter("+AWS", "Sydney", "+AWS", -1)
}*/
