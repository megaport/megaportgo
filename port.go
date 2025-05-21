package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

// PortService is an interface for interfacing with the Port endpoints of the Megaport API.
type PortService interface {
	// BuyPort buys a port from the Megaport Port API.
	BuyPort(ctx context.Context, req *BuyPortRequest) (*BuyPortResponse, error)
	// ValidatePortOrder validates a port order in the Megaport Products API.
	ValidatePortOrder(ctx context.Context, req *BuyPortRequest) error
	// ListPorts lists all ports in the Megaport Port API.
	ListPorts(ctx context.Context) ([]*Port, error)
	// GetPort gets a single port in the Megaport Port API.
	GetPort(ctx context.Context, portId string) (*Port, error)
	// ModifyPort modifies a port in the Megaport Port API.
	ModifyPort(ctx context.Context, req *ModifyPortRequest) (*ModifyPortResponse, error)
	// DeletePort deletes a port in the Megaport Port API.
	DeletePort(ctx context.Context, req *DeletePortRequest) (*DeletePortResponse, error)
	// RestorePort restores a port in the Megaport Port API.
	RestorePort(ctx context.Context, portId string) (*RestorePortResponse, error)
	// LockPort locks a port in the Megaport Port API.
	LockPort(ctx context.Context, portId string) (*LockPortResponse, error)
	// UnlockPort unlocks a port in the Megaport Port API.
	UnlockPort(ctx context.Context, portId string) (*UnlockPortResponse, error)
	// CheckPortVLANAvailability checks if a VLAN is available on a port in the Megaport Products API.
	CheckPortVLANAvailability(ctx context.Context, portId string, vlan int) (bool, error)
	// ListPortResourceTags lists the resource tags for a port in the Megaport Port API.
	ListPortResourceTags(ctx context.Context, portID string) (map[string]string, error)
	// UpdatePortResourceTags updates the resource tags for a port in the Megaport Port API.
	UpdatePortResourceTags(ctx context.Context, portID string, tags map[string]string) error
	// GetPortLOA retrieves the LOA (Letter of Authorization) for a specific port
	GetPortLOA(ctx context.Context, portID string) ([]byte, error)
	// SavePortLOAToFile retrieves the LOA for a port and saves it to the specified file path
	SavePortLOAToFile(ctx context.Context, portID string, filePath string) error
	// GetPortLOABase64 retrieves the LOA for a port and returns it as a base64-encoded string
	GetPortLOABase64(ctx context.Context, portID string) (string, error)
}

// NewPortService creates a new instance of the Port Service.
func NewPortService(c *Client) *PortServiceOp {
	return &PortServiceOp{
		Client: c,
	}
}

// PortServiceOp handles communication with Port methods of the Megaport API.
type PortServiceOp struct {
	Client *Client
}

// BuyPortRequest represents a request to buy a port.
type BuyPortRequest struct {
	Name                  string `json:"name"`
	Term                  int    `json:"term"`
	PortSpeed             int    `json:"portSpeed"`
	LocationId            int    `json:"locationId"`
	Market                string `json:"market"`
	LagCount              int    `json:"lagCount"` // A lag count of 1 or higher will order the port as a single LAG
	MarketPlaceVisibility bool   `json:"marketPlaceVisibility"`
	DiversityZone         string `json:"diversityZone"`
	CostCentre            string `json:"costCentre"`
	PromoCode             string `json:"promoCode"`

	ResourceTags map[string]string `json:"resourceTags"`

	WaitForProvision bool          // Wait until the VXC provisions before returning
	WaitForTime      time.Duration // How long to wait for the VXC to provision if WaitForProvision is true (default is 5 minutes)
}

// BuyPortResponse represents a response from buying a port.
type BuyPortResponse struct {
	TechnicalServiceUIDs []string
}

// GetPortRequest represents a request to get a port.
type GetPortRequest struct {
	PortID string
}

// ModifyPortRequest represents a request to modify a port.
type ModifyPortRequest struct {
	PortID                string
	Name                  string
	MarketplaceVisibility *bool
	CostCentre            string
	ContractTermMonths    *int

	WaitForUpdate bool          // Wait until the Port updates before returning
	WaitForTime   time.Duration // How long to wait for the Port to update if WaitForUpdate is true (default is 5 minutes)
}

// ModifyPortResponse represents a response from modifying a port.
type ModifyPortResponse struct {
	IsUpdated bool
}

// DeletePortRequest represents a request to delete a port.
type DeletePortRequest struct {
	PortID     string
	DeleteNow  bool
	SafeDelete bool // If true, the API will return an error if the port is in use by a VXC or other product
}

// DeletePortResponse represents a response from deleting a port.
type DeletePortResponse struct {
	IsDeleting bool
}

// RestorePortRequest represents a request to restore a port.
type RestorePortRequest struct {
	PortID string
}

// RestorePortResponse represents a response from restoring a port.
type RestorePortResponse struct {
	IsRestored bool
}

// LockPortRequest represents a request to lock a port.
type LockPortRequest struct {
	PortID string
}

// LockPortResponse represents a response from locking a port.
type LockPortResponse struct {
	IsLocking bool
}

// UnlockPortRequest represents a request to unlock a port.
type UnlockPortRequest struct {
	PortID string
}

// UnlockPortResponse represents a response from unlocking a port.
type UnlockPortResponse struct {
	IsUnlocking bool
}

// BuyPort buys a port from the Megaport Port API.
func (svc *PortServiceOp) BuyPort(ctx context.Context, req *BuyPortRequest) (*BuyPortResponse, error) {
	if req.Term != 1 && req.Term != 12 && req.Term != 24 && req.Term != 36 {
		return nil, ErrInvalidTerm
	}

	buyOrder := createPortOrder(req)

	responseBody, responseError := svc.Client.ProductService.ExecuteOrder(ctx, buyOrder)
	if responseError != nil {
		return nil, responseError
	}
	orderInfo := PortOrderResponse{}
	unmarshalErr := json.Unmarshal(*responseBody, &orderInfo)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	toReturn := &BuyPortResponse{
		TechnicalServiceUIDs: []string{},
	}
	for _, d := range orderInfo.Data {
		toReturn.TechnicalServiceUIDs = append(toReturn.TechnicalServiceUIDs, d.TechnicalServiceUID)
	}

	// wait until the Port is provisioned before returning if requested by the user
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
				return nil, fmt.Errorf("time expired waiting for Port %s to provision", toReturn.TechnicalServiceUIDs)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for Port %s to provision", toReturn.TechnicalServiceUIDs)
			case <-ticker.C:
				ports := []*Port{}
				for _, uid := range toReturn.TechnicalServiceUIDs {
					portDetails, err := svc.GetPort(ctx, uid)
					if err != nil {
						return nil, err
					}

					ports = append(ports, portDetails)
				}

				// if all ports are ready return
				numReady := 0
				for _, port := range ports {
					if slices.Contains(SERVICE_STATE_READY, port.ProvisioningStatus) {
						numReady++
					}
				}
				if numReady == len(ports) {
					return toReturn, nil
				}
			}
		}
	} else {
		// return the service UID right away if the user doesn't want to wait for provision
		return toReturn, nil
	}
}

func createPortOrder(req *BuyPortRequest) []PortOrder {
	return []PortOrder{{
		Name:        req.Name,
		Term:        req.Term,
		ProductType: "MEGAPORT",
		PortSpeed:   req.PortSpeed,
		LocationID:  req.LocationId,
		Config: PortOrderConfig{
			DiversityZone: req.DiversityZone,
		},
		Virtual:               false,
		Market:                req.Market,
		LagPortCount:          req.LagCount,
		MarketplaceVisibility: req.MarketPlaceVisibility,
		CostCentre:            req.CostCentre,
		PromoCode:             req.PromoCode,
		ResourceTags:          toProductResourceTags(req.ResourceTags),
	}}
}

func (svc *PortServiceOp) ValidatePortOrder(ctx context.Context, req *BuyPortRequest) error {
	if !slices.Contains(VALID_CONTRACT_TERMS, req.Term) {
		// Validate that term is one of the allowed values
		return ErrInvalidTerm
	}

	buyOrder := createPortOrder(req)

	return svc.Client.ProductService.ValidateProductOrder(ctx, buyOrder)
}

func (svc *PortServiceOp) ListPorts(ctx context.Context) ([]*Port, error) {
	allProducts, err := svc.Client.ProductService.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	ports := []*Port{}

	for _, product := range allProducts {
		if strings.ToLower(product.GetType()) == PRODUCT_MEGAPORT {
			port, ok := product.(*Port)
			if !ok {
				svc.Client.Logger.WarnContext(ctx, "Found Port product type but couldn't cast to Port struct")
				continue
			}

			ports = append(ports, port)
		}
	}

	return ports, nil
}

// GetPort gets a single port in the Megaport Port API.
func (svc *PortServiceOp) GetPort(ctx context.Context, portId string) (*Port, error) {
	path := "/v2/product/" + portId
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

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}

	portDetails := PortResponse{}
	unmarshalErr := json.Unmarshal(body, &portDetails)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	port := &portDetails.Data
	// If the port is part of a LAG, we need to fetch the LAG details
	if port.AggregationID != 0 {
		// Fetch all ports to find those with the same AggregationID
		allPorts, err := svc.ListPorts(ctx)
		if err != nil {
			return nil, err
		}

		lagPortUIDs := []string{}
		for _, p := range allPorts {
			if p.AggregationID == port.AggregationID {
				lagPortUIDs = append(lagPortUIDs, p.UID)
			}
		}

		port.LagPortUIDs = lagPortUIDs
		port.LagCount = len(lagPortUIDs)
	}

	return port, nil
}

// ModifyPort modifies a port in the Megaport Port API.
func (svc *PortServiceOp) ModifyPort(ctx context.Context, req *ModifyPortRequest) (*ModifyPortResponse, error) {
	if len(req.CostCentre) > 255 {
		return nil, ErrCostCentreTooLong
	}

	modifyReq := &ModifyProductRequest{
		ProductID:             req.PortID,
		ProductType:           PRODUCT_MEGAPORT,
		Name:                  req.Name,
		CostCentre:            req.CostCentre,
		MarketplaceVisibility: req.MarketplaceVisibility,
	}
	if req.ContractTermMonths != nil {
		modifyReq.ContractTermMonths = *req.ContractTermMonths
	}

	modifyRes, err := svc.Client.ProductService.ModifyProduct(ctx, modifyReq)
	if err != nil {
		return nil, err
	}
	toReturn := &ModifyPortResponse{
		IsUpdated: modifyRes.IsUpdated,
	}

	// wait until the Port is updated before returning if requested by the user
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
				return nil, fmt.Errorf("time expired waiting for Port %s to update", req.PortID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for Port %s to update", req.PortID)
			case <-ticker.C:
				portDetails, err := svc.GetPort(ctx, req.PortID)
				if err != nil {
					return nil, err
				}
				if slices.Contains(SERVICE_STATE_READY, portDetails.ProvisioningStatus) {
					return toReturn, nil
				}
			}
		}
	} else {
		// return the response right away if the user doesn't want to wait for update
		return toReturn, nil
	}
}

// DeletePort deletes a port in the Megaport Port API.
func (svc *PortServiceOp) DeletePort(ctx context.Context, req *DeletePortRequest) (*DeletePortResponse, error) {
	_, err := svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID:  req.PortID,
		DeleteNow:  req.DeleteNow,
		SafeDelete: req.SafeDelete,
	})
	if err != nil {
		return nil, err
	}
	return &DeletePortResponse{
		IsDeleting: true,
	}, nil
}

// RestorePort restores a port in the Megaport Port API.
func (svc *PortServiceOp) RestorePort(ctx context.Context, portId string) (*RestorePortResponse, error) {
	_, err := svc.Client.ProductService.RestoreProduct(ctx, portId)
	if err != nil {
		return nil, err
	}
	return &RestorePortResponse{
		IsRestored: true,
	}, nil
}

// LockPort locks a port in the Megaport Port API.
func (svc *PortServiceOp) LockPort(ctx context.Context, portId string) (*LockPortResponse, error) {
	port, err := svc.GetPort(ctx, portId)
	if err != nil {
		return nil, err
	}
	if !port.Locked {
		_, err = svc.Client.ProductService.ManageProductLock(ctx, &ManageProductLockRequest{
			ProductID:  portId,
			ShouldLock: true,
		})
		if err != nil {
			return nil, err
		}
		return &LockPortResponse{IsLocking: true}, nil
	} else {
		return nil, ErrPortAlreadyLocked
	}
}

// UnlockPort unlocks a port in the Megaport Port API.
func (svc *PortServiceOp) UnlockPort(ctx context.Context, portId string) (*UnlockPortResponse, error) {
	port, err := svc.GetPort(ctx, portId)
	if err != nil {
		return nil, err
	}
	if port.Locked {
		_, err = svc.Client.ProductService.ManageProductLock(ctx, &ManageProductLockRequest{
			ProductID:  portId,
			ShouldLock: false,
		})
		if err != nil {
			return nil, err
		}
		return &UnlockPortResponse{IsUnlocking: true}, nil
	} else {
		return nil, ErrPortNotLocked
	}
}

func (svc *PortServiceOp) CheckPortVLANAvailability(ctx context.Context, portId string, vlan int) (bool, error) {
	if vlan < 0 || vlan > 4094 {
		return false, ErrInvalidVLAN
	}

	_, err := svc.GetPort(ctx, portId)
	if err != nil {
		return false, err
	}

	path := "/v2/product/port/" + portId + "/vlan"
	params := url.Values{}
	params.Add("vlan", fmt.Sprintf("%d", vlan))
	url := svc.Client.BaseURL.JoinPath(path)
	url.RawQuery = params.Encode()
	urlStr := url.String()

	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return false, err
	}

	response, err := svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return false, err
	}
	defer response.Body.Close()

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return false, err
	}

	vlanResponse := PortVLANAvailabilityAPIResponse{}
	unmarshalErr := json.Unmarshal(body, &vlanResponse)
	if unmarshalErr != nil {
		return false, unmarshalErr
	}

	if slices.Contains(vlanResponse.Data, vlan) {
		return true, nil
	}
	return false, nil
}

// ListPortResourceTags lists the resource tags for a port in the Megaport Port API.
func (svc *PortServiceOp) ListPortResourceTags(ctx context.Context, portID string) (map[string]string, error) {
	tags, err := svc.Client.ProductService.ListProductResourceTags(ctx, portID)
	if err != nil {
		return nil, err
	}
	return fromProductResourceTags(tags), nil
}

func (svc *PortServiceOp) UpdatePortResourceTags(ctx context.Context, portID string, tags map[string]string) error {
	productTags := toProductResourceTags(tags)
	return svc.Client.ProductService.UpdateProductResourceTags(ctx, portID, &UpdateProductResourceTagsRequest{
		ResourceTags: productTags,
	})
}

// GetPortLOA retrieves the LOA (Letter of Authorization) for a specific port
// The response is returned as bytes which can be used to:
// - Save directly to a file
// - Convert to base64 for storing in state
// - Process in any other way needed
func (svc *PortServiceOp) GetPortLOA(ctx context.Context, portID string) ([]byte, error) {
	// Ports are products, so we can reuse the product LOA endpoint
	return svc.Client.ProductService.GetProductLOA(ctx, portID)
}

// SavePortLOAToFile retrieves the LOA for a port and saves it to the specified file path
func (svc *PortServiceOp) SavePortLOAToFile(ctx context.Context, portID string, filePath string) error {
	// Reuse the product service implementation
	return svc.Client.ProductService.SaveProductLOAToFile(ctx, portID, filePath)
}

// GetPortLOABase64 retrieves the LOA for a port and returns it as a base64-encoded string
func (svc *PortServiceOp) GetPortLOABase64(ctx context.Context, portID string) (string, error) {
	// Reuse the product service implementation
	return svc.Client.ProductService.GetProductLOABase64(ctx, portID)
}
