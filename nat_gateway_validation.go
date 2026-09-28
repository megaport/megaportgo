package megaport

import (
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

// validateNATGatewayCommonFields performs the structural checks shared by
// CreateNATGatewayRequest and UpdateNATGatewayRequest.
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

// NATGatewaySpeedSessionResult describes whether a speed/session pair is
// orderable according to a NAT Gateway availability matrix.
type NATGatewaySpeedSessionResult struct {
	// Supported is true only if the speed exists AND the session count is
	// valid at that speed.
	Supported bool
	// SpeedSupported is true if the speed appears in the matrix at all.
	SpeedSupported bool
	// SupportedSpeeds lists every speed in the matrix. Always complete.
	SupportedSpeeds []int
	// SessionsAtSpeed lists the session counts valid at the requested speed
	// (nil if the speed is unsupported).
	SessionsAtSpeed []int
}

// NATGatewaySpeedSessionSupported reports whether a speed/sessionCount pair is
// present in a NAT Gateway availability matrix obtained from
// ListNATGatewaySessions. It performs no network I/O; the caller supplies the
// matrix and validation happens locally.
func NATGatewaySpeedSessionSupported(matrix []*NATGatewaySession, speed, sessionCount int) NATGatewaySpeedSessionResult {
	res := NATGatewaySpeedSessionResult{SupportedSpeeds: make([]int, 0, len(matrix))}
	for _, entry := range matrix {
		if entry == nil {
			continue
		}
		res.SupportedSpeeds = append(res.SupportedSpeeds, entry.SpeedMbps)
		if entry.SpeedMbps == speed {
			res.SpeedSupported = true
			res.SessionsAtSpeed = entry.SessionCount
		}
	}
	res.Supported = res.SpeedSupported && slices.Contains(res.SessionsAtSpeed, sessionCount)
	return res
}
