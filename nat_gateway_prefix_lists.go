package megaport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Prefix list validation errors.
var (
	ErrNATGatewayPrefixListIDRequired         = errors.New("prefix list ID must be greater than 0")
	ErrNATGatewayPrefixListDescriptionEmpty   = errors.New("prefix list description is required")
	ErrNATGatewayPrefixListAddressFamilyEmpty = errors.New("prefix list addressFamily is required")
	ErrNATGatewayPrefixListEntriesEmpty       = errors.New("prefix list requires at least one entry")
	ErrNATGatewayPrefixListEmptyResponse      = errors.New("API returned a 2xx response with an empty prefix list payload")
)

func validateNATGatewayPrefixList(req *NATGatewayPrefixList) error {
	if req == nil {
		return ErrNATGatewayRequestNil
	}
	if req.Description == "" {
		return ErrNATGatewayPrefixListDescriptionEmpty
	}
	if req.AddressFamily == "" {
		return ErrNATGatewayPrefixListAddressFamilyEmpty
	}
	if len(req.Entries) == 0 {
		return ErrNATGatewayPrefixListEntriesEmpty
	}
	return nil
}

// ListNATGatewayPrefixLists returns all prefix list summaries for a NAT Gateway.
func (svc *NATGatewayServiceOp) ListNATGatewayPrefixLists(ctx context.Context, productUID string) ([]*NATGatewayPrefixListSummary, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_list_summaries", url.PathEscape(productUID))
	var envelope natGatewayPrefixListSummariesResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// CreateNATGatewayPrefixList creates a new prefix list on a NAT Gateway.
func (svc *NATGatewayServiceOp) CreateNATGatewayPrefixList(ctx context.Context, productUID string, req *NATGatewayPrefixList) (*NATGatewayPrefixList, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if err := validateNATGatewayPrefixList(req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_lists", url.PathEscape(productUID))
	var envelope natGatewayPrefixListResponse
	if err := svc.doJSON(ctx, http.MethodPost, path, req.toAPI(), &envelope); err != nil {
		return nil, err
	}
	if envelope.Data == nil {
		return nil, ErrNATGatewayPrefixListEmptyResponse
	}
	return envelope.Data.toPrefixList()
}

// GetNATGatewayPrefixList returns a prefix list by its numeric ID.
func (svc *NATGatewayServiceOp) GetNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) (*NATGatewayPrefixList, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if prefixListID < 1 {
		return nil, ErrNATGatewayPrefixListIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_lists/%d", url.PathEscape(productUID), prefixListID)
	var envelope natGatewayPrefixListResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	if envelope.Data == nil {
		return nil, ErrNATGatewayPrefixListEmptyResponse
	}
	return envelope.Data.toPrefixList()
}

// UpdateNATGatewayPrefixList replaces a prefix list's description, address family, and entries.
func (svc *NATGatewayServiceOp) UpdateNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int, req *NATGatewayPrefixList) (*NATGatewayPrefixList, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if prefixListID < 1 {
		return nil, ErrNATGatewayPrefixListIDRequired
	}
	if err := validateNATGatewayPrefixList(req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_lists/%d", url.PathEscape(productUID), prefixListID)
	var envelope natGatewayPrefixListResponse
	if err := svc.doJSON(ctx, http.MethodPut, path, req.toAPI(), &envelope); err != nil {
		return nil, err
	}
	if envelope.Data == nil {
		return nil, ErrNATGatewayPrefixListEmptyResponse
	}
	return envelope.Data.toPrefixList()
}

// DeleteNATGatewayPrefixList removes a prefix list from a NAT Gateway.
func (svc *NATGatewayServiceOp) DeleteNATGatewayPrefixList(ctx context.Context, productUID string, prefixListID int) error {
	if productUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	if prefixListID < 1 {
		return ErrNATGatewayPrefixListIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/prefix_lists/%d", url.PathEscape(productUID), prefixListID)
	return svc.doJSON(ctx, http.MethodDelete, path, nil, nil)
}
