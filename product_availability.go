package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ProductAvailabilityService is an interface for interfacing with the Product Availability endpoints of the Megaport API.
type ProductAvailabilityService interface {
	// ListCompanyProductAvailability retrieves the products a company can order and the markets each is available in.
	ListCompanyProductAvailability(ctx context.Context, companyUID string) ([]*CompanyProductAvailability, error)
}

// ProductAvailabilityServiceOp handles communication with the Product Availability related methods of the Megaport API.
type ProductAvailabilityServiceOp struct {
	Client *Client
}

// NewProductAvailabilityService returns a ProductAvailabilityService
func NewProductAvailabilityService(c *Client) *ProductAvailabilityServiceOp {
	return &ProductAvailabilityServiceOp{Client: c}
}

// ListCompanyProductAvailability retrieves the products a company can order and the markets each is available in.
// A product the company cannot order anywhere is absent rather than listed as unavailable. Markets are not
// filtered to the company's active billing markets. A non-staff caller may only query their own company.
func (svc *ProductAvailabilityServiceOp) ListCompanyProductAvailability(ctx context.Context, companyUID string) ([]*CompanyProductAvailability, error) {
	if companyUID == "" {
		return nil, ErrCompanyUIDRequired
	}

	path := fmt.Sprintf("/v1/availability/companies/%s/products", url.PathEscape(companyUID))
	apiURL := svc.Client.BaseURL.JoinPath(path).String()

	req, err := svc.Client.NewRequest(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.Client.Do(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	availabilityResp := CompanyProductAvailabilityAPIResponse{}
	if err := json.NewDecoder(response.Body).Decode(&availabilityResp); err != nil {
		return nil, fmt.Errorf("failed to decode ListCompanyProductAvailability response: %w", err)
	}

	if availabilityResp.Data == nil {
		return nil, ErrCompanyProductAvailabilityResponseEmpty
	}

	return availabilityResp.Data, nil
}
