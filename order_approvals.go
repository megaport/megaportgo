package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// OrderApprovalStatus represents the status of an order approval.
type OrderApprovalStatus string

const (
	OrderApprovalStatusPending   OrderApprovalStatus = "PENDING"
	OrderApprovalStatusApproved  OrderApprovalStatus = "APPROVED"
	OrderApprovalStatusRejected  OrderApprovalStatus = "REJECTED"
	OrderApprovalStatusFailed    OrderApprovalStatus = "FAILED"
	OrderApprovalStatusWithdrawn OrderApprovalStatus = "WITHDRAWN"
	OrderApprovalStatusExpired   OrderApprovalStatus = "EXPIRED"
)

// OrderApprovalType represents the type of an order approval.
type OrderApprovalType string

const (
	OrderApprovalTypeNewOrder    OrderApprovalType = "NEW_ORDER"
	OrderApprovalTypeTermChange  OrderApprovalType = "TERM_CHANGE"
	OrderApprovalTypeSpeedChange OrderApprovalType = "SPEED_CHANGE"
)

// OrderApprovalService is an interface for interfacing with the Order Approval endpoints in the Megaport API.
type OrderApprovalService interface {
	// ListOrderApprovals lists order approval requests from the Megaport API.
	ListOrderApprovals(ctx context.Context, req *ListOrderApprovalsRequest) (*ListOrderApprovalsResponse, error)
	// ApproveOrderApproval approves a pending order approval request.
	ApproveOrderApproval(ctx context.Context, orderApprovalUID string, req *OrderApprovalActionRequest) error
	// RejectOrderApproval rejects a pending order approval request.
	RejectOrderApproval(ctx context.Context, orderApprovalUID string, req *OrderApprovalActionRequest) error
	// WithdrawOrderApproval withdraws own pending order approval request.
	WithdrawOrderApproval(ctx context.Context, orderApprovalUID string, req *OrderApprovalActionRequest) error
}

// NewOrderApprovalService creates a new instance of the Order Approval Service.
func NewOrderApprovalService(c *Client) *OrderApprovalServiceOp {
	return &OrderApprovalServiceOp{
		Client: c,
	}
}

var _ OrderApprovalService = &OrderApprovalServiceOp{}

// OrderApprovalServiceOp handles communication with the Order Approval related methods of the Megaport API.
type OrderApprovalServiceOp struct {
	Client *Client
}

// OrderApproval represents an order approval from the Megaport API.
type OrderApproval struct {
	UID                string              `json:"uid"`
	ID                 int                 `json:"id"`
	ReferenceID        string              `json:"referenceId"`
	Status             OrderApprovalStatus `json:"status"`
	Type               OrderApprovalType   `json:"type"`
	Active             bool                `json:"active"`
	AcctName           string              `json:"acctName"`
	AcctRef            string              `json:"acctRef"`
	ApproverCompanyID  int                 `json:"approverCompanyId"`
	RequesterCompanyID int                 `json:"requesterCompanyId"`
	ServiceID          int                 `json:"serviceId"`
	Comment            string              `json:"comment"`
	CreateDate         int64               `json:"createDate"`
	Detail             json.RawMessage     `json:"detail"`
}

// ListOrderApprovalsRequest represents a request to list order approvals from the Megaport API.
type ListOrderApprovalsRequest struct {
	Status     *OrderApprovalStatus // Filter by approval status (optional).
	ServiceIDs []int                // Filter by service IDs (optional).
	PageNumber *int                 // Page number for pagination (default 1).
	PageSize   *int                 // Page size for pagination (1-100, default 10).
	Sort       *string              // Field to sort by (optional).
	Direction  *string              // Sort direction: ASC or DESC (default DESC).
}

// ListOrderApprovalsAPIResponse represents the Megaport API HTTP response from listing order approvals.
type ListOrderApprovalsAPIResponse struct {
	Message string           `json:"message"`
	Terms   string           `json:"terms"`
	Data    []*OrderApproval `json:"data"`
}

// ListOrderApprovalsResponse represents the Go SDK response from listing order approvals.
type ListOrderApprovalsResponse struct {
	OrderApprovals []*OrderApproval
	TotalCount     int
	Page           int
	Limit          int
	TotalPages     int
}

// OrderApprovalActionRequest represents a request to approve, reject, or withdraw an order approval.
type OrderApprovalActionRequest struct {
	Comments string `json:"comments,omitempty"`
}

// OrderApprovalActionAPIResponse represents the Megaport API HTTP response from an order approval action.
type OrderApprovalActionAPIResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    string `json:"data"`
}

// ListOrderApprovals lists order approval requests from the Megaport API.
func (svc *OrderApprovalServiceOp) ListOrderApprovals(ctx context.Context, req *ListOrderApprovalsRequest) (*ListOrderApprovalsResponse, error) {
	path := "/v3/order_approvals"
	params := url.Values{}
	if req.Status != nil {
		params.Add("status", string(*req.Status))
	}
	if len(req.ServiceIDs) > 0 {
		ids := make([]string, len(req.ServiceIDs))
		for i, id := range req.ServiceIDs {
			ids[i] = strconv.Itoa(id)
		}
		params.Add("serviceIds", strings.Join(ids, ","))
	}
	if req.PageNumber != nil {
		params.Add("pageNumber", strconv.Itoa(*req.PageNumber))
	}
	if req.PageSize != nil {
		params.Add("pageSize", strconv.Itoa(*req.PageSize))
	}
	if req.Sort != nil {
		params.Add("sort", *req.Sort)
	}
	if req.Direction != nil {
		params.Add("direction", *req.Direction)
	}

	u := svc.Client.BaseURL.JoinPath(path)
	if len(params) > 0 {
		u.RawQuery = params.Encode()
	}
	urlString := u.String()

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, urlString, nil)
	if err != nil {
		return nil, err
	}
	response, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return nil, resErr
	}
	defer response.Body.Close()

	svc.Client.Logger.DebugContext(ctx, "Listing Order Approvals", slog.String("url", urlString), slog.Int("status_code", response.StatusCode))

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	var apiResponse ListOrderApprovalsAPIResponse
	if err = json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	toReturn := &ListOrderApprovalsResponse{
		OrderApprovals: apiResponse.Data,
	}

	if v := response.Header.Get("Pagination-Total-Count"); v != "" {
		toReturn.TotalCount, _ = strconv.Atoi(v)
	}
	if v := response.Header.Get("Pagination-Page"); v != "" {
		toReturn.Page, _ = strconv.Atoi(v)
	}
	if v := response.Header.Get("Pagination-Limit"); v != "" {
		toReturn.Limit, _ = strconv.Atoi(v)
	}
	if v := response.Header.Get("Pagination-Total-Page"); v != "" {
		toReturn.TotalPages, _ = strconv.Atoi(v)
	}

	return toReturn, nil
}

// ApproveOrderApproval approves a pending order approval request.
func (svc *OrderApprovalServiceOp) ApproveOrderApproval(ctx context.Context, orderApprovalUID string, req *OrderApprovalActionRequest) error {
	return svc.doAction(ctx, orderApprovalUID, "approve", req)
}

// RejectOrderApproval rejects a pending order approval request.
func (svc *OrderApprovalServiceOp) RejectOrderApproval(ctx context.Context, orderApprovalUID string, req *OrderApprovalActionRequest) error {
	return svc.doAction(ctx, orderApprovalUID, "reject", req)
}

// WithdrawOrderApproval withdraws own pending order approval request.
func (svc *OrderApprovalServiceOp) WithdrawOrderApproval(ctx context.Context, orderApprovalUID string, req *OrderApprovalActionRequest) error {
	return svc.doAction(ctx, orderApprovalUID, "withdraw", req)
}

func (svc *OrderApprovalServiceOp) doAction(ctx context.Context, orderApprovalUID string, action string, req *OrderApprovalActionRequest) error {
	path := fmt.Sprintf("/v3/order_approvals/%s/%s", orderApprovalUID, action)
	u := svc.Client.BaseURL.JoinPath(path).String()

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return err
	}
	response, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return resErr
	}
	if response != nil {
		svc.Client.Logger.DebugContext(ctx, fmt.Sprintf("Order Approval %s", action), slog.String("url", u), slog.Int("status_code", response.StatusCode))
		defer response.Body.Close()
	}
	return nil
}
