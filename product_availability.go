package megaport

import (
	"context"
	"encoding/json"
	"net/http"
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
// Only products available in at least one market are returned, and the markets are not filtered to the company's
// active billing markets. Requires the service_availability.view permission; a non-staff caller may only query
// their own company.
func (svc *ProductAvailabilityServiceOp) ListCompanyProductAvailability(ctx context.Context, companyUID string) ([]*CompanyProductAvailability, error) {
	path := "/v1/availability/companies/" + companyUID + "/products"
	url := svc.Client.BaseURL.JoinPath(path).String()

	req, err := svc.Client.NewRequest(ctx, http.MethodGet, url, nil)
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
		return nil, err
	}

	if availabilityResp.Data == nil {
		return nil, ErrCompanyProductAvailabilityResponseEmpty
	}

	return availabilityResp.Data, nil
}
