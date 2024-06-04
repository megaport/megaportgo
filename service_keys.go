package megaport

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

// ServiceKeyService is an interface for interfacing with the Service Key endpoints in the Megaport Service Key API.
type ServiceKeyService interface {
	// CreateServiceKey creates a service key in the Megaport Service Key API.
	CreateServiceKey(ctx context.Context, req *CreateServiceKeyRequest) (*CreateServiceKeyResponse, error)
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

type CreateServiceKeyAPIResponse struct {
	Message string                           `json:"message"`
	Terms   string                           `json:"terms"`
	Data    *CreateServiceKeyAPIResponseData `json:"data"`
}

type CreateServiceKeyAPIResponseData struct {
	Key string `json:"key"`
}

type ValidFor struct {
	StartTime     *Time
	EndTime       *Time
	StartUnixNano int64 `json:"start"`
	EndUnixNano   int64 `json:"end"`
}

// CreateServiceKeyResponse represents a response from creating a service key from the Megaport Service Key API.
type CreateServiceKeyResponse struct {
	ServiceKeyUID string
}

func (svc *ServiceKeyServiceOp) CreateServiceKey(ctx context.Context, req *CreateServiceKeyRequest) (*CreateServiceKeyResponse, error) {
	if req.ValidFor != nil {
		req.ValidFor.StartUnixNano = req.ValidFor.StartTime.UnixNano()
		req.ValidFor.EndUnixNano = req.ValidFor.EndTime.UnixNano()
	}
	path := "/service/key"
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
