package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

// MVE is an interface for interfacing with the MVE endpoints
// of the Megaport API.
type MVEService interface {
	BuyMVE(ctx context.Context, req *BuyMVERequest) (*BuyMVEResponse, error)
	GetMVE(ctx context.Context, mveId string) (*MVE, error)
	ModifyMVE(ctx context.Context, req *ModifyMVERequest) (*ModifyMVEResponse, error)
	DeleteMVE(ctx context.Context, req *DeleteMVERequest) (*DeleteMVEResponse, error)
	WaitForMVEProvisioning(ctx context.Context, mveId string) (bool, error)
}

func NewMVEServiceOp(c *Client) *MVEServiceOp {
	return &MVEServiceOp{
		Client: c,
	}
}

// MVEServiceOp handles communication with MVE methods of the Megaport API.
type MVEServiceOp struct {
	Client *Client
}

type BuyMVERequest struct {
	LocationID   int
	Name         string
	Term         int
	VendorConfig vendorConfig
	Vnics        []MVENetworkInterface
}

type BuyMVEResponse struct {
	TechnicalServiceUID string
}

type ModifyMVERequest struct {
	MVEID string
	Name  string
}

type ModifyMVEResponse struct {
	MVEUpdated bool
}

type DeleteMVERequest struct {
	MVEID string
}

type DeleteMVEResponse struct {
	IsDeleted bool
}

func (svc *MVEServiceOp) BuyMVE(ctx context.Context, req *BuyMVERequest) (*BuyMVEResponse, error) {
	err := validateBuyMVERequest(req)
	if err != nil {
		return nil, err
	}
	order := &MVEOrderConfig{
		LocationID:   req.LocationID,
		Name:         req.Name,
		Term:         req.Term,
		ProductType:  strings.ToUpper(PRODUCT_MVE),
	}
	switch req.VendorConfig.(type) {
	case *ArubaConfig:
		c := req.VendorConfig.(ArubaConfig)
		order.VendorConfig = &ArubaConfig{
			Vendor: c.Vendor,
			ImageID: c.ImageID,
			ProductSize: c.ProductSize,
			AccountName: c.AccountName,
			AccountKey: c.AccountKey,
		}
	case *CiscoConfig:
		c := req.VendorConfig.(CiscoConfig)
		order.VendorConfig = &CiscoConfig{
			Vendor: c.Vendor,
			ImageID: c.ImageID,
			ProductSize: c.ProductSize,
			AdminSSHPublicKey: c.AdminSSHPublicKey,
			CloudInit: c.CloudInit,
		}
	case *FortinetConfig:
		c := req.VendorConfig.(FortinetConfig)
		order.VendorConfig = &FortinetConfig{
			Vendor: c.Vendor,
			ImageID: c.ImageID,
			ProductSize: c.ProductSize,
			AdminSSHPublicKey: c.AdminSSHPublicKey,
			LicenseData: c.LicenseData,
		}
	case *PaloAltoConfig:
		c := req.VendorConfig.(PaloAltoConfig)
		order.VendorConfig = &PaloAltoConfig{
			Vendor: c.Vendor,
			ImageID: c.ImageID,
			ProductSize: c.ProductSize,
			AdminSSHPublicKey: c.AdminSSHPublicKey,
			AdminPasswordHash: c.AdminPasswordHash,
			LicenseData: c.LicenseData,
		}
	case *VersaConfig:
		c := req.VendorConfig.(VersaConfig)
		order.VendorConfig = &VersaConfig{
			Vendor: c.Vendor,
			ImageID: c.ImageID,
			ProductSize: c.ProductSize,
			DirectorAddress: c.DirectorAddress,
			ControllerAddress: c.ControllerAddress,
			LocalAuth: c.LocalAuth,
			RemoteAuth: c.RemoteAuth,
			SerialNumber: c.SerialNumber,
			}
	case *VmwareConfig:
		c := req.VendorConfig.(VmwareConfig)
		order.VendorConfig = &VmwareConfig{
			Vendor: c.Vendor,
			ImageID: c.ImageID,
			ProductSize: c.ProductSize,
			AdminSSHPublicKey: c.AdminSSHPublicKey,
			VcoAddress: c.VcoAddress,
			VcoActivationCode: c.VcoActivationCode,
		}	
	}

	if len(req.Vnics) == 0 {
		order.NetworkInterfaces = []MVENetworkInterface{{Description: "Data Plane", VLAN: 0}}
	} else {
		order.NetworkInterfaces = req.Vnics
	}

	resp, err := svc.Client.ProductService.ExecuteOrder(ctx, []*MVEOrderConfig{order})
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


	return toReturn, nil
}

func (svc *MVEServiceOp) GetMVE(ctx context.Context, mveId string) (*MVE, error) {
	path := "/v2/product" + mveId
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

func (svc *MVEServiceOp) ModifyMVE(ctx context.Context, req *ModifyMVERequest) (*ModifyMVEResponse, error) {
	modifyProductReq := &ModifyProductRequest{
		ProductID:             req.MVEID,
		ProductType:           PRODUCT_MVE,
		Name:                  req.Name,
		CostCentre:            "",
		MarketplaceVisibility: false,
	}
	_, err := svc.Client.ProductService.ModifyProduct(ctx, modifyProductReq)
	if err != nil {
		return nil, err
	}
	return &ModifyMVEResponse{
		MVEUpdated: true,
	}, nil
}

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

func (svc *MVEServiceOp) WaitForMVEProvisioning(ctx context.Context, mveId string) (bool, error) {
	// Try for ~5mins.
	for i := 0; i < 30; i++ {
		mve, err := svc.GetMVE(ctx, mveId)
		if err != nil {
			return false, err
		}

		if slices.Contains(SERVICE_STATE_READY, mve.ProvisioningStatus) {
			return true, nil
		}

		// Wrong status, wait a bit and try again.
		svc.Client.Logger.DebugContext(ctx, "Waiting for MVE", slog.String("provisioning_status", mve.ProvisioningStatus))
		time.Sleep(10 * time.Second)
	}

	return false, errors.New(ERR_MVE_PROVISION_TIMEOUT_EXCEED)

}

func validateBuyMVERequest(req *BuyMVERequest) error {
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return errors.New(ERR_TERM_NOT_VALID)
	}
	return nil
}
