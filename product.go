package megaport

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
)

type ProductService interface {
	ExecuteOrder(ctx context.Context, requestBody interface{}) (*[]byte, error)
	ModifyProduct(ctx context.Context, req *ModifyProductRequest) (*ModifyProductResponse, error)
	DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error)
	RestoreProduct(ctx context.Context, req *RestoreProductRequest) (*RestoreProductResponse, error)
	ManageProductLock(ctx context.Context, req *ManageProductLockRequest) (*ManageProductLockResponse, error)
}

// ProductServiceOp handles communication with Product methods of the Megaport API.
type ProductServiceOp struct {
	Client *Client
}

type ModifyProductRequest struct {
	ProductID             string
	ProductType           string
	Name                  string
	CostCentre            string
	MarketplaceVisibility bool
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

func NewProductServiceOp(c *Client) *ProductServiceOp {
	return &ProductServiceOp{
		Client: c,
	}
}

func (svc *ProductServiceOp) ExecuteOrder(ctx context.Context, requestBody interface{}) (*[]byte, error) {
	path := "/v3/networkdesign/buy"

	url := svc.Client.BaseURL.JoinPath(path).String()

	req, err := svc.Client.NewRequest(ctx, http.MethodPost, url, requestBody)
	if err != nil {
		return nil, err
	}

	response, resErr := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, resErr
	}

	if response != nil {
		svc.Client.Logger.Debug("Executing product order", "url", url, "status_code", response.StatusCode)
		defer response.Body.Close()
	}

	isError, parsedError := svc.Client.IsErrorResponse(response, &resErr, 200)

	if isError {
		return nil, parsedError
	}

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	return &body, nil
}

// ModifyProduct modifies a product. The available fields to modify are Name, Cost Centre, and Marketplace Visibility.
func (svc *ProductServiceOp) ModifyProduct(ctx context.Context, req *ModifyProductRequest) (*ModifyProductResponse, error) {

	if req.ProductType == types.PRODUCT_MEGAPORT || req.ProductType == types.PRODUCT_MCR {
		update := types.ProductUpdate{
			Name:                 req.Name,
			CostCentre:           req.CostCentre,
			MarketplaceVisbility: req.MarketplaceVisibility,
		}
		path := fmt.Sprintf("/v2/product/%s/%s", req.ProductType, req.ProductID)
		url := svc.Client.BaseURL.JoinPath(path).String()

		req, err := svc.Client.NewRequest(ctx, http.MethodPut, url, update)

		if err != nil {
			return nil, err
		}

		updateResponse, err := svc.Client.Do(ctx, req, nil)

		isResErr, compiledResErr := svc.Client.IsErrorResponse(updateResponse, &err, 200)

		if isResErr {
			return nil, compiledResErr
		} else {
			return &ModifyProductResponse{IsUpdated: true}, nil
		}
	} else {
		return nil, errors.New(mega_err.ERR_WRONG_PRODUCT_MODIFY)
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

	deleteResp, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer deleteResp.Body.Close() // nolint

	isError, errorMessage := svc.Client.IsErrorResponse(deleteResp, &err, 200)
	if isError {
		return nil, errorMessage
	} else {
		return &DeleteProductResponse{}, nil
	}
}

func (svc *ProductServiceOp) RestoreProduct(ctx context.Context, req *RestoreProductRequest) (*RestoreProductResponse, error) {
	path := "/v2/product/" + req.ProductID + "/action/UN_CANCEL"
	url := svc.Client.BaseURL.JoinPath(path).String()
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close() // nolint

	isError, errorMessage := svc.Client.IsErrorResponse(response, &err, 200)
	if isError {
		return nil, errorMessage
	} else {
		return &RestoreProductResponse{}, nil
	}
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

	lockResponse, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	isResErr, compiledResErr := svc.Client.IsErrorResponse(lockResponse, &err, 200)
	if isResErr {
		return nil, compiledResErr
	} else {
		return &ManageProductLockResponse{}, nil
	}
}
