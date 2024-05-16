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
	// GetMVE gets details about a single MVE from the Megaport MVE API.
	GetMVE(ctx context.Context, mveId string) (*MVE, error)
	// ModifyMVE modifies an MVE in the Megaport MVE API.
	ModifyMVE(ctx context.Context, req *ModifyMVERequest) (*ModifyMVEResponse, error)
	// DeleteMVE deletes an MVE in the Megaport MVE API.
	DeleteMVE(ctx context.Context, req *DeleteMVERequest) (*DeleteMVEResponse, error)
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

	WaitForProvision bool          // Wait until the MVE provisions before returning
	WaitForTime      time.Duration // How long to wait for the MVE to provision if WaitForProvision is true (default is 5 minutes)
}

// BuyMVEResponse represents a response from buying an MVE
type BuyMVEResponse struct {
	TechnicalServiceUID string
}

// ModifyMVERequest represents a request to modify an MVE
type ModifyMVERequest struct {
	MVEID string
	Name  string

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
	order := &MVEOrderConfig{
		LocationID:    req.LocationID,
		Name:          req.Name,
		Term:          req.Term,
		ProductType:   strings.ToUpper(PRODUCT_MVE),
		DiversityZone: req.DiversityZone,
	}
	switch c := req.VendorConfig.(type) {
	case *ArubaConfig:
		order.VendorConfig = &ArubaConfig{
			Vendor:      c.Vendor,
			ImageID:     c.ImageID,
			ProductSize: c.ProductSize,
			MVELabel:    c.MVELabel,
			AccountName: c.AccountName,
			AccountKey:  c.AccountKey,
		}
	case *CiscoConfig:
		order.VendorConfig = &CiscoConfig{
			Vendor:             c.Vendor,
			ImageID:            c.ImageID,
			ProductSize:        c.ProductSize,
			MVELabel:           c.MVELabel,
			AdminSSHPublicKey:  c.AdminSSHPublicKey,
			SSHPublicKey:       c.SSHPublicKey,
			ManageLocally:      c.ManageLocally,
			CloudInit:          c.CloudInit,
			FMCIPAddress:       c.FMCIPAddress,
			FMCRegistrationKey: c.FMCRegistrationKey,
			FMCNatID:           c.FMCNatID,
		}
	case *FortinetConfig:
		order.VendorConfig = &FortinetConfig{
			Vendor:            c.Vendor,
			ImageID:           c.ImageID,
			ProductSize:       c.ProductSize,
			MVELabel:          c.MVELabel,
			AdminSSHPublicKey: c.AdminSSHPublicKey,
			SSHPublicKey:      c.SSHPublicKey,
			LicenseData:       c.LicenseData,
		}
	case *PaloAltoConfig:
		order.VendorConfig = &PaloAltoConfig{
			Vendor:            c.Vendor,
			ImageID:           c.ImageID,
			ProductSize:       c.ProductSize,
			MVELabel:          c.MVELabel,
			AdminSSHPublicKey: c.AdminSSHPublicKey,
			AdminPasswordHash: c.AdminPasswordHash,
			LicenseData:       c.LicenseData,
		}
	case *VersaConfig:
		order.VendorConfig = &VersaConfig{
			Vendor:            c.Vendor,
			ImageID:           c.ImageID,
			ProductSize:       c.ProductSize,
			MVELabel:          c.MVELabel,
			DirectorAddress:   c.DirectorAddress,
			ControllerAddress: c.ControllerAddress,
			LocalAuth:         c.LocalAuth,
			RemoteAuth:        c.RemoteAuth,
			SerialNumber:      c.SerialNumber,
		}
	case *VmwareConfig:
		order.VendorConfig = &VmwareConfig{
			Vendor:            c.Vendor,
			ImageID:           c.ImageID,
			ProductSize:       c.ProductSize,
			MVELabel:          c.MVELabel,
			AdminSSHPublicKey: c.AdminSSHPublicKey,
			SSHPublicKey:      c.SSHPublicKey,
			VcoAddress:        c.VcoAddress,
			VcoActivationCode: c.VcoActivationCode,
		}
	case *MerakiConfig:
		order.VendorConfig = &MerakiConfig{
			Vendor:      c.Vendor,
			ImageID:     c.ImageID,
			ProductSize: c.ProductSize,
			MVELabel:    c.MVELabel,
			Token:       c.Token,
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
		Name:                  req.Name,
		CostCentre:            "",
		MarketplaceVisibility: PtrTo(false),
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
				if mveDetails.ProvisioningStatus == "LIVE" {
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

// validateBuyMVERequest validates a BuyMVERequest for proper term length.
func validateBuyMVERequest(req *BuyMVERequest) error {
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return ErrInvalidTerm
	}
	return nil
}
