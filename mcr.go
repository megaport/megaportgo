package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"
)

// MCRService is an interface for interfacing with the MCR endpoints
// of the Megaport API.
type MCRService interface {
	// BuyMCR buys an MCR from the Megaport MCR API.
	BuyMCR(ctx context.Context, req *BuyMCRRequest) (*BuyMCRResponse, error)
	// ValidateMCROrder validates an MCR order in the Megaport Products API.
	ValidateMCROrder(ctx context.Context, req *BuyMCRRequest) error
	// ListMCRs lists all MCRs in the Megaport API. It allows you to filter by whether the provisioning status is active.
	ListMCRs(ctx context.Context, req *ListMCRsRequest) ([]*MCR, error)
	// GetMCR gets details about a single MCR from the Megaport MCR API.
	GetMCR(ctx context.Context, mcrId string) (*MCR, error)
	// CreatePrefixFilterList creates a Prefix Filter List on an MCR from the Megaport MCR API.
	CreatePrefixFilterList(ctx context.Context, req *CreateMCRPrefixFilterListRequest) (*CreateMCRPrefixFilterListResponse, error)
	// ListMCRPrefixFilterLists returns prefix filter lists for the specified MCR2 from the Megaport MCR API.
	ListMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*PrefixFilterList, error)
	// GetMCRPrefixFilterList returns a single prefix filter list by ID for the specified MCR2 from the Megaport MCR API.
	GetMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*MCRPrefixFilterList, error)
	// ModifyMCRPrefixFilterList modifies a prefix filter list on an MCR in the Megaport MCR API.
	ModifyMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int, prefixFilterList *MCRPrefixFilterList) (*ModifyMCRPrefixFilterListResponse, error)
	// DeleteMCRPrefixFilterList deletes a prefix filter list on an MCR from the Megaport MCR API.
	DeleteMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*DeleteMCRPrefixFilterListResponse, error)
	// ModifyMCR modifies an MCR in the Megaport MCR API.
	ModifyMCR(ctx context.Context, req *ModifyMCRRequest) (*ModifyMCRResponse, error)
	// DeleteMCR deletes an MCR in the Megaport MCR API.
	DeleteMCR(ctx context.Context, req *DeleteMCRRequest) (*DeleteMCRResponse, error)
	// RestoreMCR restores a deleted MCR in the Megaport MCR API.
	RestoreMCR(ctx context.Context, mcrId string) (*RestoreMCRResponse, error)
	// ListMCRResourceTags returns the resource tags for an MCR in the Megaport MCR API.
	ListMCRResourceTags(ctx context.Context, mcrID string) (map[string]string, error)
	// UpdateMCRResourceTags updates the resource tags for an MCR in the Megaport MCR API.
	UpdateMCRResourceTags(ctx context.Context, mcrID string, tags map[string]string) error
	// UpdateMCRWithAddOn adds an IPsec add-on to an existing MCR.
	UpdateMCRWithAddOn(ctx context.Context, mcrID string, req MCRAddOnRequest) error
	// UpdateMCRIPsecAddOn updates an existing IPsec add-on on an MCR. Setting tunnelCount to 0 will disable IPsec.
	UpdateMCRIPsecAddOn(ctx context.Context, mcrID string, addOnUID string, tunnelCount int) error
	// WaitForMCRReady polls until the MCR reaches a ready provisioning state.
	// A zero timeout defaults to 5 minutes. Returns ErrMCRNotFound or ErrMCRDecommissioned
	// if the MCR is deleted or decommissioned while polling.
	WaitForMCRReady(ctx context.Context, mcrID string, timeout time.Duration) error

	// GetMCRPrefixFilterLists returns prefix filter lists for the specified MCR2.
	//
	// Deprecated: Use ListMCRPrefixFilterLists instead.
	GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*PrefixFilterList, error)
}

// MCRServiceOp handles communication with MCR methods of the Megaport API.
type MCRServiceOp struct {
	Client *Client
}

// NewMCRService creates a new instance of the MCR Service.
func NewMCRService(c *Client) *MCRServiceOp {
	return &MCRServiceOp{
		Client: c,
	}
}

// BuyMCRRequest represents a request to buy an MCR
type BuyMCRRequest struct {
	LocationID    int
	Name          string
	DiversityZone string
	Term          int
	PortSpeed     int
	MCRAsn        int
	CostCentre    string
	PromoCode     string
	ResourceTags  map[string]string `json:"resourceTags,omitempty"`
	AddOns        []MCRAddOn        `json:"addOns,omitempty"`

	WaitForProvision bool          // Wait until the MCR provisions before returning
	WaitForTime      time.Duration // How long to wait for the MCR to provision if WaitForProvision is true (default is 5 minutes; must be at least 30 seconds for the poller to fire)
}

// BuyMCRResponse represents a response from buying an MCR
type BuyMCRResponse struct {
	TechnicalServiceUID string
}

// ListMCRsRequest represents a request to list MCRs. It allows you to filter by whether the provisioning status is active.
type ListMCRsRequest struct {
	IncludeInactive bool
}

// CreateMCRPrefixFilterListRequest represents a request to create a prefix filter list on an MCR
type CreateMCRPrefixFilterListRequest struct {
	MCRID            string
	PrefixFilterList MCRPrefixFilterList
}

// CreateMCRPrefixFilterListResponse represents a response from creating a prefix filter list on an MCR
type CreateMCRPrefixFilterListResponse struct {
	IsCreated          bool
	PrefixFilterListID int // The ID of the created prefix filter list
}

// ModifyMCRRequest represents a request to modify an MCR
type ModifyMCRRequest struct {
	MCRID                 string
	Name                  string
	CostCentre            string
	MarketplaceVisibility *bool
	ContractTermMonths    *int

	WaitForUpdate bool          // Wait until the MCR updates before returning
	WaitForTime   time.Duration // How long to wait for the MCR to update if WaitForUpdate is true (default is 5 minutes; must be at least 30 seconds for the poller to fire)
}

// ModifyMCRResponse represents a response from modifying an MCR
type ModifyMCRResponse struct {
	IsUpdated bool
}

// DeleteMCRRequest represents a request to delete an MCR
type DeleteMCRRequest struct {
	MCRID      string
	DeleteNow  bool
	SafeDelete bool // If true, the API will check whether the MCR has any attached resources before deleting it. If the MCR has attached resources, the API will return an error.
}

// DeleteMCRResponse represents a response from deleting an MCR
type DeleteMCRResponse struct {
	IsDeleting bool
}

// RestoreMCRequest represents a request to restore a deleted MCR
type RestoreMCRResponse struct {
	IsRestored bool
}

// ModifyMCRPrefixFilterListRequest represents a request to modify a prefix filter list on an MCR
type ModifyMCRPrefixFilterListResponse struct {
	IsUpdated bool
}

// DeleteMCRPrefixFilterListResponse represents a response from deleting a prefix filter list on an MCR
type DeleteMCRPrefixFilterListResponse struct {
	IsDeleted bool
}

type MCRAddOnRequest struct {
	AddOn MCRAddOn

	WaitForProvision bool          // Wait until the MCR reaches a ready state before returning
	WaitForTime      time.Duration // How long to wait if WaitForProvision is true (default is 5 minutes; must be at least 30 seconds for the poller to fire)
}

// BuyMCR purchases an MCR from the Megaport MCR API.
func (svc *MCRServiceOp) BuyMCR(ctx context.Context, req *BuyMCRRequest) (*BuyMCRResponse, error) {
	err := validateBuyMCRRequest(req)
	if err != nil {
		return nil, err
	}

	mcrOrders := createMCROrder(req)

	body, resErr := svc.Client.ProductService.ExecuteOrder(ctx, mcrOrders)

	if resErr != nil {
		return nil, resErr
	}

	orderInfo := mcrOrderResponse{}
	unmarshalErr := json.Unmarshal(*body, &orderInfo)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	toReturn := &BuyMCRResponse{
		TechnicalServiceUID: orderInfo.Data[0].TechnicalServiceUID,
	}

	// wait until the MCR is provisioned before returning if requested by the user.
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
				return nil, fmt.Errorf("context expired waiting for MCR %s to provision: %w", toReturn.TechnicalServiceUID, ctx.Err())
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

// validateBuyMCRRequest validates the BuyMCRRequest for a valid term and port speed.
func validateBuyMCRRequest(order *BuyMCRRequest) error {
	if !slices.Contains(VALID_CONTRACT_TERMS, order.Term) {
		return ErrInvalidTerm
	}
	if !slices.Contains(VALID_MCR_PORT_SPEEDS, order.PortSpeed) {
		return ErrMCRInvalidPortSpeed
	}

	// Validate add-ons
	for _, addOn := range order.AddOns {
		if err := validateMCRAddOn(addOn); err != nil {
			return err
		}
	}

	return nil
}

// validateMCRAddOn validates an MCR add-on configuration using the MCRAddOn interface.
func validateMCRAddOn(addOn MCRAddOn) error {
	switch t := addOn.(type) {
	case *MCRAddOnIPsecConfig:
		return validateIPsecAddOn(t)
	default:
		return ErrInvalidAddOnType
	}
}

// validateIPsecAddOn validates an IPsec add-on configuration.
func validateIPsecAddOn(addOn *MCRAddOnIPsecConfig) error {
	// Validate tunnel count (0 will default to 10; valid values are 10, 20, 30)
	if addOn.TunnelCount != 0 && !slices.Contains(ValidIPsecTunnelCounts, addOn.TunnelCount) {
		return ErrInvalidIPsecTunnelCount
	}

	return nil
}

func createMCROrder(req *BuyMCRRequest) []MCROrder {
	order := MCROrder{
		LocationID:   req.LocationID,
		Name:         req.Name,
		Term:         req.Term,
		Type:         "MCR2",
		PortSpeed:    req.PortSpeed,
		PromoCode:    req.PromoCode,
		ResourceTags: toProductResourceTags(req.ResourceTags),
		Config:       MCROrderConfig{},
	}

	if req.CostCentre != "" {
		order.CostCentre = req.CostCentre
	}

	order.Config.ASN = req.MCRAsn
	if req.DiversityZone != "" {
		order.Config.DiversityZone = req.DiversityZone
	}

	for _, addOn := range req.AddOns {
		order.AddOns = append(order.AddOns, applyAddOnDefaults(addOn))
	}

	return []MCROrder{order}
}

// applyAddOnDefaults applies default values to an MCR add-on based on its type.
func applyAddOnDefaults(addOn MCRAddOn) MCRAddOn {
	switch t := addOn.(type) {
	case *MCRAddOnIPsecConfig:
		if t.AddOnType == "" {
			t.AddOnType = AddOnTypeIPsec
		}
		if t.TunnelCount == 0 {
			t.TunnelCount = 10
		}
		if t.PackCount == 0 {
			t.PackCount = 1
		}
		return t
	default:
		return addOn
	}
}

func (svc *MCRServiceOp) ValidateMCROrder(ctx context.Context, req *BuyMCRRequest) error {
	err := validateBuyMCRRequest(req)
	if err != nil {
		return err
	}

	mcrOrders := createMCROrder(req)

	return svc.Client.ProductService.ValidateProductOrder(ctx, mcrOrders)
}

// ListMCRs lists all MCRs in the Megaport API.
func (svc *MCRServiceOp) ListMCRs(ctx context.Context, req *ListMCRsRequest) ([]*MCR, error) {
	allProducts, err := svc.Client.ProductService.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	mcrs := []*MCR{}

	for _, product := range allProducts {
		if strings.ToLower(product.GetType()) == PRODUCT_MCR {
			mcr, ok := product.(*MCR)
			if !ok {
				svc.Client.Logger.WarnContext(ctx, "Found MCR product type but couldn't cast to MCR struct")
				continue
			}

			// Filter inactive MCRs if requested
			if !req.IncludeInactive && (mcr.ProvisioningStatus == STATUS_DECOMMISSIONED || mcr.ProvisioningStatus == STATUS_CANCELLED) {
				continue
			}

			mcrs = append(mcrs, mcr)
		}
	}

	return mcrs, nil
}

// GetMCR returns the details of a single MCR in the Megaport MCR API.
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

	mcrRes := &mcrResponse{}
	unmarshalErr := json.Unmarshal(body, mcrRes)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return mcrRes.Data, nil
}

// CreatePrefixFilterList creates a Prefix Filter List on an MCR from the Megaport MCR API.
func (svc *MCRServiceOp) CreatePrefixFilterList(ctx context.Context, req *CreateMCRPrefixFilterListRequest) (*CreateMCRPrefixFilterListResponse, error) {
	url := "/v2/product/mcr2/" + req.MCRID + "/prefixList"

	clientReq, err := svc.Client.NewRequest(ctx, "POST", url, req.PrefixFilterList)

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

	createRes := &apiMCRPrefixFilterListResponse{}
	unmarshalErr := json.Unmarshal(body, createRes)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return &CreateMCRPrefixFilterListResponse{
		IsCreated:          true,
		PrefixFilterListID: createRes.Data.ID,
	}, nil
}

// GetMCRPrefixFilterLists returns prefix filter lists for the specified MCR2.
//
// Deprecated: Use ListMCRPrefixFilterLists instead.
func (svc *MCRServiceOp) GetMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*PrefixFilterList, error) {
	res, err := svc.ListMCRPrefixFilterLists(ctx, mcrId)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetMCRPrefixFilterLists returns prefix filter lists for the specified MCR2 from the Megaport MCR API.
func (svc *MCRServiceOp) ListMCRPrefixFilterLists(ctx context.Context, mcrId string) ([]*PrefixFilterList, error) {
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

	prefixFilterList := &listMCRPrefixFilterListResponse{}
	unmarshalErr := json.Unmarshal(body, prefixFilterList)

	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	return prefixFilterList.Data, nil
}

// GetMCRPrefixFilterList returns a single prefix filter list by ID for the specified MCR2 from the Megaport MCR API.
func (svc *MCRServiceOp) GetMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*MCRPrefixFilterList, error) {
	listID := strconv.Itoa(prefixFilterListID)
	url := "/v2/product/mcr2/" + mcrID + "/prefixList/" + listID

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

	apiPrefixFilterList := &apiMCRPrefixFilterListResponse{}
	unmarshalErr := json.Unmarshal(body, apiPrefixFilterList)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	prefixFilterList, err := apiPrefixFilterList.Data.ToMCRPrefixFilterList()
	if err != nil {
		return nil, err
	}
	return prefixFilterList, nil
}

// ModifyMCR modifies an MCR in the Megaport MCR API.
func (svc *MCRServiceOp) ModifyMCR(ctx context.Context, req *ModifyMCRRequest) (*ModifyMCRResponse, error) {
	if len(req.CostCentre) > 255 {
		return nil, ErrCostCentreTooLong
	}
	modifyReq := &ModifyProductRequest{
		ProductID:             req.MCRID,
		ProductType:           PRODUCT_MCR,
		Name:                  req.Name,
		CostCentre:            req.CostCentre,
		MarketplaceVisibility: req.MarketplaceVisibility,
	}
	if req.ContractTermMonths != nil {
		modifyReq.ContractTermMonths = *req.ContractTermMonths
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
				return nil, fmt.Errorf("context expired waiting for MCR %s to update: %w", req.MCRID, ctx.Err())
			case <-ticker.C:
				mcrDetails, err := svc.GetMCR(ctx, req.MCRID)
				if err != nil {
					return nil, err
				}
				if slices.Contains(SERVICE_STATE_READY, mcrDetails.ProvisioningStatus) {
					return toReturn, nil
				}
			}
		}
	} else {
		// return the response right away if the user doesn't want to wait for update
		return toReturn, nil
	}
}

// DeleteMCRPrefixFilterList deletes a prefix filter list on an MCR from the Megaport MCR API.
func (svc *MCRServiceOp) DeleteMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int) (*DeleteMCRPrefixFilterListResponse, error) {
	url := fmt.Sprintf("/v2/product/mcr2/%s/prefixList/%d", mcrID, prefixFilterListID)
	clientReq, err := svc.Client.NewRequest(ctx, "DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	_, err = svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	return &DeleteMCRPrefixFilterListResponse{
		IsDeleted: true,
	}, nil
}

// ModifyMCRPrefixFilterList modifies a prefix filter list on an MCR in the Megaport MCR API.
func (svc *MCRServiceOp) ModifyMCRPrefixFilterList(ctx context.Context, mcrID string, prefixFilterListID int, prefixFilterList *MCRPrefixFilterList) (*ModifyMCRPrefixFilterListResponse, error) {
	url := fmt.Sprintf("/v2/product/mcr2/%s/prefixList/%d", mcrID, prefixFilterListID)
	clientReq, err := svc.Client.NewRequest(ctx, "PUT", url, prefixFilterList)
	if err != nil {
		return nil, err
	}
	_, err = svc.Client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	return &ModifyMCRPrefixFilterListResponse{
		IsUpdated: true,
	}, nil
}

// DeleteMCR deletes an MCR in the Megaport MCR API.
// Note: MCR products only support immediate deletion (CANCEL_NOW). Requests with
// DeleteNow=false are rejected with ErrMCRCancelLaterNotAllowed, and accepted
// requests always call the underlying API with DeleteNow=true.
func (svc *MCRServiceOp) DeleteMCR(ctx context.Context, req *DeleteMCRRequest) (*DeleteMCRResponse, error) {
	if req == nil {
		return nil, ErrDeleteMCRRequestNil
	}
	// Enforce MCR lifecycle restriction: only CANCEL_NOW is allowed
	if !req.DeleteNow {
		return nil, ErrMCRCancelLaterNotAllowed
	}

	_, err := svc.Client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
		ProductID:  req.MCRID,
		DeleteNow:  true, // Always use CANCEL_NOW for MCR
		SafeDelete: req.SafeDelete,
	})
	if err != nil {
		return nil, err
	}
	return &DeleteMCRResponse{
		IsDeleting: true,
	}, nil
}

// Restore restores a deleted MCR in the Megaport MCR API.
func (svc *MCRServiceOp) RestoreMCR(ctx context.Context, mcrId string) (*RestoreMCRResponse, error) {
	_, err := svc.Client.ProductService.RestoreProduct(ctx, mcrId)
	if err != nil {
		return nil, err
	}
	return &RestoreMCRResponse{
		IsRestored: true,
	}, nil
}

// ListMCRResourceTags returns the resource tags for an MCR in the Megaport MCR API.
func (svc *MCRServiceOp) ListMCRResourceTags(ctx context.Context, mcrID string) (map[string]string, error) {
	tags, err := svc.Client.ProductService.ListProductResourceTags(ctx, mcrID)
	if err != nil {
		return nil, err
	}
	return fromProductResourceTags(tags), nil
}

// UpdateMCRResourceTags updates the resource tags for an MCR in the Megaport MCR API.
func (svc *MCRServiceOp) UpdateMCRResourceTags(ctx context.Context, mcrID string, tags map[string]string) error {
	return svc.Client.ProductService.UpdateProductResourceTags(ctx, mcrID, &UpdateProductResourceTagsRequest{
		ResourceTags: toProductResourceTags(tags),
	})
}

func (svc *MCRServiceOp) UpdateMCRWithAddOn(ctx context.Context, mcrID string, req MCRAddOnRequest) error {
	if req.AddOn == nil {
		return fmt.Errorf("AddOn cannot be nil")
	}

	if err := validateMCRAddOn(req.AddOn); err != nil {
		return err
	}

	url := fmt.Sprintf("/v3/product/%s/addon", mcrID)
	switch t := req.AddOn.(type) {
	case *MCRAddOnIPsecConfig:
		tunnelCount := t.TunnelCount
		if tunnelCount == 0 {
			tunnelCount = 10
		}

		payload := map[string]interface{}{
			"addOnType":   t.GetAddOnType(),
			"tunnelCount": tunnelCount,
		}

		clientReq, err := svc.Client.NewRequest(ctx, "POST", url, payload)
		if err != nil {
			return err
		}
		_, err = svc.Client.Do(ctx, clientReq, nil)
		if err != nil {
			return err
		}
	default:
		return ErrInvalidAddOnType
	}

	if req.WaitForProvision {
		return svc.WaitForMCRReady(ctx, mcrID, req.WaitForTime)
	}

	return nil
}

// UpdateMCRIPsecAddOn updates an existing IPsec add-on on an MCR.
// Set tunnelCount to 0 to disable the IPsec add-on.
// PUT /v3/product/{productUid}/addon/{addOnUid}
func (svc *MCRServiceOp) UpdateMCRIPsecAddOn(ctx context.Context, mcrID string, addOnUID string, tunnelCount int) error {
	// Validate tunnel count (0 to disable, or valid counts: 10, 20, 30)
	if tunnelCount != 0 && !slices.Contains(ValidIPsecTunnelCounts, tunnelCount) {
		return ErrInvalidIPsecTunnelCount
	}

	url := "/v3/product/" + mcrID + "/addon/" + addOnUID

	payload := map[string]interface{}{
		"addOnType":   AddOnTypeIPsec,
		"tunnelCount": tunnelCount,
	}

	clientReq, err := svc.Client.NewRequest(ctx, "PUT", url, payload)
	if err != nil {
		return err
	}
	_, err = svc.Client.Do(ctx, clientReq, nil)
	return err
}

// WaitForMCRReady polls until the MCR identified by mcrID reaches a ready
// provisioning state. A zero timeout defaults to 5 minutes.
// Returns ErrMCRNotFound if the MCR is deleted while polling, or
// ErrMCRDecommissioned if it has been decommissioned.
func (svc *MCRServiceOp) WaitForMCRReady(ctx context.Context, mcrID string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	check := func() (bool, error) {
		mcr, err := svc.GetMCR(ctx, mcrID)
		if err != nil {
			if IsServiceNotFoundError(err) {
				return false, ErrMCRNotFound
			}
			return false, err
		}
		if mcr.ProvisioningStatus == STATUS_DECOMMISSIONED {
			return false, ErrMCRDecommissioned
		}
		return slices.Contains(SERVICE_STATE_READY, mcr.ProvisioningStatus), nil
	}

	// Immediate pre-check before starting the ticker.
	if ready, err := check(); err != nil || ready {
		return err
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return fmt.Errorf("time expired waiting for MCR %s to reach a ready state", mcrID)
		case <-ticker.C:
			if ready, err := check(); err != nil || ready {
				return err
			}
		}
	}
}
