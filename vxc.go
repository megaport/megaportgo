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

// VXCService is an interface for interfacing with the VXC endpoints in the Megaport VXC API.
type VXCService interface {
	// BuyVXC buys a VXC from the Megaport VXC API.
	BuyVXC(ctx context.Context, req *BuyVXCRequest) (*BuyVXCResponse, error)
	// GetVXC gets details about a single VXC from the Megaport VXC API.
	GetVXC(ctx context.Context, id string) (*VXC, error)
	// DeleteVXC deletes a VXC in the Megaport VXC API.
	DeleteVXC(ctx context.Context, id string, req *DeleteVXCRequest) error
	// UpdateVXC updates a VXC in the Megaport VXC API.
	UpdateVXC(ctx context.Context, id string, req *UpdateVXCRequest) (*VXC, error)
	// LookupPartnerPorts looks up available partner ports in the Megaport VXC API.
	LookupPartnerPorts(ctx context.Context, req *LookupPartnerPortsRequest) (*LookupPartnerPortsResponse, error)
}

// NewVXCService creates a new instance of the VXC Service.
func NewVXCService(c *Client) *VXCServiceOp {
	return &VXCServiceOp{
		Client: c,
	}
}

var _ VXCService = &VXCServiceOp{}

// VXCServiceOp handles communication with the VXC related methods of the Megaport API.
type VXCServiceOp struct {
	Client *Client
}

// BuyVXCRequest represents a request to buy a VXC from the Megaport VXC API.
type BuyVXCRequest struct {
	PortUID           string
	VXCName           string
	RateLimit         int
	Term              int
	Shutdown          bool
	PromoCode         string
	ServiceKey        string
	CostCentre        string
	AEndConfiguration VXCOrderEndpointConfiguration
	BEndConfiguration VXCOrderEndpointConfiguration

	WaitForProvision bool          // Wait until the VXC provisions before returning
	WaitForTime      time.Duration // How long to wait for the VXC to provision if WaitForProvision is true (default is 5 minutes)
}

// BuyVXCResponse represents a response from buying a VXC from the Megaport VXC API.
type BuyVXCResponse struct {
	TechnicalServiceUID string
}

// DeleteVXCRequest represents a request to delete a VXC in the Megaport VXC API.
type DeleteVXCRequest struct {
	DeleteNow bool
}

// DeleteVXCResponse represents a response from deleting a VXC in the Megaport VXC API.
type DeleteVXCResponse struct {
	IsDeleting bool
}

// UpdateVXCRequest represents a request to update a VXC in the Megaport VXC API.
type UpdateVXCRequest struct {
	AEndVLAN       *int
	BEndVLAN       *int
	AEndProductUID *string
	BEndProductUID *string
	RateLimit      *int
	Name           *string
	CostCentre     *string
	Term           *int
	Shutdown       *bool

	WaitForUpdate bool          // Wait until the VXC updates before returning
	WaitForTime   time.Duration // How long to wait for the VXC to update if WaitForUpdate is true (default is 5 minutes)
}

// UpdateVXCResponse represents a response from updating a VXC in the Megaport VXC API.
type UpdateVXCResponse struct {
}

// LookupPartnerPortsRequest represents a request to lookup available partner ports in the Megaport VXC API.
type LookupPartnerPortsRequest struct {
	Key       string
	PortSpeed int
	Partner   string
	ProductID string
}

// LookupPartnerPortsResponse represents a response from looking up available partner ports in the Megaport VXC API.
type LookupPartnerPortsResponse struct {
	ProductUID string
}

// BuyVXC buys a VXC from the Megaport VXC API.
func (svc *VXCServiceOp) BuyVXC(ctx context.Context, req *BuyVXCRequest) (*BuyVXCResponse, error) {
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return nil, ErrInvalidTerm
	}

	buyOrder := []VXCOrder{{
		PortID: req.PortUID,
		AssociatedVXCs: []VXCOrderConfiguration{
			{
				Name:       req.VXCName,
				RateLimit:  req.RateLimit,
				Term:       req.Term,
				Shutdown:   req.Shutdown,
				PromoCode:  req.PromoCode,
				ServiceKey: req.ServiceKey,
				CostCentre: req.CostCentre,
				AEnd:       req.AEndConfiguration,
				BEnd:       req.BEndConfiguration,
			},
		},
	}}

	responseBody, responseError := svc.Client.ProductService.ExecuteOrder(ctx, buyOrder)
	if responseError != nil {
		return nil, responseError
	}

	orderInfo := VXCOrderResponse{}
	if err := json.Unmarshal(*responseBody, &orderInfo); err != nil {
		return nil, err
	}
	serviceUID := orderInfo.Data[0].TechnicalServiceUID

	// wait until the VXC is provisioned before returning if reqested by the user
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
				return nil, fmt.Errorf("time expired waiting for VXC %s to provision", serviceUID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for VXC %s to provision", serviceUID)
			case <-ticker.C:
				vxcDetails, err := svc.GetVXC(ctx, serviceUID)
				if err != nil {
					return nil, err
				}

				if slices.Contains(SERVICE_STATE_READY, vxcDetails.ProvisioningStatus) {
					return &BuyVXCResponse{
						TechnicalServiceUID: serviceUID,
					}, nil
				}

			}
		}
	} else {
		// return the service UID right away if the user doesn't want to wait for provision
		return &BuyVXCResponse{
			TechnicalServiceUID: serviceUID,
		}, nil
	}
}

// GetVXC gets details about a single VXC from the Megaport VXC API.
func (svc *VXCServiceOp) GetVXC(ctx context.Context, id string) (*VXC, error) {
	path := "/v2/product/" + id
	url := svc.Client.BaseURL.JoinPath(path).String()

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	vxcDetails := VXCResponse{}
	if err = json.Unmarshal(body, &vxcDetails); err != nil {
		return nil, err
	}

	return &vxcDetails.Data, nil
}

// DeleteVXC deletes a VXC in the Megaport VXC API.
func (svc *VXCServiceOp) DeleteVXC(ctx context.Context, id string, req *DeleteVXCRequest) error {
	_, err := svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID: id,
		DeleteNow: req.DeleteNow,
	})
	if err != nil {
		return err
	}
	return nil
}

// UpdateVXC updates a VXC in the Megaport VXC API.
func (svc *VXCServiceOp) UpdateVXC(ctx context.Context, id string, req *UpdateVXCRequest) (*VXC, error) {
	if req.Term != nil && (*req.Term != 1 && *req.Term != 12 && *req.Term != 24 && *req.Term != 36) {
		return nil, ErrInvalidTerm
	}
	if req.CostCentre != nil && len(*req.CostCentre) > 255 {
		return nil, ErrCostCentreTooLong
	}

	path := fmt.Sprintf("/v3/product/%s/%s", PRODUCT_VXC, id)
	url := svc.Client.BaseURL.JoinPath(path).String()

	update := &VXCUpdate{
		RateLimit: req.RateLimit,
		AEndVLAN:  req.AEndVLAN,
		BEndVLAN:  req.BEndVLAN,
		Term:      req.Term,
		Shutdown:  req.Shutdown,
	}

	if req.Name != nil {
		update.Name = *req.Name
	}
	if req.AEndProductUID != nil {
		update.AEndProductUID = *req.AEndProductUID
	}
	if req.BEndProductUID != nil {
		update.BEndProductUID = *req.BEndProductUID
	}
	if req.CostCentre != nil {
		update.CostCentre = *req.CostCentre
	}

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPut, url, update)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	vxcDetails := VXCResponse{}
	if err = json.Unmarshal(body, &vxcDetails); err != nil {
		return nil, err
	}

	// wait until the VXC is updated before returning if requested by the user
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
				return nil, fmt.Errorf("time expired waiting for VXC %s to update", id)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for VXC %s to update", id)
			case <-ticker.C:
				vxc, err := svc.GetVXC(ctx, id)
				if err != nil {
					return nil, err
				}

				var isUpdated bool
				if vxc.ProvisioningStatus == "LIVE" {
					isUpdated = true
				}
				if isUpdated {
					return vxc, nil
				}
			}
		}
	} else {
		// return the response right away if the user doesn't want to wait for update
		return &vxcDetails.Data, nil
	}
}

// LookupPartnerPorts looks up available partner ports in the Megaport VXC API.
func (svc *VXCServiceOp) LookupPartnerPorts(ctx context.Context, req *LookupPartnerPortsRequest) (*LookupPartnerPortsResponse, error) {
	lookupUrl := "/v2/secure/" + strings.ToLower(req.Partner) + "/" + req.Key
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, lookupUrl, nil)
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

	lookupResponse := PartnerLookupResponse{}
	parseErr := json.Unmarshal(body, &lookupResponse)

	if parseErr != nil {
		return nil, parseErr
	}

	toReturn := &LookupPartnerPortsResponse{}

	for i := 0; i < len(lookupResponse.Data.Megaports); i++ {
		if lookupResponse.Data.Megaports[i].VXC == 0 && lookupResponse.Data.Megaports[i].PortSpeed >= req.PortSpeed { // nil is 0
			// We only need the first available one that has enough speed capacity.
			if req.ProductID == "" {
				toReturn.ProductUID = lookupResponse.Data.Megaports[i].ProductUID
				return toReturn, nil
				// Try to match Product ID if provided
			} else if lookupResponse.Data.Megaports[i].ProductUID == req.ProductID {
				toReturn.ProductUID = lookupResponse.Data.Megaports[i].ProductUID
				return toReturn, nil
			}
		}
	}
	return nil, ErrNoAvailableVxcPorts
}
