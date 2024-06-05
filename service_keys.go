package megaport

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

// ServiceKeyService is an interface for interfacing with the Service Key endpoints in the Megaport Service Key API.
type ServiceKeyService interface {
	// CreateServiceKey creates a service key in the Megaport Service Key API.
	CreateServiceKey(ctx context.Context, req *CreateServiceKeyRequest) (*CreateServiceKeyResponse, error)
	// ListServiceKeys lists service keys in the Megaport Service Key API.
	ListServiceKeys(ctx context.Context, req *ListServiceKeysRequest) (*ListServiceKeysResponse, error)
	// UpdateServiceKey updates a service key in the Megaport Service Key API.
	UpdateServiceKey(ctx context.Context, req *UpdateServiceKeyRequest) (*UpdateServiceKeyResponse, error)
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
	ProductUID  string    `json:"productUid"`
	SingleUse   bool      `json:"singleUse"`
	MaxSpeed    int       `json:"maxSpeed"`
	Active      bool      `json:"active,omitempty"`
	PreApproved bool      `json:"preApproved,omitempty"`
	Description string    `json:"description,omitempty"`
	VLAN        int       `json:"vlan,omitempty"`
	ValidFor    *ValidFor `json:"validFor"`
}

// CreateServiceKeyAPIResponse represents a response from creating a service key from the Megaport Service Key API.
type CreateServiceKeyAPIResponse struct {
	Message string                           `json:"message"`
	Terms   string                           `json:"terms"`
	Data    *CreateServiceKeyAPIResponseData `json:"data"`
}

// CreateServiceKeyAPIResponseData represents the data field in the CreateServiceKeyAPIResponse.
type CreateServiceKeyAPIResponseData struct {
	Key string `json:"key"`
}

// ValidFor represents the validFor field in the CreateServiceKeyRequest.
type ValidFor struct {
	StartTime     *Time // Start time of the service key
	EndTime       *Time // End time of the service key
	StartUnixNano int64 `json:"start"` // Parsed for Megaport API
	EndUnixNano   int64 `json:"end"`   // Parsed for Megaport API
}

// CreateServiceKeyResponse represents a response from creating a service key from the Megaport Service Key API.
type CreateServiceKeyResponse struct {
	ServiceKeyUID string
}

// ListServiceKeysRequest represents a request to list service keys from the Megaport Service Key API.
type ListServiceKeysRequest struct {
	ProductUID *string // List keys linked to the Port specified by the product ID or UID. (Optional)
	Key        *string // Get details for the specified key. (Optional) You can use the first 8 digits of a key, or you can use the full value.
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
	Key       string    `json:"key"`
	ProductID int       `json:"productId"` // The Port for the service key.
	SingleUse bool      `json:"singleUse"` // Determines whether the service key is single-use or multi-use. Valid values are true (single-use) and false (multi-use). With a multi-use key, the customer that you share the key with can request multiple connections using that key.
	Active    bool      `json:"active"`    // Determines whether the service key is available for use. Valid values are true if you want the key to be available right away and false if you don’t want the key to be available right away.
	ValidFor  *ValidFor `json:"validFor"`  // The range of dates for which the service key is valid.
}

// UpdateServiceKeyResponse represents a response from updating a service key in the Megaport Service Key API.
type UpdateServiceKeyResponse struct {
	IsUpdated bool
}

// CreateServiceKey creates a service key in the Megaport Service Key API.
func (svc *ServiceKeyServiceOp) CreateServiceKey(ctx context.Context, req *CreateServiceKeyRequest) (*CreateServiceKeyResponse, error) {
	if req.ValidFor != nil {
		req.ValidFor.StartUnixNano = req.ValidFor.StartTime.UnixNano()
		req.ValidFor.EndUnixNano = req.ValidFor.EndTime.UnixNano()
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
	if req.Key != nil {
		params.Add("key", *req.Key)
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
	if response != nil {
		defer response.Body.Close()
	}
	body, fileErr := io.ReadAll(response.Body)
	if fileErr != nil {
		return nil, fileErr
	}
	var listServiceKeysAPIResponse ListServiceKeysAPIResponse
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
		toAppend.ValidFor.StartTime = &Time{}
		toAppend.ValidFor.StartTime.Time = time.Unix(toAppend.ValidFor.StartUnixNano/1000, 0)
		toAppend.ValidFor.EndTime = &Time{}
		toAppend.ValidFor.EndTime.Time = time.Unix(toAppend.ValidFor.EndUnixNano/1000, 0)
		toReturn.ServiceKeys = append(toReturn.ServiceKeys, toAppend)
	}
	return toReturn, nil
}

func (svc *ServiceKeyServiceOp) UpdateServiceKey(ctx context.Context, req *UpdateServiceKeyRequest) (*UpdateServiceKeyResponse, error) {
	path := "/v2/service/key"
	url := svc.Client.BaseURL.JoinPath(path).String()
	if req.ValidFor != nil {
		req.ValidFor.StartUnixNano = req.ValidFor.StartTime.UnixNano()
		req.ValidFor.EndUnixNano = req.ValidFor.EndTime.UnixNano()
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
