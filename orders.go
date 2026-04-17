package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// OrderService is an interface for interacting with the Megaport Orders API.
//
// The Orders API is the mechanism for purchasing (provisioning) products
// that are created in a non-purchased design state. For example, NAT
// Gateways created via POST /v3/products/nat_gateways are returned in a
// DESIGN state and must be added to an Order and bought via the endpoints
// exposed by this service before they enter the normal provisioning
// lifecycle (DEPLOYABLE -> CONFIGURED -> LIVE).
//
// Typical usage:
//
//	gw, _ := client.NATGatewayService.CreateNATGateway(ctx, createReq)
//	order, _ := client.OrderService.CreateOrder(ctx, &CreateOrderRequest{
//	    Items:     []string{gw.ProductUID},
//	    Reference: "tf-" + gw.ProductUID,
//	})
//	if _, err := client.OrderService.BuyOrder(ctx, order.UID); err != nil {
//	    // handle
//	}
type OrderService interface {
	// CreateOrder creates a new Order referencing one or more product UIDs.
	CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error)
	// GetOrder retrieves an Order by its UID.
	GetOrder(ctx context.Context, orderUID string) (*Order, error)
	// UpdateOrder updates an existing Order's items and/or reference.
	UpdateOrder(ctx context.Context, orderUID string, req *UpdateOrderRequest) (*Order, error)
	// DeleteOrder deletes an Order. The API only allows deletion while the
	// Order state is NEW.
	DeleteOrder(ctx context.Context, orderUID string) error
	// ValidateOrder validates an Order without purchasing it.
	ValidateOrder(ctx context.Context, orderUID string) (*Order, error)
	// BuyOrder purchases (provisions) an Order, causing all items in the
	// Order to begin provisioning.
	BuyOrder(ctx context.Context, orderUID string) (*Order, error)
}

// OrderServiceOp handles communication with Order methods of the Megaport API.
type OrderServiceOp struct {
	Client *Client
}

// NewOrderService creates a new instance of the Order Service.
func NewOrderService(c *Client) *OrderServiceOp {
	return &OrderServiceOp{Client: c}
}

// CreateOrderRequest is the payload for POST /v3/orders.
type CreateOrderRequest struct {
	// Items is a list of product UIDs to include in the order.
	Items []string `json:"items"`
	// Reference is an optional human-readable reference, e.g. "ORD-123456".
	Reference string `json:"reference,omitempty"`
}

// UpdateOrderRequest is the payload for PUT /v3/orders/{orderUid}.
type UpdateOrderRequest struct {
	// Items is a list of product UIDs to include in the order.
	Items []string `json:"items"`
	// Reference is an optional human-readable reference, e.g. "ORD-123456".
	Reference string `json:"reference,omitempty"`
}

// Order represents an Order resource returned by the Megaport API.
type Order struct {
	UID        string   `json:"uid"`
	CompanyUID string   `json:"companyUid"`
	CreatedBy  string   `json:"createdBy"`
	Items      []string `json:"items"`
	Reference  string   `json:"reference"`
	State      string   `json:"state"`
}

// orderResponse is the API response envelope for a single Order.
type orderResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    Order  `json:"data"`
}

// ErrOrderUIDRequired is returned when an Order UID is not provided.
var ErrOrderUIDRequired = errors.New("order UID is required")

// ErrOrderItemsRequired is returned when no items are provided on create/update.
var ErrOrderItemsRequired = errors.New("at least one order item is required")

func validateCreateOrderRequest(req *CreateOrderRequest) error {
	if req == nil || len(req.Items) == 0 {
		return ErrOrderItemsRequired
	}
	return nil
}

func validateUpdateOrderRequest(orderUID string, req *UpdateOrderRequest) error {
	if orderUID == "" {
		return ErrOrderUIDRequired
	}
	if req == nil || len(req.Items) == 0 {
		return ErrOrderItemsRequired
	}
	return nil
}

// CreateOrder creates a new Order.
func (svc *OrderServiceOp) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*Order, error) {
	if err := validateCreateOrderRequest(req); err != nil {
		return nil, err
	}
	return svc.doOrderRequest(ctx, http.MethodPost, "/v3/orders", req)
}

// GetOrder retrieves an Order by UID.
func (svc *OrderServiceOp) GetOrder(ctx context.Context, orderUID string) (*Order, error) {
	if orderUID == "" {
		return nil, ErrOrderUIDRequired
	}
	path := fmt.Sprintf("/v3/orders/%s", url.PathEscape(orderUID))
	return svc.doOrderRequest(ctx, http.MethodGet, path, nil)
}

// UpdateOrder updates an existing Order.
func (svc *OrderServiceOp) UpdateOrder(ctx context.Context, orderUID string, req *UpdateOrderRequest) (*Order, error) {
	if err := validateUpdateOrderRequest(orderUID, req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/orders/%s", url.PathEscape(orderUID))
	return svc.doOrderRequest(ctx, http.MethodPut, path, req)
}

// DeleteOrder deletes an Order. The API only allows deletion while the
// Order state is NEW.
func (svc *OrderServiceOp) DeleteOrder(ctx context.Context, orderUID string) error {
	if orderUID == "" {
		return ErrOrderUIDRequired
	}
	path := fmt.Sprintf("/v3/orders/%s", url.PathEscape(orderUID))
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	resp, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// ValidateOrder validates an Order without purchasing it.
func (svc *OrderServiceOp) ValidateOrder(ctx context.Context, orderUID string) (*Order, error) {
	if orderUID == "" {
		return nil, ErrOrderUIDRequired
	}
	path := fmt.Sprintf("/v3/orders/%s/validate", url.PathEscape(orderUID))
	return svc.doOrderRequest(ctx, http.MethodPost, path, nil)
}

// BuyOrder purchases (provisions) an Order.
func (svc *OrderServiceOp) BuyOrder(ctx context.Context, orderUID string) (*Order, error) {
	if orderUID == "" {
		return nil, ErrOrderUIDRequired
	}
	path := fmt.Sprintf("/v3/orders/%s/buy", url.PathEscape(orderUID))
	return svc.doOrderRequest(ctx, http.MethodPost, path, nil)
}

func (svc *OrderServiceOp) doOrderRequest(ctx context.Context, method, path string, body interface{}) (*Order, error) {
	clientReq, err := svc.Client.NewRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	resp, err := svc.Client.Do(ctx, clientReq, &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var envelope orderResponse
	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		return nil, err
	}
	return &envelope.Data, nil
}
