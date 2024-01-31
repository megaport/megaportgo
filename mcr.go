package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
)

// MCRService is an interface for interfacing with the MCR endpoints
// of the Megaport API.

type MCRService interface {
	BuyMCR(ctx context.Context, req *BuyMCRRequest) (*BuyMCRResponse, error)
	GetMCR(ctx context.Context, req *GetMCRRequest) (*MCR, error)
	CreatePrefixFilterList(ctx context.Context, req *CreateMCRPrefixFilterListRequest)
	ModifyMCR(ctx context.Context, req *ModifyMCRRequest) (*ModifyMCRResponse, error)
	DeleteMCR(ctx context.Context, req *DeleteMCRRequest) (*DeleteMCRResponse, error)
	RestoreMCR(ctx context.Context, req *RestoreMCRRequest) (*RestoreMCRResponse, error)
}

// ProductServiceOp handles communication with Product methods of the Megaport API.
type MCRServiceOp struct {
	Client *Client
}

func NewMCRServiceOp(c *Client) *MCRServiceOp {
	return &MCRServiceOp{
		Client: c,
	}
}

type BuyMCRRequest struct {
	BuyMCROrders []*BuyMCROrder
}

type BuyMCROrder struct {
	LocationID int
	Name       string
	Term       int
	PortSpeed  int
	MCRAsn     int
	Type       string
}

type BuyMCRResponse struct {
	MCROrderConfirmations []*MCROrderConfirmation
}

type GetMCRRequest struct {
	MCRID string
}

type CreateMCRPrefixFilterListRequest struct {
	MCRID            string
	PrefixFilterList MCRPrefixFilterList
}

type CreateMCRPrefixFilterListResponse struct {
	IsCreated bool
}

type ModifyMCRRequest struct {
	MCRID                 string
	Name                  string
	CostCentre            string
	MarketplaceVisibility bool
}

type ModifyMCRResponse struct {
	IsUpdated bool
}

type DeleteMCRRequest struct {
	MCRID     string
	DeleteNow bool
}

type DeleteMCRResponse struct {
	IsDeleting bool
}

type RestoreMCRRequest struct {
	MCRID string
}

type RestoreMCRResponse struct {
	IsRestored bool
}

// BuyMCR purchases an MCR.
func (svc *MCRServiceOp) BuyMCR(ctx context.Context, req *BuyMCRRequest) (*BuyMCRResponse, error) {
	mcrOrders := []*MCROrder{}

	for _, order := range req.BuyMCROrders {
		err := validateMCROrder(order)
		if err != nil {
			return nil, err
		}
		toAppend := &MCROrder{
			Type:       "MCR2",
			LocationID: order.LocationID,
			Name:       order.Name,
			Term:       order.Term,
			PortSpeed:  order.PortSpeed,
			Config:     MCROrderConfig{},
		}
		if order.MCRAsn != 0 {
			toAppend.Config.ASN = order.MCRAsn
		}
		mcrOrders = append(mcrOrders, toAppend)
	}

	requestBody, marshalErr := json.Marshal(mcrOrders)

	if marshalErr != nil {
		return nil, marshalErr
	}

	body, resErr := svc.Client.ProductService.ExecuteOrder(ctx, requestBody)

	if resErr != nil {
		return nil, resErr
	}

	orderInfo := MCROrderResponse{}
	unmarshalErr := json.Unmarshal(*body, &orderInfo)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	toReturn := &BuyMCRResponse{}

	for _, confirmation := range orderInfo.Data {
		toReturn.MCROrderConfirmations = append(toReturn.MCROrderConfirmations, &MCROrderConfirmation{
			TechnicalServiceUID: confirmation.TechnicalServiceUID,
		})
	}

	return toReturn, nil
}

func validateMCROrder(order *BuyMCROrder) error {
	if order.Term != 1 && order.Term != 12 && order.Term != 24 && order.Term != 36 {
		return errors.New(ERR_TERM_NOT_VALID)
	}
	if order.PortSpeed != 1000 && order.PortSpeed != 2500 && order.PortSpeed != 5000 && order.PortSpeed != 10000 {
		return errors.New(ERR_MCR_INVALID_PORT_SPEED)
	}
	return nil
}

func (svc *MCRServiceOp) GetMCR(ctx context.Context, req *GetMCRRequest) (*MCR, error) {
	url := "/v2/product/" + req.MCRID
	clientReq, err := svc.Client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, fileErr := io.ReadAll(response.Body)

	if fileErr != nil {
		return nil, fileErr
	}

	mcrRes := &MCRResponse{}
	unmarshalErr := json.Unmarshal(body, mcrRes)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return mcrRes.Data, nil
}

// CreatePrefixFilterList creates a Prefix Filter List on an MCR.
func (svc *MCRServiceOp) CreatePrefixFilterList(ctx context.Context, req *CreateMCRPrefixFilterListRequest) (*CreateMCRPrefixFilterListResponse, error) {
	res, err := svc.Client.ProductService.CreateMCRPrefixFilterList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (svc *MCRServiceOp) ModifyMCR(ctx context.Context, req *ModifyMCRRequest) (*ModifyMCRResponse, error) {
	modifyReq := &ModifyProductRequest{
		ProductID:             req.MCRID,
		ProductType:           PRODUCT_MCR,
		Name:                  req.Name,
		CostCentre:            req.CostCentre,
		MarketplaceVisibility: req.MarketplaceVisibility,
	}
	_, err := svc.Client.ProductService.ModifyProduct(ctx, modifyReq)
	if err != nil {
		return nil, err
	}
	return &ModifyMCRResponse{
		IsUpdated: true,
	}, nil
}

func (svc *MCRServiceOp) DeleteMCR(ctx context.Context, req *DeleteMCRRequest) (*DeleteMCRResponse, error) {
	_, err := svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID: req.MCRID,
		DeleteNow: req.DeleteNow,
	})
	if err != nil {
		return nil, err
	}
	return &DeleteMCRResponse{
		IsDeleting: true,
	}, nil
}

func (svc *MCRServiceOp) RestoreMCR(ctx context.Context, req *RestoreMCRRequest) (*RestoreMCRResponse, error) {
	_, err := svc.Client.ProductService.RestoreProduct(ctx, &RestoreProductRequest{
		ProductID: req.MCRID,
	})
	if err != nil {
		return nil, err
	}
	return &RestoreMCRResponse{
		IsRestored: true,
	}, nil
}
