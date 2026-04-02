package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

type OrderApprovalClientTestSuite struct {
	ClientTestSuite
}

func TestOrderApprovalClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(OrderApprovalClientTestSuite))
}

func (suite *OrderApprovalClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *OrderApprovalClientTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *OrderApprovalClientTestSuite) TestListOrderApprovals() {
	ctx := context.Background()
	status := OrderApprovalStatusPending
	pageNumber := 1
	pageSize := 10

	listReq := &ListOrderApprovalsRequest{
		Status:     &status,
		PageNumber: &pageNumber,
		PageSize:   &pageSize,
	}

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"uid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				"id": 12345,
				"referenceId": "ORD-54321-45",
				"status": "PENDING",
				"type": "NEW_ORDER",
				"active": true,
				"acctName": "Test Company",
				"acctRef": "ACCT-001",
				"approverCompanyId": 100,
				"requesterCompanyId": 200,
				"serviceId": 5678,
				"comment": "Please approve this order",
				"createDate": 1700000000000,
				"detail": {"type": "NEW_ORDER", "origin": "https://portal.megaport.com", "userName": "user@example.com", "requesterCompanyId": 200, "productRequest": [{"name": "test"}]}
			}
		]
	}`

	want := &ListOrderApprovalsResponse{
		OrderApprovals: []*OrderApproval{
			{
				UID:                "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
				ID:                 12345,
				ReferenceID:        "ORD-54321-45",
				Status:             OrderApprovalStatusPending,
				Type:               OrderApprovalTypeNewOrder,
				Active:             true,
				AcctName:           "Test Company",
				AcctRef:            "ACCT-001",
				ApproverCompanyID:  100,
				RequesterCompanyID: 200,
				ServiceID:          5678,
				Comment:            "Please approve this order",
				CreateDate:         1700000000000,
				Detail:             json.RawMessage(`{"type": "NEW_ORDER", "origin": "https://portal.megaport.com", "userName": "user@example.com", "requesterCompanyId": 200, "productRequest": [{"name": "test"}]}`),
			},
		},
		TotalCount: 1,
		Page:       1,
		Limit:      10,
		TotalPages: 1,
	}

	suite.mux.HandleFunc("/v3/order_approvals", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("PENDING", r.URL.Query().Get("status"))
		suite.Equal("1", r.URL.Query().Get("pageNumber"))
		suite.Equal("10", r.URL.Query().Get("pageSize"))
		w.Header().Set("Pagination-Total-Count", "1")
		w.Header().Set("Pagination-Page", "1")
		w.Header().Set("Pagination-Limit", "10")
		w.Header().Set("Pagination-Total-Page", "1")
		fmt.Fprint(w, jblob)
	})

	listRes, err := suite.client.OrderApprovalService.ListOrderApprovals(ctx, listReq)
	suite.NoError(err)
	suite.Equal(want, listRes)
}

func (suite *OrderApprovalClientTestSuite) TestListOrderApprovalsWithServiceIDs() {
	ctx := context.Background()

	listReq := &ListOrderApprovalsRequest{
		ServiceIDs: []int{123, 456},
	}

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": []
	}`

	suite.mux.HandleFunc("/v3/order_approvals", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("123,456", r.URL.Query().Get("serviceIds"))
		fmt.Fprint(w, jblob)
	})

	listRes, err := suite.client.OrderApprovalService.ListOrderApprovals(ctx, listReq)
	suite.NoError(err)
	suite.NotNil(listRes)
	suite.Empty(listRes.OrderApprovals)
}

func (suite *OrderApprovalClientTestSuite) TestApproveOrderApproval() {
	ctx := context.Background()
	uid := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": "Order request approved successfully"
	}`

	suite.mux.HandleFunc(fmt.Sprintf("/v3/order_approvals/%s/approve", uid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		body, _ := io.ReadAll(r.Body)
		var req OrderApprovalActionRequest
		_ = json.Unmarshal(body, &req)
		suite.Equal("Looks good", req.Comments)
		fmt.Fprint(w, jblob)
	})

	err := suite.client.OrderApprovalService.ApproveOrderApproval(ctx, uid, &OrderApprovalActionRequest{
		Comments: "Looks good",
	})
	suite.NoError(err)
}

func (suite *OrderApprovalClientTestSuite) TestRejectOrderApproval() {
	ctx := context.Background()
	uid := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": "Order request rejected successfully"
	}`

	suite.mux.HandleFunc(fmt.Sprintf("/v3/order_approvals/%s/reject", uid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		body, _ := io.ReadAll(r.Body)
		var req OrderApprovalActionRequest
		_ = json.Unmarshal(body, &req)
		suite.Equal("Not approved", req.Comments)
		fmt.Fprint(w, jblob)
	})

	err := suite.client.OrderApprovalService.RejectOrderApproval(ctx, uid, &OrderApprovalActionRequest{
		Comments: "Not approved",
	})
	suite.NoError(err)
}

func (suite *OrderApprovalClientTestSuite) TestWithdrawOrderApproval() {
	ctx := context.Background()
	uid := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": "Order request withdrawn successfully"
	}`

	suite.mux.HandleFunc(fmt.Sprintf("/v3/order_approvals/%s/withdraw", uid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		body, _ := io.ReadAll(r.Body)
		var req OrderApprovalActionRequest
		_ = json.Unmarshal(body, &req)
		suite.Equal("No longer needed", req.Comments)
		fmt.Fprint(w, jblob)
	})

	err := suite.client.OrderApprovalService.WithdrawOrderApproval(ctx, uid, &OrderApprovalActionRequest{
		Comments: "No longer needed",
	})
	suite.NoError(err)
}
