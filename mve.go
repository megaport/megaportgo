package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

// MVEService is an interface for interfacing with the MVE endpoints of the Megaport API.
type MVEService interface {
	// BuyMVE buys an MVE from the Megaport MVE API.
	BuyMVE(ctx context.Context, req *BuyMVERequest) (*BuyMVEResponse, error)
	// ValidateMVEOrder validates an MVE order in the Megaport Products API.
	ValidateMVEOrder(ctx context.Context, req *BuyMVERequest) error
	// GetMVE gets details about a single MVE from the Megaport MVE API.
	GetMVE(ctx context.Context, mveId string) (*MVE, error)
	// ModifyMVE modifies an MVE in the Megaport MVE API.
	ModifyMVE(ctx context.Context, req *ModifyMVERequest) (*ModifyMVEResponse, error)
	// DeleteMVE deletes an MVE in the Megaport MVE API.
	DeleteMVE(ctx context.Context, req *DeleteMVERequest) (*DeleteMVEResponse, error)
	// ListMVEImages returns a list of currently supported MVE images and details for each image, including image ID, version, product, and vendor. The image id returned indicates the software version and key configuration parameters of the image. The releaseImage value returned indicates whether the MVE image is available for selection when ordering an MVE.
	ListMVEImages(ctx context.Context) ([]*MVEImage, error)
	// ListAvailableMVESizes returns a list of currently available MVE sizes and details for each size. The instance size determines the MVE capabilities, such as how many concurrent connections it can support. The compute sizes are 2/8, 4/16, 8/32, and 12/48, where the first number is the CPU and the second number is the GB of available RAM. Each size has 4 GB of RAM for every vCPU allocated.
	ListAvailableMVESizes(ctx context.Context) ([]*MVESize, error)
}

// NewMVEService creates a new instance of the MVE Service.
func NewMVEService(c *Client) *MVEServiceOp {
	return &MVEServiceOp{
		Client: c,
	}
}

// MVEServiceOp handles communication with MVE methods of the Megaport API.
type MVEServiceOp struct {
	Client *Client
}

// BuyMVERequest represents a request to buy an MVE
type BuyMVERequest struct {
	LocationID    int
	Name          string
	Term          int
	VendorConfig  VendorConfig
	Vnics         []MVENetworkInterface
	DiversityZone string
	PromoCode     string
	CostCentre    string

	WaitForProvision bool          // Wait until the MVE provisions before returning
	WaitForTime      time.Duration // How long to wait for the MVE to provision if WaitForProvision is true (default is 5 minutes)
}

// BuyMVEResponse represents a response from buying an MVE
type BuyMVEResponse struct {
	TechnicalServiceUID string
}

// ModifyMVERequest represents a request to modify an MVE
type ModifyMVERequest struct {
	MVEID                 string
	Name                  string
	MarketplaceVisibility *bool
	CostCentre            string
	ContractTermMonths    *int // Contract term in months

	WaitForUpdate bool          // Wait until the MCVEupdates before returning
	WaitForTime   time.Duration // How long to wait for the MVE to update if WaitForUpdate is true (default is 5 minutes)
}

// ModifyMVEResponse represents a response from modifying an MVE
type ModifyMVEResponse struct {
	MVEUpdated bool
}

// DeleteMVERequest represents a request to delete an MVE
type DeleteMVERequest struct {
	MVEID string
}

// DeleteMVEResponse represents a response from deleting an MVE
type DeleteMVEResponse struct {
	IsDeleted bool
}

// BuyMVE buys an MVE from the Megaport MVE API.
func (svc *MVEServiceOp) BuyMVE(ctx context.Context, req *BuyMVERequest) (*BuyMVEResponse, error) {
	err := validateBuyMVERequest(req)
	if err != nil {
		return nil, err
	}

	mveOrder := createMVEOrder(req)

	resp, err := svc.Client.ProductService.ExecuteOrder(ctx, mveOrder)
	if err != nil {
		return nil, err
	}

	orderInfo := MVEOrderResponse{}

	if err := json.Unmarshal(*resp, &orderInfo); err != nil {
		return nil, err
	}

	toReturn := &BuyMVEResponse{
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
				return nil, fmt.Errorf("time expired waiting for MVE %s to provision", toReturn.TechnicalServiceUID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for MVE %s to provision", toReturn.TechnicalServiceUID)
			case <-ticker.C:
				mveDetails, err := svc.GetMVE(ctx, toReturn.TechnicalServiceUID)
				if err != nil {
					return nil, err
				}

				if slices.Contains(SERVICE_STATE_READY, mveDetails.ProvisioningStatus) {
					return toReturn, nil
				}

			}
		}
	} else {
		// return the service UID right away if the user doesn't want to wait for provision
		return toReturn, nil
	}
}

func (svc *MVEServiceOp) ValidateMVEOrder(ctx context.Context, req *BuyMVERequest) error {
	err := validateBuyMVERequest(req)
	if err != nil {
		return err
	}
	mveOrder := createMVEOrder(req)
	return svc.Client.ProductService.ValidateProductOrder(ctx, mveOrder)
}

func createMVEOrder(req *BuyMVERequest) []*MVEOrderConfig {
	order := &MVEOrderConfig{
		LocationID:   req.LocationID,
		Name:         req.Name,
		Term:         req.Term,
		PromoCode:    req.PromoCode,
		CostCentre:   req.CostCentre,
		VendorConfig: req.VendorConfig,
		ProductType:  strings.ToUpper(PRODUCT_MVE),
		Config: MVEConfig{
			DiversityZone: req.DiversityZone,
		},
	}

	if len(req.Vnics) == 0 {
		order.NetworkInterfaces = []MVENetworkInterface{{Description: "Data Plane", VLAN: 0}}
	} else {
		order.NetworkInterfaces = req.Vnics
	}

	mveOrder := []*MVEOrderConfig{order}
	return mveOrder
}

// GetMVE retrieves a single MVE from the Megaport MVE API.
func (svc *MVEServiceOp) GetMVE(ctx context.Context, mveId string) (*MVE, error) {
	path := "/v2/product/" + mveId
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	mveResp := MVEResponse{}
	if err := json.Unmarshal(body, &mveResp); err != nil {
		return nil, err
	}

	return mveResp.Data, nil
}

// ModifyMVE modifies an MVE in the Megaport MVE API.
func (svc *MVEServiceOp) ModifyMVE(ctx context.Context, req *ModifyMVERequest) (*ModifyMVEResponse, error) {
	modifyProductReq := &ModifyProductRequest{
		ProductID:             req.MVEID,
		ProductType:           PRODUCT_MVE,
		MarketplaceVisibility: PtrTo(false),
	}
	if req.Name != "" {
		modifyProductReq.Name = req.Name
	}
	if req.CostCentre != "" {
		modifyProductReq.CostCentre = req.CostCentre
	}
	if req.ContractTermMonths != nil {
		modifyProductReq.ContractTermMonths = *req.ContractTermMonths
	}

	_, err := svc.Client.ProductService.ModifyProduct(ctx, modifyProductReq)
	if err != nil {
		return nil, err
	}
	toReturn := &ModifyMVEResponse{
		MVEUpdated: true,
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
				return nil, fmt.Errorf("time expired waiting for MVE %s to update", req.MVEID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for MVE %s to update", req.MVEID)
			case <-ticker.C:
				mveDetails, err := svc.GetMVE(ctx, req.MVEID)
				if err != nil {
					return nil, err
				}
				if slices.Contains(SERVICE_STATE_READY, mveDetails.ProvisioningStatus) {
					return toReturn, nil
				}
			}
		}
	} else {
		// return the response right away if the user doesn't want to wait for update
		return toReturn, nil
	}
}

// DeleteMVE deletes an MVE in the Megaport MVE API.
func (svc *MVEServiceOp) DeleteMVE(ctx context.Context, req *DeleteMVERequest) (*DeleteMVEResponse, error) {
	_, err := svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID: req.MVEID,
		DeleteNow: true,
	})
	if err != nil {
		return nil, err
	}
	return &DeleteMVEResponse{IsDeleted: true}, nil
}

// ListMVEImages returns a list of currently supported MVE images and details for each image, including image ID, version, product, and vendor. The image id returned indicates the software version and key configuration parameters of the image. The releaseImage value returned indicates whether the MVE image is available for selection when ordering an MVE.
func (svc *MVEServiceOp) ListMVEImages(ctx context.Context) ([]*MVEImage, error) {
	path := "/v3/product/mve/images"
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	imageResp := MVEImageAPIResponse{}
	if err := json.Unmarshal(body, &imageResp); err != nil {
		return nil, err
	}
	return imageResp.Data.Images, nil
}

// ListAvailableMVESizes returns a list of currently available MVE sizes and details for each size. The instance size determines the MVE capabilities, such as how many concurrent connections it can support. The compute sizes are 2/8, 4/16, 8/32, and 12/48, where the first number is the CPU and the second number is the GB of available RAM. Each size has 4 GB of RAM for every vCPU allocated.
func (svc *MVEServiceOp) ListAvailableMVESizes(ctx context.Context) ([]*MVESize, error) {
	path := "/v3/product/mve/variants"
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	resp, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sizeResp := MVESizeAPIResponse{}
	if err := json.Unmarshal(body, &sizeResp); err != nil {
		return nil, err
	}
	return sizeResp.Data, nil
}

// validateBuyMVERequest validates a BuyMVERequest for proper term length.
func validateBuyMVERequest(req *BuyMVERequest) error {
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return ErrInvalidTerm
	}
	return nil
}
