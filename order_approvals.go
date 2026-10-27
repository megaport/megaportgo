package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
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
	OrderApprovalTypeNewOrder              OrderApprovalType = "NEW_ORDER"
	OrderApprovalTypeTermChange            OrderApprovalType = "TERM_CHANGE"
	OrderApprovalTypeSpeedChange           OrderApprovalType = "SPEED_CHANGE"
	OrderApprovalTypeEarlyTermination      OrderApprovalType = "EARLY_TERMINATION"
	OrderApprovalTypeAddOn                 OrderApprovalType = "ADD_ON"
	OrderApprovalTypeIPAddressPartnerOrder OrderApprovalType = "IP_ADDRESS_PARTNER_ORDER"
	OrderApprovalTypeIPAddressCompliance   OrderApprovalType = "IP_ADDRESS_COMPLIANCE"
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

// ErrOrderApprovalUIDRequired is returned when an order approval UID is not provided.
var ErrOrderApprovalUIDRequired = errors.New("order approval UID is required")

// OrderApproval represents an order approval from the Megaport API.
type OrderApproval struct {
	UID                string              `json:"uid"`
	ID                 int                 `json:"id"`
	ReferenceID        string              `json:"referenceId"`
	Status             OrderApprovalStatus `json:"status"`
	Type               OrderApprovalType   `json:"type"`
	Active             bool                `json:"active"`
	AcctName           string              `json:"accountName"`
	AcctRef            string              `json:"accountReference"`
	ApproverCompanyID  int                 `json:"approverCompanyId"`
	RequesterCompanyID int                 `json:"requesterCompanyId"`
	ServiceID          int                 `json:"serviceId"`
	Comment            string              `json:"comment"`
	CreateDate         *Time               `json:"createdAt"`
	CompletedAt        *Time               `json:"completedAt"`
	CreatedBy          string              `json:"createdBy"`
	Detail             json.RawMessage     `json:"detail"`
}

// ListOrderApprovalsRequest represents a request to list order approvals from the Megaport API.
type ListOrderApprovalsRequest struct {
	Status     *OrderApprovalStatus // Filter by approval status (optional).
	Type       *OrderApprovalType   // Filter by approval type (optional).
	ServiceIDs []int                // Filter by service IDs (optional).
	PageNumber *int                 // Page number for pagination; if nil, the API default is used. The API answers a page past the last with page 1.
	PageSize   *int                 // Page size for pagination; if nil, the API default is used.
	Sort       *string              // Field to sort by: createdAt, serviceId, status or type. The API rejects v3's createDate with a 400.
	Direction  *string              // Sort direction: ASC or DESC. A Sort with no Direction sorts ascending. With Sort nil, Direction applies to createdAt.
}

// ListOrderApprovalsAPIResponse matches neither the v3 nor the v4 list response.
//
// Deprecated: ListOrderApprovals no longer decodes it.
type ListOrderApprovalsAPIResponse struct {
	Message string           `json:"message"`
	Terms   string           `json:"terms"`
	Data    []*OrderApproval `json:"data"`
}

// listOrderApprovalsV4Response is the v4 response, whose data holds one page.
type listOrderApprovalsV4Response struct {
	Message string                   `json:"message"`
	Terms   string                   `json:"terms"`
	Data    listOrderApprovalsV4Page `json:"data"`
}

type listOrderApprovalsV4Page struct {
	Items      []*OrderApproval `json:"items"`
	PageNumber int              `json:"pageNumber"`
	PageSize   int              `json:"pageSize"`
	TotalItems int              `json:"totalItems"`
	TotalPages int              `json:"totalPages"`
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

// ListOrderApprovals lists order approval requests from the Megaport API.
func (svc *OrderApprovalServiceOp) ListOrderApprovals(ctx context.Context, req *ListOrderApprovalsRequest) (*ListOrderApprovalsResponse, error) {
	if req == nil {
		req = &ListOrderApprovalsRequest{}
	}
	path := "/v4/order_approvals"
	params := url.Values{}
	if req.Status != nil {
		params.Add("status", string(*req.Status))
	}
	if req.Type != nil {
		params.Add("type", string(*req.Type))
	}
	for _, id := range req.ServiceIDs {
		params.Add("service_ids", strconv.Itoa(id))
	}
	if req.PageNumber != nil {
		params.Add("page_number", strconv.Itoa(*req.PageNumber))
	}
	if req.PageSize != nil {
		params.Add("page_size", strconv.Itoa(*req.PageSize))
	}
	if sort := orderApprovalSort(req); sort != "" {
		params.Add("sort", sort)
	}

	if len(params) > 0 {
		path = path + "?" + params.Encode()
	}

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	response, resErr := svc.Client.Do(ctx, clientReq, &buf)
	if resErr != nil {
		return nil, resErr
	}
	defer response.Body.Close()

	svc.Client.Logger.DebugContext(ctx, "Listing Order Approvals", slog.String("url", path), slog.Int("status_code", response.StatusCode))

	var apiResponse listOrderApprovalsV4Response
	if err = json.Unmarshal(buf.Bytes(), &apiResponse); err != nil {
		return nil, err
	}

	toReturn := &ListOrderApprovalsResponse{
		OrderApprovals: apiResponse.Data.Items,
		TotalCount:     apiResponse.Data.TotalItems,
		Page:           apiResponse.Data.PageNumber,
		Limit:          apiResponse.Data.PageSize,
		TotalPages:     apiResponse.Data.TotalPages,
	}

	return toReturn, nil
}

// orderApprovalSort builds the v4 sort value, which carries the direction after a comma.
func orderApprovalSort(req *ListOrderApprovalsRequest) string {
	if req.Direction == nil {
		if req.Sort == nil {
			return ""
		}
		return *req.Sort
	}
	field := "createdAt"
	if req.Sort != nil {
		field = *req.Sort
	}
	return field + "," + *req.Direction
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
	if orderApprovalUID == "" {
		return ErrOrderApprovalUIDRequired
	}
	u := svc.Client.BaseURL.JoinPath("v3", "order_approvals", orderApprovalUID, action).String()

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPost, u, req)
	if err != nil {
		return err
	}
	response, resErr := svc.Client.Do(ctx, clientReq, io.Discard)
	if resErr != nil {
		return resErr
	}
	if response != nil {
		svc.Client.Logger.DebugContext(ctx, fmt.Sprintf("Order Approval %s", action), slog.String("url", u), slog.Int("status_code", response.StatusCode))
		defer response.Body.Close()
	}
	return nil
}
