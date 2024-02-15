package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"
)

// MCRService is an interface for interfacing with the MCR endpoints
// of the Megaport API.

type MCRService interface {
	BuyMCR(ctx context.Context, req *BuyMCRRequest) (*BuyMCRResponse, error)
	GetMCR(ctx context.Context, mcrId string) (*MCR, error)
	CreatePrefixFilterList(ctx context.Context, req *CreateMCRPrefixFilterListRequest) (*CreateMCRPrefixFilterListResponse, error)
	GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*PrefixFilterList, error)
	ModifyMCR(ctx context.Context, req *ModifyMCRRequest) (*ModifyMCRResponse, error)
	DeleteMCR(ctx context.Context, req *DeleteMCRRequest) (*DeleteMCRResponse, error)
	RestoreMCR(ctx context.Context, mcrId string) (*RestoreMCRResponse, error)
}

// MCRServiceOp handles communication with MCR methods of the Megaport API.
type MCRServiceOp struct {
	Client *Client
}

func NewMCRService(c *Client) *MCRServiceOp {
	return &MCRServiceOp{
		Client: c,
	}
}

type BuyMCRRequest struct {
	LocationID    int
	Name          string
	DiversityZone string
	Term          int
	PortSpeed     int
	MCRAsn        int

	WaitForProvision bool          // Wait until the MCR provisions before returning
	WaitForTime      time.Duration // How long to wait for the MCR to provision if WaitForProvision is true (default is 5 minutes)
}

type BuyMCRResponse struct {
	TechnicalServiceUID string
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

	WaitForUpdate bool          // Wait until the MCR updates before returning
	WaitForTime   time.Duration // How long to wait for the MCR to update if WaitForUpdate is true (default is 5 minutes)
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

type RestoreMCRResponse struct {
	IsRestored bool
}

// BuyMCR purchases an MCR.
func (svc *MCRServiceOp) BuyMCR(ctx context.Context, req *BuyMCRRequest) (*BuyMCRResponse, error) {
	err := validateBuyMCRRequest(req)
	if err != nil {
		return nil, err
	}

	order := MCROrder{
		LocationID:    req.LocationID,
		Name:          req.Name,
		Term:          req.Term,
		DiversityZone: req.DiversityZone,
		Type:          "MCR2",
		PortSpeed:     req.PortSpeed,
		Config:        MCROrderConfig{},
	}

	order.Config.ASN = req.MCRAsn

	mcrOrders := []MCROrder{
		order,
	}

	body, resErr := svc.Client.ProductService.ExecuteOrder(ctx, mcrOrders)

	if resErr != nil {
		return nil, resErr
	}

	orderInfo := MCROrderResponse{}
	unmarshalErr := json.Unmarshal(*body, &orderInfo)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	toReturn := &BuyMCRResponse{
		TechnicalServiceUID: orderInfo.Data[0].TechnicalServiceUID,
	}

	// wait until the MCR is provisioned before returning if requested by the user
	if req.WaitForProvision {
		toWait := req.WaitForTime
		if toWait == 0 {
			toWait = 5 * time.Minute
		}

		ticker := time.NewTicker(30 * time.Second) // check on the provision status every 30 seconds
		timer := time.NewTimer(toWait)
		defer ticker.Stop()
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				return nil, fmt.Errorf("time expired waiting for MCR %s to provision", toReturn.TechnicalServiceUID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for MCR %s to provision", toReturn.TechnicalServiceUID)
			case <-ticker.C:
				mcrDetails, err := svc.GetMCR(ctx, toReturn.TechnicalServiceUID)
				if err != nil {
					return nil, err
				}

				if slices.Contains(SERVICE_STATE_READY, mcrDetails.ProvisioningStatus) {
					return toReturn, nil
				}

			}
		}
	} else {
		// return the service UID right away if the user doesn't want to wait for provision
		return toReturn, nil
	}
}

func validateBuyMCRRequest(order *BuyMCRRequest) error {
	if order.Term != 1 && order.Term != 12 && order.Term != 24 && order.Term != 36 {
		return ErrInvalidTerm
	}
	if order.PortSpeed != 1000 && order.PortSpeed != 2500 && order.PortSpeed != 5000 && order.PortSpeed != 10000 {
		return ErrMCRInvalidPortSpeed
	}
	return nil
}

func (svc *MCRServiceOp) GetMCR(ctx context.Context, mcrId string) (*MCR, error) {
	url := "/v2/product/" + mcrId
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

// GetMCRPrefixFilterLists returns prefix filter lists for the specified MCR2.
func (svc *MCRServiceOp) GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*PrefixFilterList, error) {
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
	toReturn := &ModifyMCRResponse{
		IsUpdated: true,
	}

	// wait until the MCR is updated before returning if requested by the user
	if req.WaitForUpdate {
		toWait := req.WaitForTime
		if toWait == 0 {
			toWait = 5 * time.Minute
		}

		ticker := time.NewTicker(30 * time.Second) // check on the update status every 30 seconds
		timer := time.NewTimer(toWait)
		defer ticker.Stop()
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				return nil, fmt.Errorf("time expired waiting for MCR %s to update", req.MCRID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for MCR %s to update", req.MCRID)
			case <-ticker.C:
				mcrDetails, err := svc.GetMCR(ctx, req.MCRID)
				if err != nil {
					return nil, err
				}
				if mcrDetails.Name == req.Name && mcrDetails.CostCentre == req.CostCentre && mcrDetails.MarketplaceVisibility == req.MarketplaceVisibility && mcrDetails.ProvisioningStatus == "LIVE" {
					return toReturn, nil
				}
			}
		}
	} else {
		// return the response right away if the user doesn't want to wait for update
		return toReturn, nil
	}
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

func (svc *MCRServiceOp) RestoreMCR(ctx context.Context, mcrId string) (*RestoreMCRResponse, error) {
	_, err := svc.Client.ProductService.RestoreProduct(ctx, mcrId)
	if err != nil {
		return nil, err
	}
	return &RestoreMCRResponse{
		IsRestored: true,
	}, nil
}
