package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

// IXService is an interface for interacting with the IX endpoints of the Megaport API
type IXService interface {
	// GetIX retrieves details about a specific Internet Exchange by ID
	GetIX(ctx context.Context, id string) (*IX, error)

	// BuyIX purchases a new Internet Exchange
	BuyIX(ctx context.Context, req *BuyIXRequest) (*BuyIXResponse, error)

	// ValidateIXOrder validates an Internet Exchange order without submitting it
	ValidateIXOrder(ctx context.Context, req *BuyIXRequest) error

	// UpdateIX updates an existing Internet Exchange
	UpdateIX(ctx context.Context, id string, req *UpdateIXRequest) (*IX, error)

	// DeleteIX deletes an Internet Exchange
	DeleteIX(ctx context.Context, id string, req *DeleteIXRequest) error
}

// IXServiceOp handles communication with the IX related methods of the Megaport API
type IXServiceOp struct {
	Client *Client
}

func NewIXService(c *Client) IXService {
	return &IXServiceOp{
		Client: c,
	}
}

type BuyIXRequest struct {
	ProductUID         string        `json:"productUid"`         // The productUid of the port to attach the IX to
	Name               string        `json:"productName"`        // Name of the IX
	NetworkServiceType string        `json:"networkServiceType"` // The IX type/network service to connect to (e.g. "Los Angeles IX")
	ASN                int           `json:"asn"`                // ASN (Autonomous System Number) for BGP peering
	MACAddress         string        `json:"macAddress"`         // MAC address for the IX interface
	RateLimit          int           `json:"rateLimit"`          // Rate limit in Mbps
	VLAN               int           `json:"vlan"`               // VLAN ID for the IX connection
	Shutdown           bool          `json:"shutdown"`           // Whether the IX is initially shut down (true) or enabled (false)
	WaitForProvision   bool          // Client-side option to wait until IX is provisioned before returning
	PromoCode          string        `json:"promoCode,omitempty"` // Optional promotion code for discounts. The code is not validated by the API, so if the code doesn't exist or doesn't work for the service, the call will still be successful.
	WaitForTime        time.Duration // Maximum duration to wait for provisioning
}

type BuyIXResponse struct {
	TechnicalServiceUID string `json:"technicalServiceUid,omitempty"` // Unique identifier for the newly created IX service
}

// UpdateIXRequest represents a request to update an existing IX
type UpdateIXRequest struct {
	Name           *string `json:"name,omitempty"`           // Name of the IX
	RateLimit      *int    `json:"rateLimit,omitempty"`      // Rate limit in Mbps
	CostCentre     *string `json:"costCentre,omitempty"`     // For invoicing purposes
	VLAN           *int    `json:"vlan,omitempty"`           // VLAN ID for the IX connection
	MACAddress     *string `json:"macAddress,omitempty"`     // MAC address for the IX interface
	ASN            *int    `json:"asn,omitempty"`            // ASN (Autonomous System Number) - Be very careful about changing this
	Password       *string `json:"password,omitempty"`       // BGP password
	PublicGraph    *bool   `json:"publicGraph,omitempty"`    // Whether the IX usage statistics are publicly viewable
	ReverseDns     *string `json:"reverseDns,omitempty"`     // DNS lookup of a domain name from an IP address. You can change this value to enter a custom hostname for your IP address
	AEndProductUid *string `json:"aEndProductUid,omitempty"` // Move the IX by changing the A-End of the IX. Provide the productUid of the new A-End
	Shutdown       *bool   `json:"shutdown,omitempty"`       // Shut down and re-enable the IX. Valid values are true (shut down) and false (enabled). If not provided, defaults to false (enabled)

	WaitForUpdate bool          // Client-side option to wait until IX is updated before returning
	WaitForTime   time.Duration // Maximum duration to wait for updating
}

type DeleteIXRequest struct {
	DeleteNow bool // If true, delete immediately; if false, cancel at end of term
}

// BuyIX purchases a new Internet Exchange from the Megaport API
func (svc *IXServiceOp) BuyIX(ctx context.Context, req *BuyIXRequest) (*BuyIXResponse, error) {
	// Validate the order first
	err := svc.ValidateIXOrder(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert the request to an order
	order := ConvertBuyIXRequestToIXOrder(*req)

	// Execute the order through the product service
	responseBody, err := svc.Client.ProductService.ExecuteOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response
	var orderInfo struct {
		Data []struct {
			TechnicalServiceUID string `json:"technicalServiceUid"`
		} `json:"data"`
	}

	if err := json.Unmarshal(*responseBody, &orderInfo); err != nil {
		return nil, err
	}

	// Ensure we have data
	if len(orderInfo.Data) == 0 {
		return nil, fmt.Errorf("no IX created")
	}

	// Extract the technical service UID
	toReturn := &BuyIXResponse{
		TechnicalServiceUID: orderInfo.Data[0].TechnicalServiceUID,
	}

	// Wait until the IX is provisioned before returning if requested
	if req.WaitForProvision {
		toWait := req.WaitForTime
		if toWait == 0 {
			toWait = 5 * time.Minute
		}

		ticker := time.NewTicker(30 * time.Second) // check every 30 seconds
		timer := time.NewTimer(toWait)
		defer ticker.Stop()
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				return nil, fmt.Errorf("time expired waiting for IX %s to provision", toReturn.TechnicalServiceUID)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for IX %s to provision", toReturn.TechnicalServiceUID)
			case <-ticker.C:
				ix, err := svc.GetIX(ctx, toReturn.TechnicalServiceUID)
				if err != nil {
					return nil, err
				}

				if slices.Contains(SERVICE_STATE_READY, ix.ProvisioningStatus) {
					return toReturn, nil
				}
			}
		}
	}

	// Return the service UID right away if the user doesn't want to wait
	return toReturn, nil
}

// ValidateIXOrder validates an Internet Exchange order without submitting it
func (svc *IXServiceOp) ValidateIXOrder(ctx context.Context, req *BuyIXRequest) error {
	ixOrder := ConvertBuyIXRequestToIXOrder(*req)

	return svc.Client.ProductService.ValidateProductOrder(ctx, ixOrder)
}

// GetIX retrieves an IX by its ID
func (svc *IXServiceOp) GetIX(ctx context.Context, id string) (*IX, error) {
	path := fmt.Sprintf("/v2/product/%s", id)
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

	ixResponse := IXResponse{}

	unmarshalErr := json.Unmarshal(body, &ixResponse)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &ixResponse.Data, nil
}

// UpdateIX updates an existing Internet Exchange
func (svc *IXServiceOp) UpdateIX(ctx context.Context, id string, req *UpdateIXRequest) (*IX, error) {
	// Validate inputs
	if req.CostCentre != nil && len(*req.CostCentre) > 255 {
		return nil, ErrCostCentreTooLong
	}

	// Create the update path
	path := fmt.Sprintf("/v2/product/%s/%s", PRODUCT_IX, id)
	url := svc.Client.BaseURL.JoinPath(path).String()

	// Create the update structure
	update := &IXUpdate{}

	// Set fields from request
	if req.Name != nil {
		update.Name = *req.Name
	}
	if req.RateLimit != nil {
		update.RateLimit = req.RateLimit
	}
	if req.CostCentre != nil {
		update.CostCentre = *req.CostCentre
	}
	if req.VLAN != nil {
		update.VLAN = req.VLAN
	}
	if req.MACAddress != nil {
		update.MACAddress = *req.MACAddress
	}
	if req.ASN != nil {
		update.ASN = req.ASN
	}
	if req.Password != nil {
		update.Password = *req.Password
	}
	if req.PublicGraph != nil {
		update.PublicGraph = req.PublicGraph
	}
	if req.ReverseDns != nil {
		update.ReverseDns = *req.ReverseDns
	}
	if req.AEndProductUid != nil {
		update.AEndProductUid = *req.AEndProductUid
	}
	if req.Shutdown != nil {
		update.Shutdown = req.Shutdown
	}

	// Send the update request
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

	// Parse the response
	ixResponse := IXResponse{}
	if err = json.Unmarshal(body, &ixResponse); err != nil {
		return nil, err
	}

	// Wait for update to complete if requested
	if req.WaitForUpdate {
		toWait := req.WaitForTime
		if toWait == 0 {
			toWait = 5 * time.Minute
		}

		ticker := time.NewTicker(30 * time.Second) // check every 30 seconds
		timer := time.NewTimer(toWait)
		defer ticker.Stop()
		defer timer.Stop()

		for {
			select {
			case <-timer.C:
				return nil, fmt.Errorf("time expired waiting for IX %s to update", id)
			case <-ctx.Done():
				return nil, fmt.Errorf("context expired waiting for IX %s to update", id)
			case <-ticker.C:
				ix, err := svc.GetIX(ctx, id)
				if err != nil {
					return nil, err
				}

				if ix.ProvisioningStatus == "LIVE" || ix.ProvisioningStatus == "CONFIGURED" {
					return ix, nil
				}
			}
		}
	} else {
		// Return without waiting
		return &ixResponse.Data, nil
	}
}

// DeleteIX deletes an Internet Exchange
func (svc *IXServiceOp) DeleteIX(ctx context.Context, id string, req *DeleteIXRequest) error {
	_, err := svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID: id,
		DeleteNow: req.DeleteNow,
	})

	return err
}
