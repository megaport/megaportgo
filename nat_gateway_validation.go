package megaport

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

// ErrNATGatewayRequestNil is returned when a nil request is passed to a NAT
// Gateway validator. The previous code dereferenced these without a guard
// and would panic on a nil pointer; the validators now return this error
// instead.
var ErrNATGatewayRequestNil = errors.New("request must not be nil")

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
var ErrNATGatewayInvalidTerm = fmt.Errorf("term must be one of: %s", intSliceToString(VALID_CONTRACT_TERMS))

// ErrNATGatewaySpeedNotSupported is returned by ValidateNATGatewaySpeedSession
// when the requested speed is not present in the live availability matrix.
var ErrNATGatewaySpeedNotSupported = errors.New("nat gateway speed is not supported")

// ErrNATGatewaySessionCountNotSupported is returned by
// ValidateNATGatewaySpeedSession when the requested session count is not
// permitted for the requested speed.
var ErrNATGatewaySessionCountNotSupported = errors.New("nat gateway session count is not supported for the requested speed")

// validateNATGatewayCommonFields performs the structural checks shared by
// CreateNATGatewayRequest and UpdateNATGatewayRequest. It does not consult
// the live availability matrix; use ValidateNATGatewaySpeedSession for
// matrix-aware checks.
func validateNATGatewayCommonFields(productName string, locationID, speed, term int) error {
	if productName == "" {
		return ErrNATGatewayProductNameRequired
	}
	if locationID < 1 {
		return ErrNATGatewayLocationIDRequired
	}
	if speed < 1 {
		return ErrNATGatewaySpeedRequired
	}
	if !slices.Contains(VALID_CONTRACT_TERMS, term) {
		return ErrNATGatewayInvalidTerm
	}
	return nil
}

// validateCreateNATGatewayRequest validates the request parameters for creating a NAT Gateway.
func validateCreateNATGatewayRequest(req *CreateNATGatewayRequest) error {
	if req == nil {
		return ErrNATGatewayRequestNil
	}
	return validateNATGatewayCommonFields(req.ProductName, req.LocationID, req.Speed, req.Term)
}

// validateUpdateNATGatewayRequest validates the request parameters for updating a NAT Gateway.
func validateUpdateNATGatewayRequest(req *UpdateNATGatewayRequest) error {
	if req == nil {
		return ErrNATGatewayRequestNil
	}
	if req.ProductUID == "" {
		return ErrNATGatewayProductUIDRequired
	}
	return validateNATGatewayCommonFields(req.ProductName, req.LocationID, req.Speed, req.Term)
}

// validateGetNATGatewayTelemetryRequest validates the request parameters.
func validateGetNATGatewayTelemetryRequest(req *GetNATGatewayTelemetryRequest) error {
	if req == nil {
		return ErrNATGatewayRequestNil
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

// ValidateNATGatewaySpeedSession checks the requested speed and session
// count against the live availability matrix. The matrix is fetched lazily
// on first call and cached on the service.
func (svc *NATGatewayServiceOp) ValidateNATGatewaySpeedSession(ctx context.Context, speed, sessionCount int) error {
	matrix, err := svc.ensureSessionMatrixCache().GetOrFetch(ctx)
	if err != nil {
		return err
	}
	supportedSpeeds := make([]int, 0, len(matrix))
	for _, entry := range matrix {
		if entry == nil {
			continue
		}
		supportedSpeeds = append(supportedSpeeds, entry.SpeedMbps)
		if entry.SpeedMbps != speed {
			continue
		}
		if slices.Contains(entry.SessionCount, sessionCount) {
			return nil
		}
		return fmt.Errorf("%w: got %d for %d Mbps; supported session counts: %v", ErrNATGatewaySessionCountNotSupported, sessionCount, speed, entry.SessionCount)
	}
	return fmt.Errorf("%w: got %d Mbps; supported speeds: %v", ErrNATGatewaySpeedNotSupported, speed, supportedSpeeds)
}
