package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// ProductService is an interface for interfacing with the Product endpoints of the Megaport API.
type ProductService interface {
	// ExecuteOrder is responsible for executing an order for a product in the Megaport Products API.
	ExecuteOrder(ctx context.Context, requestBody interface{}) (*[]byte, error)
	// ListProducts retrieves a list of products from the Megaport Products API. It returns a slice of Product interfaces, which can be of different types (Port, MCR, MVE). The function handles the parsing of the response and unmarshals it into the appropriate product type based on the product type field.
	ListProducts(ctx context.Context) ([]Product, error)
	// ModifyProduct modifies a product in the Megaport Products API. The available fields to modify are Name, Cost Centre, and Marketplace Visibility.
	ModifyProduct(ctx context.Context, req *ModifyProductRequest) (*ModifyProductResponse, error)
	// DeleteProduct is responsible for either scheduling a product for deletion "CANCEL" or deleting a product immediately "CANCEL_NOW" in the Megaport Products API.
	DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error)
	// RestoreProduct is responsible for restoring a product in the Megaport Products API. The product must be in a "CANCELLED" state to be restored.
	RestoreProduct(ctx context.Context, productId string) (*RestoreProductResponse, error)
	// ManageProductLock is responsible for locking or unlocking a product in the Megaport Products API.
	ManageProductLock(ctx context.Context, req *ManageProductLockRequest) (*ManageProductLockResponse, error)
	// ValidateProductOrder is responsible for validating an order for a product in the Megaport Products API.
	ValidateProductOrder(ctx context.Context, requestBody interface{}) error
	// ListProductResourceTags is responsible for retrieving the resource tags for a product in the Megaport Products API.
	ListProductResourceTags(ctx context.Context, productID string) ([]ResourceTag, error)
	// UpdateProductResourceTags is responsible for updating the resource tags for a product in the Megaport Products API.
	UpdateProductResourceTags(ctx context.Context, productUID string, tagsReq *UpdateProductResourceTagsRequest) error
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
	ContractTermMonths    int    `json:"term,omitempty"`
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
	Message string            `json:"message"`
	Terms   string            `json:"terms"`
	Data    []json.RawMessage `json:"data"`
}

// Product defines the common interface for all Megaport products
type Product interface {
	GetType() string
	GetUID() string
	GetProvisioningStatus() string
}

type ParsedProduct struct {
	Type string `json:"productType"`
}

// ResourceTagsResponse represents a response from the Megaport Products API after retrieving the resource tags for a product.
type ResourceTagsResponse struct {
	Message string                    `json:"message"`
	Terms   string                    `json:"terms"`
	Data    *ResourceTagsResponseData `json:"data"`
}

type ResourceTagsResponseData struct {
	ResourceTags []ResourceTag `json:"resourceTags"`
}

type UpdateProductResourceTagsRequest struct {
	ResourceTags []ResourceTag `json:"resourceTags"`
}

type ResourceTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
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

// ListProducts retrieves a list of products from the Megaport Products API.
// It returns a slice of Product interfaces, which can be of different types (Port, MCR, MVE).
// The function handles the parsing of the response and unmarshals it into the appropriate product type based on the product type field.
// It also logs any errors encountered during the unmarshalling process.
func (svc *ProductServiceOp) ListProducts(ctx context.Context) ([]Product, error) {
	path := "/v2/products"
	url := svc.Client.BaseURL.JoinPath(path).String()
	req, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// Parse response into a structure with raw JSON messages
	var parsed ParsedProductsResponse

	if err := json.NewDecoder(response.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	products := []Product{}

	for i, rawProduct := range parsed.Data {
		// First extract just the type field
		var parsedProduct ParsedProduct

		if err := json.Unmarshal(rawProduct, &parsedProduct); err != nil {
			svc.Client.Logger.WarnContext(ctx, fmt.Sprintf("Item %d: Could not extract product type: %v", i, err))
			continue
		}

		// Then unmarshal into the appropriate struct based on type
		switch strings.ToLower(parsedProduct.Type) {
		case PRODUCT_MEGAPORT:
			var port Port
			if err := json.Unmarshal(rawProduct, &port); err != nil {
				svc.Client.Logger.WarnContext(ctx, fmt.Sprintf("Item %d: Could not unmarshal as PORT: %v", i, err))
				continue
			}
			products = append(products, &port)
		case PRODUCT_MCR:
			var mcr MCR
			if err := json.Unmarshal(rawProduct, &mcr); err != nil {
				svc.Client.Logger.WarnContext(ctx, fmt.Sprintf("Item %d: Could not unmarshal as MCR: %v", i, err))
				continue
			}
			products = append(products, &mcr)
		case PRODUCT_MVE:
			var mve MVE
			if err := json.Unmarshal(rawProduct, &mve); err != nil {
				svc.Client.Logger.WarnContext(ctx, fmt.Sprintf("Item %d: Could not unmarshal as MVE: %v", i, err))
				continue
			}
			products = append(products, &mve)
		}
	}
	return products, nil
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

// ValidateProductOrder is responsible for validating an order for a product in the Megaport Products API.
func (svc *ProductServiceOp) ValidateProductOrder(ctx context.Context, requestBody interface{}) error {
	path := "/v3/networkdesign/validate"

	url := svc.Client.BaseURL.JoinPath(path).String()

	req, err := svc.Client.NewRequest(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		return err
	}

	_, resErr := svc.Client.Do(ctx, req, nil)
	if resErr != nil {
		return resErr
	}

	return nil
}

// ListProductResourceTags is responsible for retrieving the resource tags for a product in the Megaport Products API.
func (svc *ProductServiceOp) ListProductResourceTags(ctx context.Context, productUID string) ([]ResourceTag, error) {
	path := fmt.Sprintf("/v2/product/%s/tags", productUID)
	url := svc.Client.BaseURL.JoinPath(path).String()
	req, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	tagsResponse := &ResourceTagsResponse{}
	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	err = json.Unmarshal(body, tagsResponse)
	if err != nil {
		return nil, err
	}

	return tagsResponse.Data.ResourceTags, nil
}

// UpdateProductResourceTags is responsible for updating the resource tags for a product in the Megaport Products API.
func (svc *ProductServiceOp) UpdateProductResourceTags(ctx context.Context, productUID string, tagsReq *UpdateProductResourceTagsRequest) error {
	path := fmt.Sprintf("/v2/product/%s/tags", productUID)
	url := svc.Client.BaseURL.JoinPath(path).String()
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPut, url, tagsReq)
	if err != nil {
		return err
	}

	_, err = svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return err
	}

	return nil
}

func toProductResourceTags(in map[string]string) []ResourceTag {
	tags := make([]ResourceTag, 0, len(in))
	for key, value := range in {
		tags = append(tags, ResourceTag{Key: key, Value: value})
	}
	return tags
}

func fromProductResourceTags(in []ResourceTag) map[string]string {
	tags := make(map[string]string, len(in))
	for _, tag := range in {
		tags[tag.Key] = tag.Value
	}
	return tags
}
