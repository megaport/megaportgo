package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// CompanyService is an interface for interfacing with the company endpoints of the Megaport API.
type CompanyService interface {
	// GetCurrentCompany retrieves the company the authenticated caller acts in.
	GetCurrentCompany(ctx context.Context) (*CurrentCompany, error)
}

// CompanyServiceOp handles communication with the company related methods of the Megaport API.
type CompanyServiceOp struct {
	Client *Client
}

// NewCompanyService returns a CompanyService
func NewCompanyService(c *Client) *CompanyServiceOp {
	return &CompanyServiceOp{Client: c}
}

// GetCurrentCompany retrieves the company the authenticated caller acts in, using GET /v2/me.
// When the client was created with WithCallContext, the company returned is whichever company
// the API resolves for that call context.
func (svc *CompanyServiceOp) GetCurrentCompany(ctx context.Context) (*CurrentCompany, error) {
	apiURL := svc.Client.BaseURL.JoinPath("/v2/me").String()

	req, err := svc.Client.NewRequest(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	resp := currentCompanyAPIResponse{}
	if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode GetCurrentCompany response: %w", err)
	}

	if resp.Data == nil || resp.Data.CompanyUID == "" {
		return nil, ErrCurrentCompanyResponseEmpty
	}

	return resp.Data, nil
}
