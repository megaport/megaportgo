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
	"time"

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
	approvalType := OrderApprovalTypeNewOrder
	// The API caps the page size at 100 and answers a page past the last with page 1.
	pageNumber := 4
	pageSize := 500
	sort := "createdAt"
	direction := "ASC"

	listReq := &ListOrderApprovalsRequest{
		Status:     &status,
		Type:       &approvalType,
		ServiceIDs: []int{123, 456},
		PageNumber: &pageNumber,
		PageSize:   &pageSize,
		Sort:       &sort,
		Direction:  &direction,
	}

	jblob := `{
		"message": "Success",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {
			"pageNumber": 1,
			"pageSize": 100,
			"totalItems": 25,
			"totalPages": 1,
			"items": [
				{
					"uid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
					"id": 12345,
					"referenceId": "ORD-54321-45",
					"status": "PENDING",
					"type": "NEW_ORDER",
					"active": true,
					"accountName": "Test Company",
					"accountReference": "ACCT-001",
					"approverCompanyId": 100,
					"requesterCompanyId": 200,
					"serviceId": 5678,
					"comment": "Please approve this order",
					"createdAt": "2026-10-22T23:22:39.977Z",
					"completedAt": null,
					"createdBy": "0a1717f0-7f85-470e-954d-7ad208392a2c",
					"detail": {"origin": "https://portal.megaport.com", "userName": "user@example.com", "requesterCompanyId": 200}
				},
				{
					"uid": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
					"id": 12346,
					"status": "APPROVED",
					"type": "EARLY_TERMINATION",
					"createdAt": "2026-10-21T01:00:00Z",
					"completedAt": "2026-10-22T02:30:15.250Z"
				},
				{
					"uid": "c3d4e5f6-a7b8-9012-cdef-123456789012",
					"id": 12347,
					"status": "PENDING",
					"type": "ADD_ON"
				},
				{
					"uid": "d4e5f6a7-b8c9-0123-def0-234567890123",
					"id": 12348,
					"status": "PENDING",
					"type": "IP_ADDRESS_PARTNER_ORDER"
				},
				{
					"uid": "e5f6a7b8-c9d0-1234-ef01-345678901234",
					"id": 12349,
					"status": "PENDING",
					"type": "IP_ADDRESS_COMPLIANCE"
				}
			]
		}
	}`

	suite.mux.HandleFunc("/v4/order_approvals", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		query := r.URL.Query()
		suite.Equal("PENDING", query.Get("status"))
		suite.Equal("NEW_ORDER", query.Get("type"))
		suite.Equal([]string{"123", "456"}, query["service_ids"])
		suite.Equal("4", query.Get("page_number"))
		suite.Equal("500", query.Get("page_size"))
		suite.Equal([]string{"createdAt,ASC"}, query["sort"])
		suite.Empty(query["direction"])
		fmt.Fprint(w, jblob)
	})

	listRes, err := suite.client.OrderApprovalService.ListOrderApprovals(ctx, listReq)
	suite.Require().NoError(err)

	suite.Equal(25, listRes.TotalCount)
	suite.Equal(1, listRes.Page)
	suite.Equal(100, listRes.Limit)
	suite.Equal(1, listRes.TotalPages)
	suite.Require().Len(listRes.OrderApprovals, 5)

	first := listRes.OrderApprovals[0]
	suite.Equal("a1b2c3d4-e5f6-7890-abcd-ef1234567890", first.UID)
	suite.Equal(12345, first.ID)
	suite.Equal("ORD-54321-45", first.ReferenceID)
	suite.Equal(OrderApprovalStatusPending, first.Status)
	suite.Equal(OrderApprovalTypeNewOrder, first.Type)
	suite.True(first.Active)
	suite.Equal("Test Company", first.AcctName)
	suite.Equal("ACCT-001", first.AcctRef)
	suite.Equal(100, first.ApproverCompanyID)
	suite.Equal(200, first.RequesterCompanyID)
	suite.Equal(5678, first.ServiceID)
	suite.Equal("Please approve this order", first.Comment)
	suite.Equal("0a1717f0-7f85-470e-954d-7ad208392a2c", first.CreatedBy)
	suite.Require().NotNil(first.CreateDate)
	suite.True(first.CreateDate.Equal(time.Date(2026, 10, 22, 23, 22, 39, 977000000, time.UTC)))
	suite.Nil(first.CompletedAt)
	suite.JSONEq(`{"origin": "https://portal.megaport.com", "userName": "user@example.com", "requesterCompanyId": 200}`, string(first.Detail))

	second := listRes.OrderApprovals[1]
	suite.Equal(OrderApprovalStatusApproved, second.Status)
	suite.Equal(OrderApprovalTypeEarlyTermination, second.Type)
	suite.Require().NotNil(second.CompletedAt)
	suite.True(second.CompletedAt.Equal(time.Date(2026, 10, 22, 2, 30, 15, 250000000, time.UTC)))

	suite.Equal(OrderApprovalTypeAddOn, listRes.OrderApprovals[2].Type)
	suite.Equal(OrderApprovalTypeIPAddressPartnerOrder, listRes.OrderApprovals[3].Type)
	suite.Equal(OrderApprovalTypeIPAddressCompliance, listRes.OrderApprovals[4].Type)
}

func (suite *OrderApprovalClientTestSuite) TestListOrderApprovalsSort() {
	sort := "serviceId"
	v3Sort := "createDate"
	direction := "DESC"

	tests := []struct {
		name      string
		sort      *string
		direction *string
		want      []string
	}{
		{name: "sort and direction", sort: &sort, direction: &direction, want: []string{"serviceId,DESC"}},
		{name: "direction only", direction: &direction, want: []string{"createdAt,DESC"}},
		{name: "sort only", sort: &sort, want: []string{"serviceId"}},
		{name: "v3 createDate and direction", sort: &v3Sort, direction: &direction, want: []string{"createdAt,DESC"}},
		{name: "v3 createDate only", sort: &v3Sort, want: []string{"createdAt"}},
		{name: "neither", want: nil},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			defer server.Close()
			client := NewClient(nil, nil)
			client.BaseURL, _ = url.Parse(server.URL)

			mux.HandleFunc("/v4/order_approvals", func(w http.ResponseWriter, r *http.Request) {
				suite.Equal(tt.want, r.URL.Query()["sort"])
				suite.Empty(r.URL.Query()["direction"])
				fmt.Fprint(w, `{"message": "Success", "data": {"items": []}}`)
			})

			_, err := client.OrderApprovalService.ListOrderApprovals(context.Background(), &ListOrderApprovalsRequest{
				Sort:      tt.sort,
				Direction: tt.direction,
			})
			suite.NoError(err)
		})
	}
}

func (suite *OrderApprovalClientTestSuite) TestListOrderApprovalsNilRequest() {
	ctx := context.Background()

	jblob := `{
		"message": "Success",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": {"pageNumber": 1, "pageSize": 10, "totalItems": 0, "totalPages": 0, "items": []}
	}`

	suite.mux.HandleFunc("/v4/order_approvals", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Empty(r.URL.Query())
		fmt.Fprint(w, jblob)
	})

	listRes, err := suite.client.OrderApprovalService.ListOrderApprovals(ctx, nil)
	suite.NoError(err)
	suite.NotNil(listRes)
	suite.Empty(listRes.OrderApprovals)
	suite.Equal(1, listRes.Page)
	suite.Equal(10, listRes.Limit)
	suite.Equal(0, listRes.TotalPages)
}

func (suite *OrderApprovalClientTestSuite) TestListOrderApprovalsDecodeError() {
	suite.mux.HandleFunc("/v4/order_approvals", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"message": "Success", "data": []}`)
	})

	listRes, err := suite.client.OrderApprovalService.ListOrderApprovals(context.Background(), nil)
	suite.Error(err)
	suite.Nil(listRes)
}

func (suite *OrderApprovalClientTestSuite) TestActionEmptyUID() {
	ctx := context.Background()
	err := suite.client.OrderApprovalService.ApproveOrderApproval(ctx, "", &OrderApprovalActionRequest{})
	suite.ErrorIs(err, ErrOrderApprovalUIDRequired)

	err = suite.client.OrderApprovalService.RejectOrderApproval(ctx, "", &OrderApprovalActionRequest{})
	suite.ErrorIs(err, ErrOrderApprovalUIDRequired)

	err = suite.client.OrderApprovalService.WithdrawOrderApproval(ctx, "", &OrderApprovalActionRequest{})
	suite.ErrorIs(err, ErrOrderApprovalUIDRequired)
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
		body, err := io.ReadAll(r.Body)
		suite.Require().NoError(err)
		var req OrderApprovalActionRequest
		suite.Require().NoError(json.Unmarshal(body, &req))
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
		body, err := io.ReadAll(r.Body)
		suite.Require().NoError(err)
		var req OrderApprovalActionRequest
		suite.Require().NoError(json.Unmarshal(body, &req))
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
		body, err := io.ReadAll(r.Body)
		suite.Require().NoError(err)
		var req OrderApprovalActionRequest
		suite.Require().NoError(json.Unmarshal(body, &req))
		suite.Equal("No longer needed", req.Comments)
		fmt.Fprint(w, jblob)
	})

	err := suite.client.OrderApprovalService.WithdrawOrderApproval(ctx, uid, &OrderApprovalActionRequest{
		Comments: "No longer needed",
	})
	suite.NoError(err)
}
