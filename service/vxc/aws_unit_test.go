// +build unit

// This file is present as a "Guide" for the seperation of unit and
// Integration testing.

package vxc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

/*********************

  CONSTANTS AND SUCH

*********************/

const (
	TEST_1_REFERENCE_STRING = `[{"associatedVxcs":[{"productName":"VXC MCR to AWSHC from GO","rateLimit":500,"aEnd":{"vlan":0,"partnerConfig":{"interfaces":[{"ipAddresses":["10.192.0.25/29"],"bfd":{"txInterval":300,"rxInterval":300,"multiplier":3},"bgpConnections":[{"peerAsn":64512,"localIpAddress":"10.192.0.25","peerIpAddress":"10.192.0.26","password":"cnn6eaeaETSjvjvjvjv","shutdown":false,"description":"BGP with MED and BFD enabled","medIn":100,"medOut":100,"bfdEnabled":true}]}]}},"bEnd":{"productUid":"b2e0b6b8-2943-4c44-8a07-9ec13060afb2","partnerConfig":{"connectType":"AWSHC","type":"private","ownerAccount":"684021030471"}}}],"productUid":"mcr-id-here"}]`
)

func SUCCESS_RESPONSE_MOCK(req *http.Request) *http.Response {

	mockResponseBody := types.VXCOrderResponse{
		Message: "mock message",
		Terms:   "mock terms",
		Data: []types.VXCOrderConfirmation{
			{TechnicalServiceUID: "mock-uid-1234-5678"},
		},
	}

	mockResponseBodyJson, _ := json.Marshal(mockResponseBody)

	// return a mock response
	return &http.Response{
		Body:          ioutil.NopCloser(bytes.NewBufferString(string(mockResponseBodyJson))),
		StatusCode:    200,
		Status:        "200 OK",
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		ContentLength: int64(len(string(mockResponseBodyJson))),
		Request:       req,
		Header:        make(http.Header, 0),
	}

}

/*********************

    MOCKING HERE

*********************/

// Mock HttpClient Implementation
type MockHttpClient struct {
	mock.Mock
}

// implememen Do method of httpClien
func (h *MockHttpClient) Do(req *http.Request) (*http.Response, error) {

	// extract body
	body, err := ioutil.ReadAll(req.Body)

	// break out error
	if err != nil {
		var mockError error = err
		return nil, mockError
	}

	// Compare prepared payload to expected payload
	if string(body) != TEST_1_REFERENCE_STRING {
		var mockError error = errors.New("Request body did not match")
		return nil, mockError
	}

	return SUCCESS_RESPONSE_MOCK(req), nil
}

/***************

TESTS START HERE

***************/

func Test_VXC_MCR_AWS(t *testing.T) {

	mockClient := new(MockHttpClient)

	cfg := config.Config{
		Log:      config.NewDefaultLogger(),
		Endpoint: "https://api-staging.megaport.com/",
		Client:   mockClient,
	}

	cfg.SessionToken = "fake-token"

	vxcService := New(&cfg)

	vxcAmazonProductUidHC := "b2e0b6b8-2943-4c44-8a07-9ec13060afb2"
	vxcName := "VXC MCR to AWSHC from GO"
	vxcRteLimit := 500

	partnerConfigInterface := types.PartnerConfigInterface{
		IpAddresses: []string{"10.192.0.25/29"},
		Bfd: types.BfdConfig{
			TxInterval: 300,
			RxInterval: 300,
			Multiplier: 3,
		},
		BgpConnections: []types.BgpConnectionConfig{
			{
				PeerAsn:        64512,
				LocalIpAddress: "10.192.0.25",
				PeerIpAddress:  "10.192.0.26",
				Password:       "cnn6eaeaETSjvjvjvjv",
				Shutdown:       false,
				Description:    "BGP with MED and BFD enabled",
				MedIn:          100,
				MedOut:         100,
				BfdEnabled:     true,
			},
		},
	}

	aEndConfiguration := types.VXCOrderAEndConfiguration{
		VLAN: 0,
		PartnerConfig: types.VXCOrderAEndPartnerConfig{
			Interfaces: []types.PartnerConfigInterface{
				partnerConfigInterface,
			},
		},
	}

	bEndConfiguration := types.AWSVXCOrderBEndConfiguration{
		ProductUID: vxcAmazonProductUidHC,
		PartnerConfig: types.AWSVXCOrderBEndPartnerConfig{
			ConnectType:  "AWSHC",
			Type:         "private",
			OwnerAccount: "684021030471",
		},
	}

	vxcId, vxcError := vxcService.BuyAWSVXC(
		"mcr-id-here",
		vxcName,
		vxcRteLimit,
		aEndConfiguration,
		bEndConfiguration,
	)

	assert.True(t, vxcId != "", vxcError)

}
