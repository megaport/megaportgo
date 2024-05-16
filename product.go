package megaport

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// ProductService is an interface for interfacing with the Product endpoints of the Megaport API.
type ProductService interface {
	// ExecuteOrder is responsible for executing an order for a product in the Megaport Products API.
	ExecuteOrder(ctx context.Context, requestBody interface{}) (*[]byte, error)
	// ModifyProduct modifies a product in the Megaport Products API. The available fields to modify are Name, Cost Centre, and Marketplace Visibility.
	ModifyProduct(ctx context.Context, req *ModifyProductRequest) (*ModifyProductResponse, error)
	// DeleteProduct is responsible for either scheduling a product for deletion "CANCEL" or deleting a product immediately "CANCEL_NOW" in the Megaport Products API.
	DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error)
	// RestoreProduct is responsible for restoring a product in the Megaport Products API. The product must be in a "CANCELLED" state to be restored.
	RestoreProduct(ctx context.Context, productId string) (*RestoreProductResponse, error)
	// ManageProductLock is responsible for locking or unlocking a product in the Megaport Products API.
	ManageProductLock(ctx context.Context, req *ManageProductLockRequest) (*ManageProductLockResponse, error)
}

// ProductServiceOp handles communication with Product methods of the Megaport API.
type ProductServiceOp struct {
	Client *Client
}

// NewProductService creates a new instance of the Product Service.
func NewProductService(c *Client) *ProductServiceOp {
	return &ProductServiceOp{
		Client: c,
	}
}

// ModifyProductRequest represents a request to modify a product in the Megaport Products API.
type ModifyProductRequest struct {
	ProductID             string
	ProductType           string
	Name                  string `json:"name,omitempty"`
	CostCentre            string `json:"costCentre,omitempty"`
	MarketplaceVisibility *bool  `json:"marketplaceVisibility,omitempty"`
}

// ModifyProductResponse represents a response from the Megaport Products API after modifying a product.
type ModifyProductResponse struct {
	IsUpdated bool
}

// DeleteProductRequest represents a request to delete a product in the Megaport Products API.
type DeleteProductRequest struct {
	ProductID string
	DeleteNow bool
}

// DeleteProductResponse represents a response from the Megaport Products API after deleting a product.
type DeleteProductResponse struct{}

// RestoreProductRequest represents a request to restore a product in the Megaport Products API.
type RestoreProductRequest struct {
	ProductID string
}

// RestoreProductResponse represents a response from the Megaport Products API after restoring a product.
type RestoreProductResponse struct{}

// ManageProductLockRequest represents a request to lock or unlock a product in the Megaport Products API.
type ManageProductLockRequest struct {
	ProductID  string
	ShouldLock bool
}

// ManageProductLockResponse represents a response from the Megaport Products API after locking or unlocking a product.
type ManageProductLockResponse struct{}

// ParsedProductsResponse represents a response from the Megaport Products API prior to parsing the response.
type ParsedProductsResponse struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    []interface{} `json:"data"`
}

// ExecuteOrder is responsible for executing an order for a product in the Megaport Products API.
func (svc *ProductServiceOp) ExecuteOrder(ctx context.Context, requestBody interface{}) (*[]byte, error) {
	path := "/v3/networkdesign/buy"

	url := svc.Client.BaseURL.JoinPath(path).String()

	req, err := svc.Client.NewRequest(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		return nil, err
	}

	response, resErr := svc.Client.Do(ctx, req, nil)
	if resErr != nil {
		return nil, resErr
	}

	if response != nil {
		svc.Client.Logger.DebugContext(ctx, "Executing product order", slog.String("url", url), slog.Int("status_code", response.StatusCode))
		defer response.Body.Close()
	}

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	return &body, nil
}

// ModifyProduct modifies a product in the Megaport Products API. The available fields to modify are Name, Cost Centre, and Marketplace Visibility.
func (svc *ProductServiceOp) ModifyProduct(ctx context.Context, req *ModifyProductRequest) (*ModifyProductResponse, error) {
	if req.ProductType == PRODUCT_MEGAPORT || req.ProductType == PRODUCT_MCR || req.ProductType == PRODUCT_MVE {
		path := fmt.Sprintf("/v2/product/%s/%s", req.ProductType, req.ProductID)
		url := svc.Client.BaseURL.JoinPath(path).String()

		req, err := svc.Client.NewRequest(ctx, http.MethodPut, url, req)

		if err != nil {
			return nil, err
		}

		_, err = svc.Client.Do(ctx, req, nil)
		if err != nil {
			return nil, err
		}
		return &ModifyProductResponse{IsUpdated: true}, nil
	} else {
		return nil, ErrWrongProductModify
	}
}

// DeleteProduct is responsible for either scheduling a product for deletion "CANCEL" or deleting a product immediately "CANCEL_NOW" in the Megaport Products API.
func (svc *ProductServiceOp) DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error) {
	var action string

	if req.DeleteNow {
		action = "CANCEL_NOW"
	} else {
		action = "CANCEL"
	}

	path := "/v3/product/" + req.ProductID + "/action/" + action
	url := svc.Client.BaseURL.JoinPath(path).String()

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	_, err = svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	return &DeleteProductResponse{}, nil
}

// RestoreProduct is responsible for restoring a product in the Megaport Products API. The product must be in a "CANCELLED" state to be restored.
func (svc *ProductServiceOp) RestoreProduct(ctx context.Context, productId string) (*RestoreProductResponse, error) {
	path := "/v3/product/" + productId + "/action/UN_CANCEL"
	url := svc.Client.BaseURL.JoinPath(path).String()
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	_, err = svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}

	return &RestoreProductResponse{}, nil
}

// ManageProductLock is responsible for locking or unlocking a product in the Megaport Products API.
func (svc *ProductServiceOp) ManageProductLock(ctx context.Context, req *ManageProductLockRequest) (*ManageProductLockResponse, error) {
	verb := "POST"

	if !req.ShouldLock {
		verb = "DELETE"
	}

	path := fmt.Sprintf("/v2/product/%s/lock", req.ProductID)
	url := svc.Client.BaseURL.JoinPath(path).String()

	clientReq, err := svc.Client.NewRequest(ctx, verb, url, nil)
	if err != nil {
		return nil, err
	}

	_, err = svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	return &ManageProductLockResponse{}, nil
}
