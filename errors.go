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

// ErrVnicsOnNonMVE is returned when vNIC updates are supplied for a non-MVE product
var ErrVnicsOnNonMVE = errors.New("vNICs can only be modified on MVE products")

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

// ErrInvalidAddOnType is returned when an invalid add-on type is provided
var ErrInvalidAddOnType = errors.New("invalid add-on type, currently only IP_SEC is supported")

// ErrInvalidIPsecTunnelCount is returned when the IPsec tunnel count is not valid
var ErrInvalidIPsecTunnelCount = errors.New("invalid IPsec tunnel count, valid values are 10, 20, or 30 (0 defaults to 10)")

// ErrInvalidMonth is returned when RTT statistics are requested for an invalid month
var ErrInvalidMonth = errors.New("invalid month, must be between 1 and 12")

// ErrInvalidYear is returned when RTT statistics are requested for an invalid year
var ErrInvalidYear = errors.New("invalid year, must be between 0 and 99")

// ErrMCRCancelLaterNotAllowed is returned when attempting to schedule MCR deletion for later (only CANCEL_NOW is allowed)
var ErrMCRCancelLaterNotAllowed = errors.New("mcr products do not support scheduled deletion (cancel later), only immediate deletion (CANCEL_NOW) is allowed")

// ErrPortCancelLaterNotAllowed is returned when attempting to schedule Port deletion for later (only CANCEL_NOW is allowed)
var ErrPortCancelLaterNotAllowed = errors.New("port products do not support scheduled deletion (cancel later), only immediate deletion (CANCEL_NOW) is allowed")

// ErrMCRNotFound is returned when an MCR cannot be found (deleted or never existed).
var ErrMCRNotFound = errors.New("mcr not found or deleted")

// ErrMCRDecommissioned is returned when an MCR has been decommissioned.
var ErrMCRDecommissioned = errors.New("mcr has been decommissioned")

// IsServiceNotFoundError reports whether err is a Megaport API not-found response —
// either HTTP 404 or the non-standard HTTP 400 "Could not find a service with UID" form.
func IsServiceNotFoundError(err error) bool {
	apiErr, ok := err.(*ErrorResponse)
	if !ok || apiErr.Response == nil {
		return false
	}
	if apiErr.Response.StatusCode == http.StatusNotFound {
		return true
	}
	return apiErr.Response.StatusCode == http.StatusBadRequest &&
		strings.Contains(apiErr.Message, "Could not find a service with UID")
}

// ErrTransitVXCCancelLaterNotAllowed is returned when attempting to schedule Transit VXC deletion for later (only CANCEL_NOW is allowed)
var ErrTransitVXCCancelLaterNotAllowed = errors.New("transit vxc (megaport internet) does not support scheduled deletion (cancel later), only immediate deletion (CANCEL_NOW) is allowed")

// ErrDeleteVXCRequestNil is returned when DeleteVXC is called with a nil request.
var ErrDeleteVXCRequestNil = errors.New("delete VXC request cannot be nil")

// ErrDeleteMCRRequestNil is returned when DeleteMCR is called with a nil request.
var ErrDeleteMCRRequestNil = errors.New("delete MCR request cannot be nil")

// ErrDeletePortRequestNil is returned when DeletePort is called with a nil request.
var ErrDeletePortRequestNil = errors.New("delete port request cannot be nil")

// ErrBuyIXRequestNil is returned when BuyIX or ValidateIXOrder is called with a nil request.
var ErrBuyIXRequestNil = errors.New("buy IX request cannot be nil")

// ErrUpdateIXRequestNil is returned when UpdateIX is called with a nil request.
var ErrUpdateIXRequestNil = errors.New("update IX request cannot be nil")

// ErrDeleteIXRequestNil is returned when DeleteIX is called with a nil request.
var ErrDeleteIXRequestNil = errors.New("delete IX request cannot be nil")

// ErrBuyMCRRequestNil is returned when BuyMCR or ValidateMCROrder is called with a nil request.
var ErrBuyMCRRequestNil = errors.New("buy MCR request cannot be nil")

// ErrCreateMCRPrefixFilterListRequestNil is returned when CreatePrefixFilterList is called with a nil request.
var ErrCreateMCRPrefixFilterListRequestNil = errors.New("create MCR prefix filter list request cannot be nil")

// ErrModifyMCRRequestNil is returned when ModifyMCR is called with a nil request.
var ErrModifyMCRRequestNil = errors.New("modify MCR request cannot be nil")

// ErrBuyMVERequestNil is returned when BuyMVE or ValidateMVEOrder is called with a nil request.
var ErrBuyMVERequestNil = errors.New("buy MVE request cannot be nil")

// ErrModifyMVERequestNil is returned when ModifyMVE is called with a nil request.
var ErrModifyMVERequestNil = errors.New("modify MVE request cannot be nil")

// ErrDeleteMVERequestNil is returned when DeleteMVE is called with a nil request.
var ErrDeleteMVERequestNil = errors.New("delete MVE request cannot be nil")

// ErrBuyVXCRequestNil is returned when BuyVXC or ValidateVXCOrder is called with a nil request.
var ErrBuyVXCRequestNil = errors.New("buy VXC request cannot be nil")

// ErrUpdateVXCRequestNil is returned when UpdateVXC is called with a nil request.
var ErrUpdateVXCRequestNil = errors.New("update VXC request cannot be nil")

// ErrLookupPartnerPortsRequestNil is returned when LookupPartnerPorts is called with a nil request.
var ErrLookupPartnerPortsRequestNil = errors.New("lookup partner ports request cannot be nil")

// ErrListPartnerPortsRequestNil is returned when ListPartnerPorts is called with a nil request.
var ErrListPartnerPortsRequestNil = errors.New("list partner ports request cannot be nil")

// ErrBuyPortRequestNil is returned when BuyPort or ValidatePortOrder is called with a nil request.
var ErrBuyPortRequestNil = errors.New("buy port request cannot be nil")

// ErrModifyPortRequestNil is returned when ModifyPort is called with a nil request.
var ErrModifyPortRequestNil = errors.New("modify port request cannot be nil")

// ErrModifyProductRequestNil is returned when ModifyProduct is called with a nil request.
var ErrModifyProductRequestNil = errors.New("modify product request cannot be nil")

// ErrDeleteProductRequestNil is returned when DeleteProduct is called with a nil request.
var ErrDeleteProductRequestNil = errors.New("delete product request cannot be nil")

// ErrManageProductLockRequestNil is returned when ManageProductLock is called with a nil request.
var ErrManageProductLockRequestNil = errors.New("manage product lock request cannot be nil")

// ErrCreateServiceKeyRequestNil is returned when CreateServiceKey is called with a nil request.
var ErrCreateServiceKeyRequestNil = errors.New("create service key request cannot be nil")

// ErrUpdateServiceKeyRequestNil is returned when UpdateServiceKey is called with a nil request.
var ErrUpdateServiceKeyRequestNil = errors.New("update service key request cannot be nil")

// ErrCreateUserRequestNil is returned when CreateUser is called with a nil request.
var ErrCreateUserRequestNil = errors.New("create user request cannot be nil")

// ErrUpdateUserRequestNil is returned when UpdateUser is called with a nil request.
var ErrUpdateUserRequestNil = errors.New("update user request cannot be nil")

// maintenanceStatesToString converts a slice of MaintenanceState to a slice of strings
func maintenanceStatesToString(states []MaintenanceState) []string {
	strs := make([]string, len(states))
	for i, v := range states {
		strs[i] = string(v)
	}
	return strs
}

var ErrInvalidMaintenanceState = fmt.Errorf("invalid maintenance state, valid states are %s", strings.Join(maintenanceStatesToString(VALID_MAINTENANCE_STATES), ", "))

// outageStatesToString converts a slice of OutageState to a slice of strings
func outageStatesToString(states []OutageState) []string {
	strs := make([]string, len(states))
	for i, v := range states {
		strs[i] = string(v)
	}
	return strs
}

var ErrInvalidOutageState = fmt.Errorf("invalid outage state, valid states are %s", strings.Join(outageStatesToString(VALID_OUTAGE_STATES), ", "))
