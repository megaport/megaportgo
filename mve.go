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
	// ListMVEs lists all MVEs in the Megaport API.
	ListMVEs(ctx context.Context, req *ListMVEsRequest) ([]*MVE, error)
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
	// ListMVEResourceTags returns a list of resource tags for an MVE in the Megaport MVE API.
	ListMVEResourceTags(ctx context.Context, mveID string) (map[string]string, error)
	// UpdateMVEResourceTags updates the resource tags for an MVE in the Megaport MVE API.
	UpdateMVEResourceTags(ctx context.Context, mveID string, tags map[string]string) error
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

	ResourceTags map[string]string `json:"resourceTags,omitempty"`

	WaitForProvision bool          // Wait until the MVE provisions before returning
	WaitForTime      time.Duration // How long to wait for the MVE to provision if WaitForProvision is true (default is 5 minutes)
}

// BuyMVEResponse represents a response from buying an MVE
type BuyMVEResponse struct {
	TechnicalServiceUID string
}

// ListMVEsRequest represents a request to list MVEs. It allows you to determine whether to include inactive MVEs in the response. The default is to include only active MVEs.
type ListMVEsRequest struct {
	IncludeInactive bool
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
	MVEID      string
	SafeDelete bool // If true, the API will check whether the MVE has any attached resources before deleting it. If the MVE has attached resources, the API will return an error.
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

	orderInfo := mveOrderResponse{}

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
		ResourceTags: toProductResourceTags(req.ResourceTags),
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

// ListMVEs lists all MVEs in the Megaport API.
func (svc *MVEServiceOp) ListMVEs(ctx context.Context, req *ListMVEsRequest) ([]*MVE, error) {
	allProducts, err := svc.Client.ProductService.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	mves := []*MVE{}

	for _, product := range allProducts {
		if strings.ToLower(product.GetType()) == PRODUCT_MVE {
			mve, ok := product.(*MVE)
			if !ok {
				svc.Client.Logger.WarnContext(ctx, "Found MVE product type but couldn't cast to MVE struct")
				continue
			}

			// Filter inactive MVEs if requested
			if !req.IncludeInactive && (mve.ProvisioningStatus == STATUS_DECOMMISSIONED || mve.ProvisioningStatus == STATUS_CANCELLED) {
				continue
			}

			mves = append(mves, mve)
		}
	}

	return mves, nil
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
	mveResp := mveResponse{}
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
		ProductID:  req.MVEID,
		DeleteNow:  true,
		SafeDelete: req.SafeDelete,
	})
	if err != nil {
		return nil, err
	}
	return &DeleteMVEResponse{IsDeleted: true}, nil
}

// ListMVEImages returns a list of currently supported MVE images and details for each image, including image ID, version, product, and vendor. The image id returned indicates the software version and key configuration parameters of the image. The releaseImage value returned indicates whether the MVE image is available for selection when ordering an MVE.
// This method uses the v4 API and flattens the nested response to maintain backward compatibility.
func (svc *MVEServiceOp) ListMVEImages(ctx context.Context) ([]*MVEImage, error) {
	path := "/v4/product/mve/images"
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
	imageResp := MVEImageAPIResponseV4{}
	if err := json.Unmarshal(body, &imageResp); err != nil {
		return nil, err
	}

	// Flatten the nested v4 response structure to maintain backward compatibility
	// Product and Vendor are denormalized from the parent level to each image
	var flatImages []*MVEImage
	for _, productGroup := range imageResp.Data.Images {
		for _, img := range productGroup.Images {
			flatImages = append(flatImages, &MVEImage{
				ID:                img.ID,
				Version:           img.Version,
				Product:           productGroup.Product, // Denormalized from parent
				Vendor:            productGroup.Vendor,  // Denormalized from parent
				VendorDescription: img.VendorDescription,
				ReleaseImage:      img.ReleaseImage,
				ProductCode:       img.ProductCode,
				AvailableSizes:    img.AvailableSizes, // New field from v4 API
			})
		}
	}
	return flatImages, nil
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
	sizeResp := mveSizeAPIResponse{}
	if err := json.Unmarshal(body, &sizeResp); err != nil {
		return nil, err
	}
	return sizeResp.Data, nil
}

// validateBuyMVERequest validates a BuyMVERequest for proper term length.
func validateBuyMVERequest(req *BuyMVERequest) error {
	if !slices.Contains(VALID_CONTRACT_TERMS, req.Term) {
		return ErrInvalidTerm
	}
	return nil
}

func (svc *MVEServiceOp) ListMVEResourceTags(ctx context.Context, mveID string) (map[string]string, error) {
	tags, err := svc.Client.ProductService.ListProductResourceTags(ctx, mveID)
	if err != nil {
		return nil, err
	}
	return fromProductResourceTags(tags), nil
}

// UpdateMVEResourceTags updates the resource tags for an MVE in the Megaport MVE API.
func (svc *MVEServiceOp) UpdateMVEResourceTags(ctx context.Context, mveID string, tags map[string]string) error {
	return svc.Client.ProductService.UpdateProductResourceTags(ctx, mveID, &UpdateProductResourceTagsRequest{
		ResourceTags: toProductResourceTags(tags),
	})
}
