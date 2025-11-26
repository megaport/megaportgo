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
	// ValidateVXCOrder validates a VXC order in the Megaport Products API.
	ValidateVXCOrder(ctx context.Context, req *BuyVXCRequest) error
	// ListVXCs lists all VXCs in the Megaport VXC API.
	ListVXCs(ctx context.Context, req *ListVXCsRequest) ([]*VXC, error)
	// GetVXC gets details about a single VXC from the Megaport VXC API.
	GetVXC(ctx context.Context, id string) (*VXC, error)
	// DeleteVXC deletes a VXC in the Megaport VXC API.
	DeleteVXC(ctx context.Context, id string, req *DeleteVXCRequest) error
	// UpdateVXC updates a VXC in the Megaport VXC API.
	UpdateVXC(ctx context.Context, id string, req *UpdateVXCRequest) (*VXC, error)
	// LookupPartnerPorts looks up available partner ports in the Megaport VXC API.
	LookupPartnerPorts(ctx context.Context, req *LookupPartnerPortsRequest) (*LookupPartnerPortsResponse, error)
	// ListPartnerPorts lists available partner ports in the Megaport VXC API.
	ListPartnerPorts(ctx context.Context, req *ListPartnerPortsRequest) (*ListPartnerPortsResponse, error)
	// ListVXCResourceTags lists the resource tags for a VXC in the Megaport Products API.
	ListVXCResourceTags(ctx context.Context, vxcID string) (map[string]string, error)
	// UpdateVXCResourceTags updates the resource tags for a VXC in the Megaport Products API.
	UpdateVXCResourceTags(ctx context.Context, vxcID string, tags map[string]string) error
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

	ResourceTags map[string]string `json:"resourceTags,omitempty"`
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
	AEndVLAN       *int    // A unique VLAN ID for this connection. Values can range from 2 to 4093. If this value is 0, the system allocates a valid VLAN. If the value is -1, the system untags the VLAN and sets it to null.
	BEndVLAN       *int    // A unique VLAN ID for this connection. Values can range from 2 to 4093. If this value is 0, the system allocates a valid VLAN. If the value is -1, the system untags the VLAN and sets it to null.
	AEndProductUID *string // When moving a VXC, this is the new A-End for the connection.
	BEndProductUID *string // When moving a VXC, this is the new B-End for the connection.
	RateLimit      *int    // A new speed for the connection.
	Name           *string // Customer name for the connection - this name appears in the Portal.
	CostCentre     *string // A customer reference number to be included in billing information and invoices. Also known as the Service Level Reference (SLR).
	Term           *int
	Shutdown       *bool // Temporarily shut down and re-enable the VXC. Valid values are true (shut down) and false (enabled). If not provided, it defaults to false (enabled).

	AEndInnerVLAN *int
	BEndInnerVLAN *int

	AVnicIndex *int // When moving a VXC for an MVE, this is the new A-End vNIC for the connection.
	BVnicIndex *int // When moving a VXC for an MVE, this is the new B-End vNIC for the connection.

	IsApproved *bool //  Define whether the VXC is approved or rejected via the Megaport Marketplace. Set to true (Approved) or false (Rejected).

	AEndPartnerConfig VXCPartnerConfiguration
	BEndPartnerConfig VXCPartnerConfiguration

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

// ListPartnerPortsRequest represents a request to list available partner ports in the Megaport VXC API.
type ListPartnerPortsRequest struct {
	Key     string
	Partner string
}

type ListPartnerPortsResponse struct {
	Data PartnerLookup
}

// ListVXCsRequest represents a request to list VXCs in the Megaport VXC API.
type ListVXCsRequest struct {
	// Basic filters
	Name         string // Filter by name (exact match)
	NameContains string // Filter by partial name match

	// Status filters
	Status []string // Filter by specific provisioning statuses (e.g. "LIVE", "CONFIGURED")

	// Connection filters
	AEndProductUID string // Filter by A-End product UID
	BEndProductUID string // Filter by B-End product UID

	// Other common filters
	RateLimit       int  // Filter by specific rate limit (in Mbps)
	IncludeInactive bool // Include inactive VXCs in the results
}

// BuyVXC buys a VXC from the Megaport VXC API.
func (svc *VXCServiceOp) BuyVXC(ctx context.Context, req *BuyVXCRequest) (*BuyVXCResponse, error) {
	if !slices.Contains(VALID_CONTRACT_TERMS, req.Term) {
		return nil, ErrInvalidTerm
	}

	buyOrder := createVXCOrder(req)

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

// ListVXCResourceTags lists the resource tags for a VXC in the Megaport Products API.
func (svc *VXCServiceOp) ListVXCResourceTags(ctx context.Context, vxcID string) (map[string]string, error) {
	tags, err := svc.Client.ProductService.ListProductResourceTags(ctx, vxcID)
	if err != nil {
		return nil, err
	}
	return fromProductResourceTags(tags), nil
}

// UpdateVXCResourceTags updates the resource tags for a VXC in the Megaport Products API.
func (svc *VXCServiceOp) UpdateVXCResourceTags(ctx context.Context, vxcID string, tags map[string]string) error {
	return svc.Client.ProductService.UpdateProductResourceTags(ctx, vxcID, &UpdateProductResourceTagsRequest{
		ResourceTags: toProductResourceTags(tags),
	})
}

func createVXCOrder(req *BuyVXCRequest) []VXCOrder {
	return []VXCOrder{{
		PortID: req.PortUID,
		AssociatedVXCs: []VXCOrderConfiguration{
			{
				Name:         req.VXCName,
				RateLimit:    req.RateLimit,
				Term:         req.Term,
				Shutdown:     req.Shutdown,
				PromoCode:    req.PromoCode,
				ServiceKey:   req.ServiceKey,
				CostCentre:   req.CostCentre,
				AEnd:         req.AEndConfiguration,
				BEnd:         req.BEndConfiguration,
				ResourceTags: toProductResourceTags(req.ResourceTags),
			},
		},
	}}
}

// ValidateVXCOrder validates a VXC order in the Megaport VXC API.
func (svc *VXCServiceOp) ValidateVXCOrder(ctx context.Context, req *BuyVXCRequest) error {
	buyOrder := createVXCOrder(req)

	return svc.Client.ProductService.ValidateProductOrder(ctx, buyOrder)
}

// isTransitVXC checks if a VXC is a Transit VXC (Megaport Internet) by examining the partner configuration.
// A VXC is considered a Transit VXC if either A-End or B-End has connectType "TRANSIT".
func isTransitVXC(vxc *VXC) bool {
	if vxc == nil || vxc.Resources == nil || vxc.Resources.CSPConnection == nil {
		return false
	}

	for _, csp := range vxc.Resources.CSPConnection.CSPConnection {
		if transitCSP, ok := csp.(CSPConnectionTransit); ok && transitCSP.ConnectType == "TRANSIT" {
			return true
		}
	}
	return false
}

// DeleteVXC deletes a VXC in the Megaport VXC API.
// Note: Transit VXCs (Megaport Internet) only support immediate deletion (CANCEL_NOW).
// If the VXC is a Transit VXC, the DeleteNow flag will be automatically enforced.
func (svc *VXCServiceOp) DeleteVXC(ctx context.Context, id string, req *DeleteVXCRequest) error {
	// Check if this is a Transit VXC that requires immediate deletion
	vxc, err := svc.GetVXC(ctx, id)
	if err != nil {
		return err
	}

	// Enforce Transit VXC lifecycle restriction: only CANCEL_NOW is allowed
	if isTransitVXC(vxc) && !req.DeleteNow {
		return ErrTransitVXCCancelLaterNotAllowed
	}

	_, err = svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
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
	if req.Term != nil && !slices.Contains(VALID_CONTRACT_TERMS, *req.Term) {
		return nil, ErrInvalidTerm
	}
	if req.CostCentre != nil && len(*req.CostCentre) > 255 {
		return nil, ErrCostCentreTooLong
	}

	path := fmt.Sprintf("/v3/product/%s/%s", PRODUCT_VXC, id)
	url := svc.Client.BaseURL.JoinPath(path).String()

	update := &VXCUpdate{
		RateLimit:  req.RateLimit,
		AEndVLAN:   req.AEndVLAN,
		BEndVLAN:   req.BEndVLAN,
		Term:       req.Term,
		Shutdown:   req.Shutdown,
		IsApproved: req.IsApproved,
		AVnicIndex: req.AVnicIndex,
		BVnicIndex: req.BVnicIndex,
	}

	if req.AEndPartnerConfig != nil {
		// Only allow AENdPartnerConfig or VROUTER Partner Config for AEndPartnerConfig in VXC Updates
		switch req.AEndPartnerConfig.(type) {
		case VXCPartnerConfiguration, *VXCOrderVrouterPartnerConfig, VXCOrderVrouterPartnerConfig:
			update.AEndPartnerConfig = req.AEndPartnerConfig
		default:
			return nil, ErrInvalidVXCAEndPartnerConfig
		}
	}

	if req.BEndPartnerConfig != nil {
		// Only allow Vrouter Partner Config for BEndPartnerConfig in VXC Updates
		switch req.BEndPartnerConfig.(type) {
		case *VXCOrderVrouterPartnerConfig, VXCOrderVrouterPartnerConfig:
			update.BEndPartnerConfig = req.BEndPartnerConfig
		default:
			return nil, ErrInvalidVXCBEndPartnerConfig
		}
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
		update.CostCentre = req.CostCentre
	}
	if req.AEndInnerVLAN != nil {
		update.AEndInnerVLAN = req.AEndInnerVLAN
	}
	if req.BEndInnerVLAN != nil {
		update.BEndInnerVLAN = req.BEndInnerVLAN
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

// LookupPartnerPorts looks up available partner ports in the Megaport VXC API.
func (svc *VXCServiceOp) ListPartnerPorts(ctx context.Context, req *ListPartnerPortsRequest) (*ListPartnerPortsResponse, error) {
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

	return &ListPartnerPortsResponse{
		Data: lookupResponse.Data,
	}, nil
}

// ListVXCs lists all VXCs in the Megaport VXC API.
func (svc *VXCServiceOp) ListVXCs(ctx context.Context, req *ListVXCsRequest) ([]*VXC, error) {
	// Create a map to track unique VXCs by their UID
	uniqueVXCs := make(map[string]*VXC)

	// Get all products with a single API call
	allProducts, err := svc.Client.ProductService.ListProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	// Process each product to extract associated VXCs
	for _, product := range allProducts {
		// Check the associated VXCs
		for _, vxc := range product.GetAssociatedVXCs() {
			// If the VXC is already in the map, skip it
			if _, exists := uniqueVXCs[vxc.UID]; exists {
				continue
			}

			// Add the VXC to the map
			uniqueVXCs[vxc.UID] = vxc
		}
	}

	// Create a filtered slice of VXCs
	vxcs := make([]*VXC, 0, len(uniqueVXCs))
	for _, vxc := range uniqueVXCs {
		// Apply filters
		if shouldIncludeVXC(vxc, req) {
			vxcs = append(vxcs, vxc)
		}
	}

	return vxcs, nil
}

// Helper function to determine if a VXC matches the filter criteria
func shouldIncludeVXC(vxc *VXC, req *ListVXCsRequest) bool {
	if req == nil {
		return true
	}

	// Name filter
	if req.Name != "" && vxc.Name != req.Name {
		return false
	}

	// Name contains filter
	if req.NameContains != "" && !strings.Contains(vxc.Name, req.NameContains) {
		return false
	}

	// Status filter
	if len(req.Status) > 0 && !slices.Contains(req.Status, vxc.ProvisioningStatus) {
		return false
	}

	// A-End filter
	if req.AEndProductUID != "" && vxc.AEndConfiguration.UID != req.AEndProductUID {
		return false
	}

	// B-End filter
	if req.BEndProductUID != "" && vxc.BEndConfiguration.UID != req.BEndProductUID {
		return false
	}

	// Rate limit filter
	if req.RateLimit > 0 && vxc.RateLimit != req.RateLimit {
		return false
	}

	// Skip inactive VXCs if IncludeInactive is false
	if !req.IncludeInactive &&
		(vxc.ProvisioningStatus == STATUS_DECOMMISSIONED ||
			vxc.ProvisioningStatus == STATUS_CANCELLED) {
		return false
	}

	return true
}
