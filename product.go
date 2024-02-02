package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type ProductService interface {
	ExecuteOrder(ctx context.Context, requestBody interface{}) (*[]byte, error)
	ModifyProduct(ctx context.Context, req *ModifyProductRequest) (*ModifyProductResponse, error)
	DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error)
	RestoreProduct(ctx context.Context, productId string) (*RestoreProductResponse, error)
	ManageProductLock(ctx context.Context, req *ManageProductLockRequest) (*ManageProductLockResponse, error)
	GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*PrefixFilterList, error)
	CreateMCRPrefixFilterList(ctx context.Context, req *CreateMCRPrefixFilterListRequest) (*CreateMCRPrefixFilterListResponse, error)
}

// ProductServiceOp handles communication with Product methods of the Megaport API.
type ProductServiceOp struct {
	Client *Client
}

func NewProductServiceOp(c *Client) *ProductServiceOp {
	return &ProductServiceOp{
		Client: c,
	}
}

type ModifyProductRequest struct {
	ProductID             string
	ProductType           string
	Name                  string `json:"name"`
	CostCentre            string `json:"costCentre"`
	MarketplaceVisibility bool   `json:"marketplaceVisibility"`
}

type ModifyProductResponse struct {
	IsUpdated bool
}
type DeleteProductRequest struct {
	ProductID string
	DeleteNow bool
}

type DeleteProductResponse struct{}

type RestoreProductRequest struct {
	ProductID string
}

type RestoreProductResponse struct{}

type ManageProductLockRequest struct {
	ProductID  string
	ShouldLock bool
}

type ManageProductLockResponse struct{}

type ParsedProductsResponse struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    []interface{} `json:"data"`
}

func (svc *ProductServiceOp) ExecuteOrder(ctx context.Context, requestBody interface{}) (*[]byte, error) {
	path := "/v3/networkdesign/buy"

	url := svc.Client.BaseURL.JoinPath(path).String()

	req, err := svc.Client.NewRequest(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		return nil, err
	}

	response, resErr := svc.Client.Do(ctx, req, nil)
	if resErr != nil {
		svc.Client.Logger.DebugContext(ctx, "", slog.String("trace-id", response.Header.Get("Trace-Id")))
		return nil, resErr
	}

	if response != nil {
		svc.Client.Logger.DebugContext(ctx, "Executing product order", "url", url, "status_code", response.StatusCode)
		defer response.Body.Close()
	}

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	return &body, nil
}

// ModifyProduct modifies a product. The available fields to modify are Name, Cost Centre, and Marketplace Visibility.
func (svc *ProductServiceOp) ModifyProduct(ctx context.Context, req *ModifyProductRequest) (*ModifyProductResponse, error) {

	if req.ProductType == PRODUCT_MEGAPORT || req.ProductType == PRODUCT_MCR {
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
		return nil, errors.New(ERR_WRONG_PRODUCT_MODIFY)
	}
}

// DeleteProduct is responsible for either scheduling a product for deletion "CANCEL" or deleting a product immediately
// "CANCEL_NOW".
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

// GetMCRPrefixFilterLists returns prefix filter lists for the specified MCR2.
func (svc *ProductServiceOp) GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*PrefixFilterList, error) {
	url := "/v2/product/mcr2/" + mcrId + "/prefixLists?"

	req, err := svc.Client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	prefixFilterList := &MCRPrefixFilterListResponse{}
	unmarshalErr := json.Unmarshal(body, prefixFilterList)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return prefixFilterList.Data, nil
}

// CreateMCRPrefixFilterList will create an MCR2 product prefix filter list.
func (svc *ProductServiceOp) CreateMCRPrefixFilterList(ctx context.Context, req *CreateMCRPrefixFilterListRequest) (*CreateMCRPrefixFilterListResponse, error) {
	url := "/v2/product/mcr2/" + req.MCRID + "/prefixList"

	clientReq, err := svc.Client.NewRequest(ctx, "POST", url, req.PrefixFilterList)

	if err != nil {
		return nil, err
	}

	_, err = svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}

	return &CreateMCRPrefixFilterListResponse{
		IsCreated: true,
	}, nil
}
