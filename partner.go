package megaport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

// PartnerService is an interface for interfacing with the Partner Port endpoints of the Megaport API.
type PartnerService interface {
	// ListPartnerMegaports gets a list of all partner megaports in the Megaport Marketplace via the Megaport API.
	ListPartnerMegaports(ctx context.Context) ([]*PartnerMegaport, error)
	// FilterPartnerMegaportByProductName filters a list of partner megaports by product name in the Megaport API.
	FilterPartnerMegaportByProductName(ctx context.Context, partners []*PartnerMegaport, productName string, exactMatch bool) ([]*PartnerMegaport, error)
	// FilterPartnerMegaportByConnectType filters a list of partner megaports by connect type in the Megaport API.
	FilterPartnerMegaportByConnectType(ctx context.Context, partners []*PartnerMegaport, connectType string, exactMatch bool) ([]*PartnerMegaport, error)
	// FilterPartnerMegaportByCompanyName filters a list of partner megaports by company name in the Megaport API.
	FilterPartnerMegaportByCompanyName(ctx context.Context, partners []*PartnerMegaport, companyName string, exactMatch bool) ([]*PartnerMegaport, error)
	// FilterPartnerMegaportByLocationId filters a list of partner megaports by location ID in the Megaport API.
	FilterPartnerMegaportByLocationId(ctx context.Context, partners []*PartnerMegaport, locationId int) ([]*PartnerMegaport, error)
	// FilterPartnerMegaportByDiversityZone filters a list of partner megaports by diversity zone in the Megaport API.
	FilterPartnerMegaportByDiversityZone(ctx context.Context, partners []*PartnerMegaport, diversityZone string, exactMatch bool) ([]*PartnerMegaport, error)
}

// NewPartnerService creates a new instance of the PartnerService.
func NewPartnerService(c *Client) *PartnerServiceOp {
	return &PartnerServiceOp{
		Client: c,
	}
}

// PartnerServiceOp handles communication with Partner Port methods of the Megaport API.
type PartnerServiceOp struct {
	Client *Client
}

// ListPartnerMegaports gets a list of all partner megaports in the Megaport Marketplace via the Megaport API.
func (svc *PartnerServiceOp) ListPartnerMegaports(ctx context.Context) ([]*PartnerMegaport, error) {
	partnerMegaportUrl := "/v2/dropdowns/partner/megaports"
	req, err := svc.Client.NewRequest(ctx, http.MethodGet, partnerMegaportUrl, nil)
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

	partnerMegaportResponse := PartnerMegaportResponse{}
	unmarshalErr := json.Unmarshal(body, &partnerMegaportResponse)
	if unmarshalErr != nil {
		partnerResp := PartnerMegaportResponse{}
		svc.Client.LogJSONUnmarshalError(ctx, string(body), "list_partner_megaports_response", partnerResp, unmarshalErr)
		return nil, unmarshalErr
	}

	return partnerMegaportResponse.Data, nil
}

// FilterPartnerMegaportByProductName filters a list of partner megaports by product name in the Megaport API.
func (svc *PartnerServiceOp) FilterPartnerMegaportByProductName(ctx context.Context, partners []*PartnerMegaport, productName string, exactMatch bool) ([]*PartnerMegaport, error) {
	toReturn := []*PartnerMegaport{}
	for _, partner := range partners {
		match := false
		if productName != "" {
			if exactMatch { // Exact Match
				if productName == partner.ProductName {
					match = true
				}
			} else {
				if fuzzy.Match(productName, partner.ProductName) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match && partner.VXCPermitted {
			toReturn = append(toReturn, partner)
		}
	}
	if len(toReturn) == 0 {
		return nil, ErrNoPartnerPortsFound
	}
	return toReturn, nil
}

// FilterPartnerMegaportByConnectType filters a list of partner megaports by connect type in the Megaport API.
func (svc *PartnerServiceOp) FilterPartnerMegaportByConnectType(ctx context.Context, partners []*PartnerMegaport, connectType string, exactMatch bool) ([]*PartnerMegaport, error) {
	toReturn := []*PartnerMegaport{}
	for _, partner := range partners {
		match := false
		if connectType != "" {
			if exactMatch { // Exact Match
				if connectType == partner.ConnectType {
					match = true
				}
			} else {
				if fuzzy.Match(connectType, partner.ConnectType) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match && partner.VXCPermitted {
			toReturn = append(toReturn, partner)
		}
	}
	if len(toReturn) == 0 {
		return nil, ErrNoPartnerPortsFound
	}
	return toReturn, nil
}

// FilterPartnerMegaportByCompanyName filters a list of partner megaports by company name in the Megaport API.
func (svc *PartnerServiceOp) FilterPartnerMegaportByCompanyName(ctx context.Context, partners []*PartnerMegaport, companyName string, exactMatch bool) ([]*PartnerMegaport, error) {
	toReturn := []*PartnerMegaport{}
	for _, partner := range partners {
		match := false
		if companyName != "" {
			if exactMatch { // Exact Match
				if companyName == partner.CompanyName {
					match = true
				}
			} else {
				if fuzzy.Match(companyName, partner.CompanyName) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match && partner.VXCPermitted {
			toReturn = append(toReturn, partner)
		}
	}
	if len(toReturn) == 0 {
		return nil, ErrNoPartnerPortsFound
	}
	return toReturn, nil
}

// FilterPartnerMegaportByLocationId filters a list of partner megaports by location ID in the Megaport API.
func (svc *PartnerServiceOp) FilterPartnerMegaportByLocationId(ctx context.Context, partners []*PartnerMegaport, locationId int) ([]*PartnerMegaport, error) {
	toReturn := []*PartnerMegaport{}
	for _, partner := range partners {
		if locationId >= 0 {
			if locationId == partner.LocationId && partner.VXCPermitted {
				toReturn = append(toReturn, partner)
			}
		} else {
			toReturn = append(toReturn, partner)
		}
	}
	if len(toReturn) == 0 {
		return nil, ErrNoPartnerPortsFound
	}
	return toReturn, nil
}

// FilterPartnerMegaportByDiversityZone filters a list of partner megaports by diversity zone in the Megaport API.
func (svc *PartnerServiceOp) FilterPartnerMegaportByDiversityZone(ctx context.Context, partners []*PartnerMegaport, diversityZone string, exactMatch bool) ([]*PartnerMegaport, error) {
	toReturn := []*PartnerMegaport{}
	for _, partner := range partners {
		match := false
		if diversityZone != "" {
			if exactMatch { // Exact Match
				if diversityZone == partner.DiversityZone {
					match = true
				}
			} else {
				if fuzzy.Match(diversityZone, partner.DiversityZone) {
					match = true
				}
			}
		} else {
			match = true
		}

		if match && partner.VXCPermitted {
			toReturn = append(toReturn, partner)
		}
	}
	if len(toReturn) == 0 {
		return nil, ErrNoPartnerPortsFound
	}
	return toReturn, nil
}
