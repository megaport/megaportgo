package megaport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

// ProductAvailabilityClientTestSuite tests the ProductAvailability service.
type ProductAvailabilityClientTestSuite struct {
	ClientTestSuite
}

func TestProductAvailabilityClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ProductAvailabilityClientTestSuite))
}

func (suite *ProductAvailabilityClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url

	suite.client.ProductAvailabilityService = NewProductAvailabilityService(suite.client)
}

func (suite *ProductAvailabilityClientTestSuite) TearDownTest() {
	suite.server.Close()
}

const testAvailabilityCompanyUID = "3f2504e0-4f89-11d3-9a0c-0305e82c3301"

func (suite *ProductAvailabilityClientTestSuite) TestListCompanyProductAvailability() {
	ctx := context.Background()
	path := "/v1/availability/companies/" + testAvailabilityCompanyUID + "/products"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"message": "Available products",
			"terms": "This data is subject to the Acceptable Use Policy",
			"data": [
				{
					"category": "PRODUCT",
					"displayName": "AWS VXC",
					"variant": "VXC_AWS",
					"markets": [
						{
							"market": "MEGAPORT_AUSTRALIA",
							"marketName": "Australia",
							"marketOwner": "Megaport (Australia) Pty Ltd",
							"countryCode": "AU",
							"source": "MARKET"
						},
						{
							"market": "MEGAPORT_US",
							"marketName": "United States",
							"marketOwner": "Megaport (USA) Inc",
							"countryCode": "US",
							"source": "OVERRIDE"
						}
					]
				}
			]
		}`)
	})

	got, err := suite.client.ProductAvailabilityService.ListCompanyProductAvailability(ctx, testAvailabilityCompanyUID)
	suite.NoError(err)
	suite.Len(got, 1)

	want := &CompanyProductAvailability{
		Category:    AvailableProductCategoryProduct,
		DisplayName: "AWS VXC",
		Variant:     "VXC_AWS",
		Markets: []CompanyProductMarket{
			{
				Market:      "MEGAPORT_AUSTRALIA",
				MarketName:  "Australia",
				MarketOwner: "Megaport (Australia) Pty Ltd",
				CountryCode: "AU",
				Source:      AvailabilitySourceMarket,
			},
			{
				Market:      "MEGAPORT_US",
				MarketName:  "United States",
				MarketOwner: "Megaport (USA) Inc",
				CountryCode: "US",
				Source:      AvailabilitySourceOverride,
			},
		},
	}
	suite.Equal(want, got[0])
}

// An empty data array is a real answer: the company can order nothing.
func (suite *ProductAvailabilityClientTestSuite) TestListCompanyProductAvailabilityEmptyData() {
	ctx := context.Background()
	path := "/v1/availability/companies/" + testAvailabilityCompanyUID + "/products"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"Available products","terms":"terms","data":[]}`)
	})

	got, err := suite.client.ProductAvailabilityService.ListCompanyProductAvailability(ctx, testAvailabilityCompanyUID)
	suite.NoError(err)
	suite.Empty(got)
}

func (suite *ProductAvailabilityClientTestSuite) TestListCompanyProductAvailabilityNullData() {
	ctx := context.Background()
	path := "/v1/availability/companies/" + testAvailabilityCompanyUID + "/products"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"Available products","terms":"terms","data":null}`)
	})

	got, err := suite.client.ProductAvailabilityService.ListCompanyProductAvailability(ctx, testAvailabilityCompanyUID)
	suite.Nil(got)
	suite.ErrorIs(err, ErrCompanyProductAvailabilityResponseEmpty)
}

func (suite *ProductAvailabilityClientTestSuite) TestListCompanyProductAvailabilityForbidden() {
	ctx := context.Background()
	path := "/v1/availability/companies/" + testAvailabilityCompanyUID + "/products"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"message":"Access denied","errorCode":"permission.denied"}`)
	})

	got, err := suite.client.ProductAvailabilityService.ListCompanyProductAvailability(ctx, testAvailabilityCompanyUID)
	suite.Nil(got)
	suite.Error(err)

	var errResp *ErrorResponse
	suite.True(errors.As(err, &errResp))
	suite.Equal(http.StatusForbidden, errResp.Response.StatusCode)
}

// A body that declares more bytes than it writes fails the decode mid-read.
func (suite *ProductAvailabilityClientTestSuite) TestListCompanyProductAvailabilityBodyReadError() {
	ctx := context.Background()
	path := "/v1/availability/companies/" + testAvailabilityCompanyUID + "/products"

	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", "4096")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message":"test-message","data":[`)
	})

	got, err := suite.client.ProductAvailabilityService.ListCompanyProductAvailability(ctx, testAvailabilityCompanyUID)
	suite.Nil(got)
	suite.Error(err)
}
