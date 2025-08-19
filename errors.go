package megaport

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string `json:"message"`

	// Error Data
	Data string `json:"data"`

	// Trace ID returned from the API.
	TraceID string `json:"trace_id"`
}

// Error returns the string representation of the error
func (r *ErrorResponse) Error() string {
	if r.TraceID != "" {
		return fmt.Sprintf("%v %v: %d (trace_id %q) %s %s",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.TraceID, r.Message, r.Data)
	}
	return fmt.Sprintf("%v %v: %d %s %s",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message, r.Data)
}

// ErrWrongProductModify is returned when a user attempts to modify a product that can't be modified
var ErrWrongProductModify = errors.New("you can only update Ports, MCR, and MVE using this method")

// ErrInvalidTerm creates an error indicating an invalid contract term, dynamically listing the valid terms.
var ErrInvalidTerm = fmt.Errorf("invalid term, valid terms are %s months", intSliceToString(VALID_CONTRACT_TERMS))

// ErrMCRInvalidPortSpeed creates an error indicating an invalid MCR port speed, dynamically listing the valid speeds.
var ErrMCRInvalidPortSpeed = fmt.Errorf("invalid mcr port speed, valid speeds are %s", intSliceToString(VALID_MCR_PORT_SPEEDS))

// intSliceToString converts a slice of integers to a comma-separated string.
func intSliceToString(slice []int) string {
	strSlice := make([]string, len(slice))
	for i, v := range slice {
		strSlice[i] = strconv.Itoa(v)
	}
	return strings.Join(strSlice, ", ")
}

// ErrPortAlreadyLocked is returned when a port is already locked
var ErrPortAlreadyLocked = errors.New("that port is already locked, cannot lock")

// ErrPortNotLocked is returned when a port is not locked
var ErrPortNotLocked = errors.New("that port not locked, cannot unlock")

// ErrLocationNotFound is returned when a location can't be found
var ErrLocationNotFound = errors.New("unable to find location")

// ErrNoMatchingLocations is returned when a fuzzy search for a location doesn't return any results
var ErrNoMatchingLocations = errors.New("could not find any matching locations from search")

// ErrNoPartnerPortsFound is returned when no partner ports could be found matching the filters provided
var ErrNoPartnerPortsFound = errors.New("sorry there were no results returned based on the given filters")

// ErrNoAvailableVxcPorts is returned when there are no available ports for a user to connect to
var ErrNoAvailableVxcPorts = errors.New("there are no available ports for you to connect to")

// ErrCostCentreTooLong is returned when a cost centre is longer than 255 characters
var ErrCostCentreTooLong = errors.New("cost centre must be less than 255 characters")

// ErrManagedAccountNotFound is returned when a managed account can't be found
var ErrManagedAccountNotFound = errors.New("managed account not found")

// ErrInvalidVLAN is returned when a VLAN is invalid
var ErrInvalidVLAN = errors.New("invalid VLAN, must be between 0 and 4094")

// ErrInvalidVXCAEndPartnerConfig is returned when an invalid VXC A-End partner config is provided
var ErrInvalidVXCAEndPartnerConfig = errors.New("invalid vxc a-end partner config")

// ErrInvalidVXCBEndPartnerConfig is returned when an invalid VXC B-End partner config is provided
var ErrInvalidVXCBEndPartnerConfig = errors.New("invalid vxc b-end partner config")

// maintenanceStatesToString converts a slice of MaintenanceState to a slice of string
func maintenanceStatesToString(states []MaintenanceState) []string {
	strs := make([]string, len(states))
	for i, v := range states {
		strs[i] = string(v)
	}
	return strs
}

var ErrInvalidMaintenanceState = fmt.Errorf("invalid maintenance state, valid states are %s", strings.Join(maintenanceStatesToString(VALID_MAINTENANCE_STATES), ", "))

// outageStatesToString converts a slice of OutageState to a slice of string
func outageStatesToString(states []OutageState) []string {
	strs := make([]string, len(states))
	for i, v := range states {
		strs[i] = string(v)
	}
	return strs
}

var ErrInvalidOutageState = fmt.Errorf("invalid outage state, valid states are %s", strings.Join(outageStatesToString(VALID_OUTAGE_STATES), ", "))
