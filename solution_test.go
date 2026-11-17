package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	testSolutionUID  = "8da180e1-61ea-4961-a6e3-d80152ee9b14"
	testSolutionPath = "/v1/solutions/" + testSolutionUID
)

// SolutionClientTestSuite tests the Solution service.
type SolutionClientTestSuite struct {
	ClientTestSuite
	ctx context.Context
}

func TestSolutionClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(SolutionClientTestSuite))
}

func (suite *SolutionClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, err := url.Parse(suite.server.URL)
	suite.NoError(err)
	suite.client.BaseURL = url
	suite.ctx = context.Background()
}

func (suite *SolutionClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// dciBuyBody is the body of a DCI order, decoded loosely so a test can assert on the wire shape.
type dciBuyBody struct {
	NetworkDesign []map[string]any `json:"networkDesign"`
	DiscountCodes []string         `json:"discountCodes"`
	Solutions     []map[string]any `json:"solutions"`
}

// handleDCIBuy answers a DCI order with the Solution it declares, and returns the body the client sent.
func (suite *SolutionClientTestSuite) handleDCIBuy() *dciBuyBody {
	got := &dciBuyBody{}
	suite.mux.HandleFunc("/v4/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		body, err := io.ReadAll(r.Body)
		suite.NoError(err)
		suite.NoError(json.Unmarshal(body, got))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"message": "Order placed",
			"data": [{"technicalServiceUid": "a-end"}, {"technicalServiceUid": "b-end"}],
			"solutions": [
				{"uid": "another-solution", "ref": "another-ref", "type": "DCI", "name": "Another"},
				{"uid": %q, "ref": %q, "type": "DCI", "name": "My DCI"}
			]
		}`, testSolutionUID, got.Solutions[0]["ref"])
	})
	return got
}

func (suite *SolutionClientTestSuite) TestBuyDCI() {
	got := suite.handleDCIBuy()

	resp, err := suite.client.SolutionService.BuyDCI(suite.ctx, &BuyDCIRequest{
		Name:                  "My DCI",
		AEndLocationID:        143,
		BEndLocationID:        145,
		PortSpeed:             10000,
		Term:                  12,
		AutoRenewTerm:         true,
		ServiceLevelReference: "SL-001",
		AEndDiversityZone:     "red",
	})
	suite.NoError(err)
	suite.Equal(&BuyDCIResponse{SolutionUID: testSolutionUID}, resp)

	suite.Equal([]string{}, got.DiscountCodes)

	suite.Require().Len(got.Solutions, 1)
	ref := got.Solutions[0]["ref"]
	suite.NotEmpty(ref)
	suite.Equal(map[string]any{"ref": ref, "type": "DCI", "name": "My DCI", "serviceLevelReference": "SL-001"}, got.Solutions[0])

	suite.Require().Len(got.NetworkDesign, 2)
	aEnd, bEnd := got.NetworkDesign[0], got.NetworkDesign[1]
	aEndUID, bEndUID := aEnd["productUid"], bEnd["productUid"]
	suite.NotEmpty(aEndUID)
	suite.NotEmpty(bEndUID)
	suite.NotEqual(aEndUID, bEndUID)

	vxcs, ok := aEnd["associatedVxcs"].([]any)
	suite.Require().True(ok)
	suite.Require().Len(vxcs, 1)
	vxc, ok := vxcs[0].(map[string]any)
	suite.Require().True(ok)
	vxcUID := vxc["productUid"]
	suite.NotEmpty(vxcUID)

	suite.Equal(map[string]any{
		"productUid":         aEndUID,
		"productName":        "My DCI",
		"productType":        "MEGAPORT",
		"provisioningStatus": "DESIGN",
		"locationId":         float64(143),
		"portSpeed":          float64(10000),
		"term":               float64(12),
		"autoRenewTerm":      true,
		"solutionRef":        ref,
		"config":             map[string]any{"diversityZone": "red"},
		"associatedVxcs":     vxcs,
	}, aEnd)
	suite.Equal(map[string]any{
		"productUid":         vxcUID,
		"productName":        "My DCI",
		"productType":        "VXC",
		"provisioningStatus": "DESIGN",
		"rateLimit":          float64(10000),
		"term":               float64(12),
		"autoRenewTerm":      true,
		"connectType":        "DEFAULT",
		"aEnd":               map[string]any{"productUid": aEndUID},
		"bEnd":               map[string]any{"productUid": bEndUID},
		"solutionRef":        ref,
	}, vxc)
	// The B-End Port sends an empty list, not null, for the VXCs it does not carry.
	suite.Equal(map[string]any{
		"productUid":         bEndUID,
		"productName":        "My DCI",
		"productType":        "MEGAPORT",
		"provisioningStatus": "DESIGN",
		"locationId":         float64(145),
		"portSpeed":          float64(10000),
		"term":               float64(12),
		"autoRenewTerm":      true,
		"solutionRef":        ref,
		"config":             map[string]any{},
		"associatedVxcs":     []any{},
	}, bEnd)
}

// An order without a Service Level Reference leaves it out of the Solution.
func (suite *SolutionClientTestSuite) TestBuyDCINoServiceLevelReference() {
	got := suite.handleDCIBuy()

	_, err := suite.client.SolutionService.BuyDCI(suite.ctx, &BuyDCIRequest{
		Name: "My DCI", AEndLocationID: 143, BEndLocationID: 145, PortSpeed: 1000, Term: 36,
	})
	suite.NoError(err)
	suite.Require().Len(got.Solutions, 1)
	suite.NotContains(got.Solutions[0], "serviceLevelReference")
	suite.Equal(false, got.NetworkDesign[0]["autoRenewTerm"])
}

func (suite *SolutionClientTestSuite) TestBuyDCIInvalidRequest() {
	suite.mux.HandleFunc("/v4/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("an invalid request must not reach the API")
	})

	_, err := suite.client.SolutionService.BuyDCI(suite.ctx, nil)
	suite.ErrorIs(err, ErrBuyDCIRequestNil)

	_, err = suite.client.SolutionService.BuyDCI(suite.ctx, &BuyDCIRequest{Name: "  "})
	suite.ErrorIs(err, ErrSolutionNameRequired)
}

// The API's own error reaches the caller, as for a term a DCI is not sold on.
func (suite *SolutionClientTestSuite) TestBuyDCIAPIError() {
	const message = "Solution 'ref' cannot be ordered on a 1 month term; a DCI is available on 12, 24, 36 months only"
	suite.mux.HandleFunc("/v4/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"message": %q}`, message)
	})

	_, err := suite.client.SolutionService.BuyDCI(suite.ctx, &BuyDCIRequest{
		Name: "My DCI", AEndLocationID: 143, BEndLocationID: 145, PortSpeed: 1000, Term: 1,
	})
	var apiErr *ErrorResponse
	suite.Require().True(errors.As(err, &apiErr))
	suite.Equal(http.StatusBadRequest, apiErr.Response.StatusCode)
	suite.Equal(message, apiErr.Message)
}

// An order the API accepts without naming the DCI is an error, because the caller could never read or cancel it.
func (suite *SolutionClientTestSuite) TestBuyDCIResponseMissingSolution() {
	suite.mux.HandleFunc("/v4/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message": "Order placed", "data": [{"technicalServiceUid": "a-end"}]}`)
	})

	_, err := suite.client.SolutionService.BuyDCI(suite.ctx, &BuyDCIRequest{
		Name: "My DCI", AEndLocationID: 143, BEndLocationID: 145, PortSpeed: 1000, Term: 12,
	})
	suite.ErrorIs(err, ErrDCIOrderResponseMissingSolution)
}

func (suite *SolutionClientTestSuite) TestGetSolution() {
	suite.mux.HandleFunc(testSolutionPath, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"message": "Solution retrieved successfully",
			"terms": "terms",
			"data": {
				"uid": %q,
				"type": "DCI",
				"name": "My DCI",
				"active": true,
				"locked": false,
				"serviceLevelReference": "SL-001",
				"projectUid": null,
				"provisioningStatus": "CONFIGURED",
				"aLocationId": 143,
				"bLocationId": 145,
				"speedMbps": 10000,
				"createDate": "2026-10-21T14:00:00Z",
				"createdBy": "6a7b8c9d-1e2f-4a3b-8c5d-6e7f8a9b0c1d",
				"liveDate": null,
				"companyUid": "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
				"companyName": "Example Company Pty Ltd",
				"members": [
					{"position": 0, "role": "A_END_PORT", "diversityZone": "red", "locationId": 143, "provisioningStatus": "LIVE", "serviceName": "My DCI", "serviceType": "MEGAPORT", "serviceUid": "a-end-uid"},
					{"position": 1, "role": "B_END_PORT", "diversityZone": null, "locationId": 145, "provisioningStatus": "CONFIGURED", "serviceName": "My DCI", "serviceType": "MEGAPORT", "serviceUid": "b-end-uid"},
					{"position": 2, "role": "VXC", "diversityZone": null, "locationId": null, "provisioningStatus": "LIVE", "serviceName": "My DCI", "serviceType": "VXC", "serviceUid": "vxc-uid"}
				]
			}
		}`, testSolutionUID)
	})

	got, err := suite.client.SolutionService.GetSolution(suite.ctx, testSolutionUID)
	suite.NoError(err)
	suite.Equal(&Solution{
		UID:                   testSolutionUID,
		Type:                  SolutionTypeDCI,
		Name:                  "My DCI",
		Active:                true,
		ServiceLevelReference: "SL-001",
		ProvisioningStatus:    SERVICE_CONFIGURED,
		ALocationID:           143,
		BLocationID:           145,
		SpeedMbps:             10000,
		CreateDate:            &Time{Time: time.Date(2026, 10, 21, 14, 0, 0, 0, time.UTC)},
		CreatedBy:             "6a7b8c9d-1e2f-4a3b-8c5d-6e7f8a9b0c1d",
		CompanyUID:            "1a2b3c4d-5e6f-4a7b-8c9d-0e1f2a3b4c5d",
		CompanyName:           "Example Company Pty Ltd",
		Members: []SolutionMember{
			{ServiceUID: "a-end-uid", ServiceName: "My DCI", ServiceType: "MEGAPORT", ProvisioningStatus: SERVICE_LIVE, Role: SolutionMemberRoleAEndPort, Position: 0, LocationID: 143, DiversityZone: "red"},
			{ServiceUID: "b-end-uid", ServiceName: "My DCI", ServiceType: "MEGAPORT", ProvisioningStatus: SERVICE_CONFIGURED, Role: SolutionMemberRoleBEndPort, Position: 1, LocationID: 145},
			{ServiceUID: "vxc-uid", ServiceName: "My DCI", ServiceType: "VXC", ProvisioningStatus: SERVICE_LIVE, Role: SolutionMemberRoleVXC, Position: 2},
		},
	}, got)

	suite.Equal("a-end-uid", got.MemberUID(SolutionMemberRoleAEndPort))
	suite.Equal("b-end-uid", got.MemberUID(SolutionMemberRoleBEndPort))
	suite.Equal("vxc-uid", got.MemberUID(SolutionMemberRoleVXC))
	suite.Empty(got.MemberUID("UNKNOWN_ROLE"))
}

// A Solution the API cannot find comes back as its 404, which a caller reads to drop the Solution.
func (suite *SolutionClientTestSuite) TestGetSolutionNotFound() {
	suite.mux.HandleFunc(testSolutionPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"message": "The requested endpoint is not found."}`)
	})

	_, err := suite.client.SolutionService.GetSolution(suite.ctx, testSolutionUID)
	var apiErr *ErrorResponse
	suite.Require().True(errors.As(err, &apiErr))
	suite.Equal(http.StatusNotFound, apiErr.Response.StatusCode)
}

func (suite *SolutionClientTestSuite) TestGetSolutionEmptyResponse() {
	suite.mux.HandleFunc(testSolutionPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message": "ok", "data": null}`)
	})

	_, err := suite.client.SolutionService.GetSolution(suite.ctx, testSolutionUID)
	suite.ErrorIs(err, ErrSolutionResponseEmpty)
}

func (suite *SolutionClientTestSuite) TestGetSolutionBillable() {
	suite.mux.HandleFunc(testSolutionPath+"/billable", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"message": "Found billing for solution",
			"data": {
				"autoRenewTerm": true,
				"canReterm": false,
				"contractStartDate": 1767225600000,
				"contractEndDate": null,
				"contractTermMonths": 12,
				"price": null
			}
		}`)
	})

	got, err := suite.client.SolutionService.GetSolutionBillable(suite.ctx, testSolutionUID)
	suite.NoError(err)
	suite.Equal(&SolutionBillable{
		ContractTermMonths: 12,
		AutoRenewTerm:      true,
		ContractStartDate:  &Time{Time: time.Unix(1767225600, 0)},
	}, got)
}

func (suite *SolutionClientTestSuite) TestSolutionUIDRequired() {
	_, err := suite.client.SolutionService.GetSolution(ctx, "")
	suite.ErrorIs(err, ErrSolutionUIDRequired)

	_, err = suite.client.SolutionService.GetSolutionBillable(ctx, "")
	suite.ErrorIs(err, ErrSolutionUIDRequired)
}

func TestNewUUID(t *testing.T) {
	t.Parallel()
	version4 := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	seen := map[string]bool{}
	for i := 0; i < 100; i++ {
		uid, err := newUUID()
		require.NoError(t, err)
		require.Regexp(t, version4, uid)
		require.False(t, seen[uid], "%q was generated twice", uid)
		seen[uid] = true
	}
}
