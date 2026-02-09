package megaport

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

// uid and names for testing
var (
	companyUid         = "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	companyName        = "Test Company"
	companyUid2        = "2111de9a-6211-40e3-8e6a-65ab106d01f4"
	companyName2       = "Other Company"
	productUid         = "36b3f68e-2f54-4331-bf94-f8984449365f"
	productUid2        = "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	productUid3        = "91ededc2-473f-4a30-ad24-0703c7f35e50"
	productUid4        = "9a05f787-8166-4470-94f1-4906db86f698"
	partnerMegaportUrl = "/v2/dropdowns/partner/megaports"
)

// PartnerClientTestSuite tests the Partner Service.
type PartnerClientTestSuite struct {
	ClientTestSuite
}

func TestPartnerClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PartnerClientTestSuite))
}

func (suite *PartnerClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *PartnerClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// AWS Partner
var awsPartner = &PartnerMegaport{
	ProductUID:    productUid,
	CompanyUID:    companyUid,
	CompanyName:   companyName,
	ProductName:   "Test Partner AWS",
	LocationId:    1,
	Speed:         10000,
	DiversityZone: "red",
	ConnectType:   "AWS",
	VXCPermitted:  true,
}

// Azure Partner
var azurePartner = &PartnerMegaport{
	ProductUID:   productUid2,
	CompanyUID:   companyUid,
	CompanyName:  companyName,
	ProductName:  "Test Partner Azure",
	LocationId:   2,
	Speed:        10000,
	ConnectType:  "AZURE",
	VXCPermitted: true,
}

// Default Partner
var defaultPartner = &PartnerMegaport{
	ProductUID:    productUid3,
	CompanyUID:    companyUid2,
	CompanyName:   companyName2,
	ProductName:   "Partner Default",
	LocationId:    3,
	Speed:         10000,
	DiversityZone: "red",
	ConnectType:   "DEFAULT",
	VXCPermitted:  true,
}

// AWS Hosted Connection Partner
var awsHcPartner = &PartnerMegaport{
	ProductUID:   productUid4,
	CompanyUID:   companyUid2,
	CompanyName:  companyName2,
	ProductName:  "Partner AWSHC",
	LocationId:   3,
	Speed:        10000,
	ConnectType:  "AWSHC",
	VXCPermitted: true,
}

// JSON blob for testing
var jblob = `{
	"message": "All Partner Megaports",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": [
		{
			"productUid": "36b3f68e-2f54-4331-bf94-f8984449365f",
			"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"companyName": "Test Company",
			"title": "Test Partner AWS",
			"locationId": 1,
			"speed": 10000,
			"maxVxcSpeed": 10000,
			"vxcPermitted": true,
			"diversityZone": "red",
			"connectType": "AWS"
        },
        {
			"productUid": "9b1c46c7-1e8d-4035-bf38-1bc60d346d57",
			"companyUid": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
			"companyName": "Test Company",
			"title": "Test Partner Azure",
			"locationId": 2,
			"speed": 10000,
			"maxVxcSpeed": 10000,
			"vxcPermitted": true,
			"diversityZone": null,
			"connectType": "AZURE"
        },
        {
			"productUid": "91ededc2-473f-4a30-ad24-0703c7f35e50",
			"companyUid": "2111de9a-6211-40e3-8e6a-65ab106d01f4",
			"companyName": "Other Company",
			"title": "Partner Default",
			"locationId": 3,
			"speed": 10000,
			"maxVxcSpeed": 10000,
			"vxcPermitted": true,
			"diversityZone": "red",
			"connectType": "DEFAULT"
        },
        {
			"productUid": "9a05f787-8166-4470-94f1-4906db86f698",
			"companyUid": "2111de9a-6211-40e3-8e6a-65ab106d01f4",
			"companyName": "Other Company",
			"title": "Partner AWSHC",
			"locationId": 3,
			"speed": 10000,
			"maxVxcSpeed": 10000,
			"vxcPermitted": true,
			"diversityZone": null,
			"connectType": "AWSHC"
        }
    ]
}`

// TestListPartnerMegaports tests the ListPartnerMegaports method.
func (suite *PartnerClientTestSuite) TestListPartnerMegaports() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)
}

// TestFilterPartnerMegaportByProductName tests the FilterPartnerMegaportByProductName method.
func (suite *PartnerClientTestSuite) TestFilterPartnerMegaportByProductName() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)

	wantFiltered := []*PartnerMegaport{awsPartner, azurePartner}

	gotFiltered, err := partnerSvc.FilterPartnerMegaportByProductName(ctx, partners, "Test", false)
	suite.NoError(err)
	suite.Equal(wantFiltered, gotFiltered)
}

// TestFilterPartnerMegaportByProductNameExact tests the FilterPartnerMegaportByProductName method with exact match.
func (suite *PartnerClientTestSuite) TestFilterPartnerMegaportByProductNameExact() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)

	wantFiltered := []*PartnerMegaport{awsPartner}

	gotFiltered, err := partnerSvc.FilterPartnerMegaportByProductName(ctx, partners, "Test Partner AWS", true)
	suite.NoError(err)
	suite.Equal(wantFiltered, gotFiltered)
}

// TestFilterPartnerMegaportByConnectType tests the FilterPartnerMegaportByConnectType method.
func (suite *PartnerClientTestSuite) TestFilterPartnerMegaportByConnectType() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)

	wantFiltered := []*PartnerMegaport{awsPartner, awsHcPartner}

	gotFiltered, err := partnerSvc.FilterPartnerMegaportByConnectType(ctx, partners, "AWS", false)
	suite.NoError(err)
	suite.Equal(wantFiltered, gotFiltered)
}

func (suite *PartnerClientTestSuite) TestFilterPartnerMegaportByConnectTypeExact() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)

	wantFiltered := []*PartnerMegaport{awsHcPartner}

	gotFiltered, err := partnerSvc.FilterPartnerMegaportByConnectType(ctx, partners, "AWSHC", true)
	suite.NoError(err)
	suite.Equal(wantFiltered, gotFiltered)
}

// Test FilterPartnerMegaportByCompanyName tests the FilterPartnerMegaportByCompanyName method.
func (suite *PartnerClientTestSuite) TestFilterPartnerMegaportByCompanyName() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)

	wantFiltered := []*PartnerMegaport{awsPartner, azurePartner}

	gotFiltered, err := partnerSvc.FilterPartnerMegaportByCompanyName(ctx, partners, companyName, true)
	suite.NoError(err)
	suite.Equal(wantFiltered, gotFiltered)
}

// TestFilterPartnerMegaportByLocationId tests the FilterPartnerMegaportByLocationId method.
func (suite *PartnerClientTestSuite) TestFilterPartnerMegaportByLocationId() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)

	wantFiltered := []*PartnerMegaport{defaultPartner, awsHcPartner}

	gotFiltered, err := partnerSvc.FilterPartnerMegaportByLocationId(ctx, partners, 3)
	suite.NoError(err)
	suite.Equal(wantFiltered, gotFiltered)
}

// TestFilterPartnerMegaportByMetro tests the FilterPartnerMegaportByMetro method.
func (suite *PartnerClientTestSuite) TestFilterPartnerMegaportByMetro() {
	partnerSvc := suite.client.PartnerService
	locSvc := suite.client.LocationService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	locPath := "/v3/locations"
	locJblob := `{
		"message": "List public locations",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"id": 1,
				"name": "Test Location Sydney",
				"metro": "Sydney",
				"market": "AU",
				"status": "Active",
				"address": {"street":"","suburb":"","city":"Sydney","state":"NSW","postcode":"2000","country":"Australia"},
				"dataCentre": {"id": 1, "name": "Test DC"},
				"latitude": -33.8,
				"longitude": 151.2
			},
			{
				"id": 2,
				"name": "Test Location Sydney 2",
				"metro": "Sydney",
				"market": "AU",
				"status": "Active",
				"address": {"street":"","suburb":"","city":"Sydney","state":"NSW","postcode":"2000","country":"Australia"},
				"dataCentre": {"id": 2, "name": "Test DC 2"},
				"latitude": -33.8,
				"longitude": 151.2
			},
			{
				"id": 3,
				"name": "Test Location Melbourne",
				"metro": "Melbourne",
				"market": "AU",
				"status": "Active",
				"address": {"street":"","suburb":"","city":"Melbourne","state":"VIC","postcode":"3000","country":"Australia"},
				"dataCentre": {"id": 3, "name": "Test DC 3"},
				"latitude": -37.8,
				"longitude": 144.9
			}
		]
	}`
	suite.mux.HandleFunc(locPath, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, locJblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)

	// Filter by "Sydney" metro — locations 1 and 2 are in Sydney.
	// awsPartner (locationId=1) and azurePartner (locationId=2) should match.
	wantFiltered := []*PartnerMegaport{awsPartner, azurePartner}

	gotFiltered, err := partnerSvc.FilterPartnerMegaportByMetro(ctx, partners, locSvc, "Sydney")
	suite.NoError(err)
	suite.Equal(wantFiltered, gotFiltered)

	// Filter by "Melbourne" metro — location 3 is in Melbourne.
	// defaultPartner (locationId=3) and awsHcPartner (locationId=3) should match.
	wantMelbourne := []*PartnerMegaport{defaultPartner, awsHcPartner}

	gotMelbourne, err := partnerSvc.FilterPartnerMegaportByMetro(ctx, partners, locSvc, "Melbourne")
	suite.NoError(err)
	suite.Equal(wantMelbourne, gotMelbourne)

	// Filter by non-existent metro should return ErrNoPartnerPortsFound.
	_, err = partnerSvc.FilterPartnerMegaportByMetro(ctx, partners, locSvc, "Auckland")
	suite.ErrorIs(err, ErrNoPartnerPortsFound)
}

// TestFilterPartnerMegaportByDiversityZone tests the FilterPartnerMegaportByDiversityZone method.
func (suite *PartnerClientTestSuite) TestFilterPartnerMegaportByDiversityZone() {
	partnerSvc := suite.client.PartnerService
	ctx := context.Background()

	want := []*PartnerMegaport{awsPartner, azurePartner, defaultPartner, awsHcPartner}

	suite.mux.HandleFunc(partnerMegaportUrl, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	partners, err := partnerSvc.ListPartnerMegaports(ctx)
	suite.NoError(err)
	suite.Equal(want, partners)

	wantFiltered := []*PartnerMegaport{awsPartner, defaultPartner}

	gotFiltered, err := partnerSvc.FilterPartnerMegaportByDiversityZone(ctx, partners, "red", true)
	suite.NoError(err)
	suite.Equal(wantFiltered, gotFiltered)
}
