package megaport

import (
	"errors"
	"fmt"
	"net/http"
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

// ErrInvalidTerm is returned for an invalid product term
var ErrInvalidTerm = errors.New("invalid term, valid values are 1, 12, 24, and 36")

// ErrPortAlreadyLocked is returned when a port is already locked
var ErrPortAlreadyLocked = errors.New("that port is already locked, cannot lock")

// ErrPortNotLocked is returned when a port is not locked
var ErrPortNotLocked = errors.New("that port not locked, cannot unlock")

// ErrMCRInvalidPortSpeed is returned for an invalid MCR port speed
var ErrMCRInvalidPortSpeed = errors.New("invalid port speed, valid speeds are 1000, 2500, 5000, and 10000")

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
