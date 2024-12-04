package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// ServiceKeyService is an interface for interfacing with the Service Key endpoints in the Megaport Service Key API.
type ServiceKeyService interface {
	// CreateServiceKey creates a service key in the Megaport Service Key API.
	CreateServiceKey(ctx context.Context, req *CreateServiceKeyRequest) (*CreateServiceKeyResponse, error)
	// ListServiceKeys lists service keys in the Megaport Service Key API.
	ListServiceKeys(ctx context.Context, req *ListServiceKeysRequest) (*ListServiceKeysResponse, error)
	// UpdateServiceKey updates a service key in the Megaport Service Key API.
	UpdateServiceKey(ctx context.Context, req *UpdateServiceKeyRequest) (*UpdateServiceKeyResponse, error)
	// GetServiceKey gets a service key in the Megaport Service Key API.
	GetServiceKey(ctx context.Context, keyId string) (*ServiceKey, error)
}

// NewServiceKeyService creates a new instance of the Service Key Service.
func NewServiceKeyService(c *Client) *ServiceKeyServiceOp {
	return &ServiceKeyServiceOp{
		Client: c,
	}
}

var _ ServiceKeyService = &ServiceKeyServiceOp{}

// ServiceKeyServiceOp handles communication with the Service Key related methods of the Megaport API.
type ServiceKeyServiceOp struct {
	Client *Client
}

type ServiceKey struct {
	Key         string    `json:"key"`
	CreateDate  *Time     `json:"createDate"`
	CompanyID   int       `json:"companyId"`
	CompanyUID  string    `json:"companyUid"`
	CompanyName string    `json:"companyName"`
	Description string    `json:"description"`
	ProductID   int       `json:"productId"`
	ProductUID  string    `json:"productUid"`
	ProductName string    `json:"productName"`
	VLAN        int       `json:"vlan"`
	MaxSpeed    int       `json:"maxSpeed"`
	PreApproved bool      `json:"preApproved"`
	SingleUse   bool      `json:"singleUse"`
	LastUsed    *Time     `json:"lastUsed"`
	Active      bool      `json:"active"`
	ValidFor    *ValidFor `json:"validFor"`
	Expired     bool      `json:"expired"`
	Valid       bool      `json:"valid"`
	PromoCode   string    `json:"promoCode"`
}

// CreateServiceKeyRequest represents a request to create a service key from the Megaport Service Key API.
type CreateServiceKeyRequest struct {
	ProductUID    string         `json:"productUid,omitempty"` // The Port ID for the service key. API can take either UID or ID.
	ProductID     int            `json:"productId,omitempty"`  // The Port UID for the service key. API can take either UID or ID.
	SingleUse     bool           `json:"singleUse"`            // Determines whether to create a single-use or multi-use service key. Valid values are true (single-use) and false (multi-use). With a multi-use key, the customer that you share the key with can request multiple connections using that key. For single-use keys only, specify a VLAN ID (vlan).
	MaxSpeed      int            `json:"maxSpeed"`
	Active        bool           `json:"active,omitempty"`      // Determines whether the service key is available for use. Valid values are true if you want the key to be available right away and false if you don’t want the key to be available right away.
	PreApproved   bool           `json:"preApproved,omitempty"` // Whether the service key is pre-approved for use.
	Description   string         `json:"description,omitempty"` // A description for the service key.
	VLAN          int            `json:"vlan,omitempty"`        // The VLAN ID for the service key. Required for single-use keys only.
	OrderValidFor *OrderValidFor `json:"validFor,omitempty"`    // The ValidFor field parsed for the Megaport API
	ValidFor      *ValidFor      // The range of dates for which the service key is valid.
}

// CreateServiceKeyAPIResponse represents a response from creating a service key from the Megaport Service Key API.
type CreateServiceKeyAPIResponse struct {
	Message string                          `json:"message"`
	Terms   string                          `json:"terms"`
	Data    CreateServiceKeyAPIResponseData `json:"data"`
}

// CreateServiceKeyAPIResponseData represents the data field in the CreateServiceKeyAPIResponse.
type CreateServiceKeyAPIResponseData struct {
	Key string `json:"key"`
}

// ValidFor represents the valid times for the service key
type ValidFor struct {
	StartTime *Time `json:"start"` // Parsed for Megaport API
	EndTime   *Time `json:"end"`   // Parsed for Megaport API
}

// OrderValidFor represents the ValidFor input with the Megaport API using integer values
type OrderValidFor struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

// CreateServiceKeyResponse represents a response from creating a service key from the Megaport Service Key API.
type CreateServiceKeyResponse struct {
	ServiceKeyUID string
}

// ListServiceKeysRequest represents a request to list service keys from the Megaport Service Key API.
type ListServiceKeysRequest struct {
	ProductUID *string // List keys linked to the Port specified by the product ID or UID. (Optional)
}

// GetServiceKeyAPIResponse represents the Megaport API HTTP response from getting a service key from the Megaport Service Key API.
type GetServiceKeyAPIResponse struct {
	Message string      `json:"message"`
	Terms   string      `json:"terms"`
	Data    *ServiceKey `json:"data"`
}

// ListServiceKeysAPIResponse represents the Megaport API HTTP response from listing service keys from the Megaport Service Key API.
type ListServiceKeysAPIResponse struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    []*ServiceKey `json:"data"`
}

// ListServiceKeysResponse represents the Go SDK response from listing service keys from the Megaport Service Key API.
type ListServiceKeysResponse struct {
	ServiceKeys []*ServiceKey
}

// UpdateServiceKeyRequest represents a request to update a service key in the Megaport Service Key API.
type UpdateServiceKeyRequest struct {
	Key           string         `json:"key"`
	ProductUID    string         `json:"productUid,omitempty"` // The Product UID for the service key. API can take either UID or ID.
	ProductID     int            `json:"productId,omitempty"`  // The Product ID for the service key. API can take either UID or ID.
	SingleUse     bool           `json:"singleUse"`            // Determines whether the service key is single-use or multi-use. Valid values are true (single-use) and false (multi-use). With a multi-use key, the customer that you share the key with can request multiple connections using that key.
	Active        bool           `json:"active"`               // Determines whether the service key is available for use. Valid values are true if you want the key to be available right away and false if you don’t want the key to be available right away.
	OrderValidFor *OrderValidFor `json:"validFor,omitempty"`   // The range of dates for which the service key is valid.
	ValidFor      *ValidFor
}

// UpdateServiceKeyResponse represents a response from updating a service key in the Megaport Service Key API.
type UpdateServiceKeyResponse struct {
	IsUpdated bool
}

// CreateServiceKey creates a service key in the Megaport Service Key API.
func (svc *ServiceKeyServiceOp) CreateServiceKey(ctx context.Context, req *CreateServiceKeyRequest) (*CreateServiceKeyResponse, error) {
	if req.ValidFor != nil {
		req.OrderValidFor = &OrderValidFor{
			Start: req.ValidFor.StartTime.Unix() * 1000,
			End:   req.ValidFor.EndTime.Unix() * 1000,
		}
	}
	path := "/v2/service/key"
	url := svc.Client.BaseURL.JoinPath(path).String()
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPost, url, req)
	if err != nil {
		return nil, err
	}
	response, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return nil, resErr
	}

	if response != nil {
		svc.Client.Logger.DebugContext(ctx, "Ordering Service Key", slog.String("url", url), slog.Int("status_code", response.StatusCode))
		defer response.Body.Close()
	}

	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}
	var createServiceKeyAPIResponse CreateServiceKeyAPIResponse
	if err = json.Unmarshal(body, &createServiceKeyAPIResponse); err != nil {
		return nil, err
	}
	toReturn := &CreateServiceKeyResponse{
		ServiceKeyUID: createServiceKeyAPIResponse.Data.Key,
	}
	return toReturn, nil
}

func (svc *ServiceKeyServiceOp) ListServiceKeys(ctx context.Context, req *ListServiceKeysRequest) (*ListServiceKeysResponse, error) {
	path := "/v2/service/key"
	params := url.Values{}
	if req.ProductUID != nil {
		params.Add("productIdOrUid", *req.ProductUID)
	}
	url := svc.Client.BaseURL.JoinPath(path)
	if len(params) > 0 {
		url.RawQuery = params.Encode()
	}
	urlString := url.String()
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, urlString, req)
	if err != nil {
		return nil, err
	}
	response, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return nil, resErr
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	listServiceKeysAPIResponse := ListServiceKeysAPIResponse{}
	if err = json.Unmarshal(body, &listServiceKeysAPIResponse); err != nil {
		return nil, err
	}
	toReturn := &ListServiceKeysResponse{}
	for _, key := range listServiceKeysAPIResponse.Data {
		toAppend := &ServiceKey{
			Key:         key.Key,
			CreateDate:  key.CreateDate,
			CompanyID:   key.CompanyID,
			CompanyUID:  key.CompanyUID,
			CompanyName: key.CompanyName,
			Description: key.Description,
			ProductID:   key.ProductID,
			ProductUID:  key.ProductUID,
			ProductName: key.ProductName,
			VLAN:        key.VLAN,
			MaxSpeed:    key.MaxSpeed,
			PreApproved: key.PreApproved,
			SingleUse:   key.SingleUse,
			LastUsed:    key.LastUsed,
			Active:      key.Active,
			ValidFor:    key.ValidFor,
		}
		toReturn.ServiceKeys = append(toReturn.ServiceKeys, toAppend)
	}
	return toReturn, nil
}

func (svc *ServiceKeyServiceOp) GetServiceKey(ctx context.Context, keyId string) (*ServiceKey, error) {
	path := fmt.Sprintf("/v2/service/key?key=%s", keyId)
	url := svc.Client.BaseURL.JoinPath(path).String()
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	response, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return nil, resErr
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	parsedAPIResponse := GetServiceKeyAPIResponse{}
	if err = json.Unmarshal(body, &parsedAPIResponse); err != nil {
		return nil, err
	}
	return parsedAPIResponse.Data, nil
}

func (svc *ServiceKeyServiceOp) UpdateServiceKey(ctx context.Context, req *UpdateServiceKeyRequest) (*UpdateServiceKeyResponse, error) {
	path := "/v2/service/key"
	url := svc.Client.BaseURL.JoinPath(path).String()
	if req.ValidFor != nil {
		req.OrderValidFor = &OrderValidFor{
			Start: req.ValidFor.StartTime.Unix() * 1000,
			End:   req.ValidFor.EndTime.Unix() * 1000,
		}
	}
	clientReq, err := svc.Client.NewRequest(ctx, http.MethodPut, url, req)
	if err != nil {
		return nil, err
	}
	_, resErr := svc.Client.Do(ctx, clientReq, nil)
	if resErr != nil {
		return nil, resErr
	}
	return &UpdateServiceKeyResponse{
		IsUpdated: true,
	}, nil
}
