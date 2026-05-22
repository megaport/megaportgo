package megaport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// Packet filter validation errors.
var (
	ErrNATGatewayPacketFilterIDRequired       = errors.New("packet filter ID must be greater than 0")
	ErrNATGatewayPacketFilterDescriptionEmpty = errors.New("packet filter description is required")
	ErrNATGatewayPacketFilterEntriesEmpty     = errors.New("packet filter requires at least one entry")
)

func validateNATGatewayPacketFilterRequest(req *NATGatewayPacketFilterRequest) error {
	if req == nil || req.Description == "" {
		return ErrNATGatewayPacketFilterDescriptionEmpty
	}
	if len(req.Entries) == 0 {
		return ErrNATGatewayPacketFilterEntriesEmpty
	}
	return nil
}

// ListNATGatewayPacketFilters returns all packet filter summaries for a NAT Gateway.
func (svc *NATGatewayServiceOp) ListNATGatewayPacketFilters(ctx context.Context, productUID string) ([]*NATGatewayPacketFilterSummary, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filter_summaries", url.PathEscape(productUID))
	var envelope natGatewayPacketFilterSummariesResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// CreateNATGatewayPacketFilter creates a new packet filter on a NAT Gateway.
func (svc *NATGatewayServiceOp) CreateNATGatewayPacketFilter(ctx context.Context, productUID string, req *NATGatewayPacketFilterRequest) (*NATGatewayPacketFilter, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if err := validateNATGatewayPacketFilterRequest(req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filters", url.PathEscape(productUID))
	var envelope natGatewayPacketFilterResponse
	if err := svc.doJSON(ctx, http.MethodPost, path, req, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// GetNATGatewayPacketFilter returns a packet filter by its numeric ID.
func (svc *NATGatewayServiceOp) GetNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) (*NATGatewayPacketFilter, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if packetFilterID < 1 {
		return nil, ErrNATGatewayPacketFilterIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filters/%d", url.PathEscape(productUID), packetFilterID)
	var envelope natGatewayPacketFilterResponse
	if err := svc.doJSON(ctx, http.MethodGet, path, nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// UpdateNATGatewayPacketFilter replaces a packet filter's description and entries.
func (svc *NATGatewayServiceOp) UpdateNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int, req *NATGatewayPacketFilterRequest) (*NATGatewayPacketFilter, error) {
	if productUID == "" {
		return nil, ErrNATGatewayProductUIDRequired
	}
	if packetFilterID < 1 {
		return nil, ErrNATGatewayPacketFilterIDRequired
	}
	if err := validateNATGatewayPacketFilterRequest(req); err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filters/%d", url.PathEscape(productUID), packetFilterID)
	var envelope natGatewayPacketFilterResponse
	if err := svc.doJSON(ctx, http.MethodPut, path, req, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// DeleteNATGatewayPacketFilter removes a packet filter from a NAT Gateway.
func (svc *NATGatewayServiceOp) DeleteNATGatewayPacketFilter(ctx context.Context, productUID string, packetFilterID int) error {
	if productUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	if packetFilterID < 1 {
		return ErrNATGatewayPacketFilterIDRequired
	}
	path := fmt.Sprintf("/v3/products/nat_gateways/%s/packet_filters/%d", url.PathEscape(productUID), packetFilterID)
	return svc.doJSON(ctx, http.MethodDelete, path, nil, nil)
}
