package megaport

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// ErrNATGatewayProductUIDRequired is returned when a ProductUID is not provided.
var ErrNATGatewayProductUIDRequired = errors.New("product UID is required")

// ErrNATGatewayTelemetryTypesRequired is returned when no telemetry types are provided.
var ErrNATGatewayTelemetryTypesRequired = errors.New("at least one telemetry type is required")

// ErrNATGatewayTelemetryTimeExclusive is returned when both Days and From/To are provided.
var ErrNATGatewayTelemetryTimeExclusive = errors.New("days and from/to are mutually exclusive")

// ErrNATGatewayTelemetryDaysOutOfRange is returned when Days is not between 1 and 180.
var ErrNATGatewayTelemetryDaysOutOfRange = errors.New("days must be between 1 and 180")

// ErrNATGatewayTelemetryFromToIncomplete is returned when only one of From/To is provided.
var ErrNATGatewayTelemetryFromToIncomplete = errors.New("both from and to must be provided together")

// ErrNATGatewayProductNameRequired is returned when a ProductName is not provided.
var ErrNATGatewayProductNameRequired = errors.New("product name is required")

// ErrNATGatewayLocationIDRequired is returned when a LocationID is not provided or is invalid.
var ErrNATGatewayLocationIDRequired = errors.New("location ID must be greater than 0")

// ErrNATGatewaySpeedRequired is returned when a Speed is not provided or is invalid.
var ErrNATGatewaySpeedRequired = errors.New("speed must be greater than 0")

// ErrNATGatewayInvalidTerm is returned when a Term is not a valid contract term.
// The message is derived from VALID_CONTRACT_TERMS so it stays in sync if the
// allowed set ever changes.
var ErrNATGatewayInvalidTerm = fmt.Errorf("term must be one of: %s", formatValidContractTerms())

func formatValidContractTerms() string {
	parts := make([]string, len(VALID_CONTRACT_TERMS))
	for i, t := range VALID_CONTRACT_TERMS {
		parts[i] = strconv.Itoa(t)
	}
	return strings.Join(parts, ", ")
}

// validateCreateNATGatewayRequest validates the request parameters for creating a NAT Gateway.
func validateCreateNATGatewayRequest(req *CreateNATGatewayRequest) error {
	if req == nil {
		return ErrNATGatewayProductNameRequired
	}
	if req.ProductName == "" {
		return ErrNATGatewayProductNameRequired
	}
	if req.LocationID < 1 {
		return ErrNATGatewayLocationIDRequired
	}
	if req.Speed < 1 {
		return ErrNATGatewaySpeedRequired
	}
	if !slices.Contains(VALID_CONTRACT_TERMS, req.Term) {
		return ErrNATGatewayInvalidTerm
	}
	return nil
}

// validateUpdateNATGatewayRequest validates the request parameters for updating a NAT Gateway.
func validateUpdateNATGatewayRequest(req *UpdateNATGatewayRequest) error {
	if req == nil {
		return ErrNATGatewayProductUIDRequired
	}
	if req.ProductUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	if req.ProductName == "" {
		return ErrNATGatewayProductNameRequired
	}
	if req.LocationID < 1 {
		return ErrNATGatewayLocationIDRequired
	}
	if req.Speed < 1 {
		return ErrNATGatewaySpeedRequired
	}
	if !slices.Contains(VALID_CONTRACT_TERMS, req.Term) {
		return ErrNATGatewayInvalidTerm
	}
	return nil
}

// validateGetNATGatewayTelemetryRequest validates the request parameters.
func validateGetNATGatewayTelemetryRequest(req *GetNATGatewayTelemetryRequest) error {
	if req == nil {
		return ErrNATGatewayProductUIDRequired
	}
	if req.ProductUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	if len(req.Types) == 0 {
		return ErrNATGatewayTelemetryTypesRequired
	}
	if req.Days != nil && (req.From != nil || req.To != nil) {
		return ErrNATGatewayTelemetryTimeExclusive
	}
	if req.Days != nil && (*req.Days < 1 || *req.Days > 180) {
		return ErrNATGatewayTelemetryDaysOutOfRange
	}
	if (req.From != nil) != (req.To != nil) {
		return ErrNATGatewayTelemetryFromToIncomplete
	}
	return nil
}
