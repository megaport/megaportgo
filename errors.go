package megaport

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string `json:"message"`

	// RequestID returned from the API, useful to contact support.
	RequestID string `json:"request_id"`
}

// IsErrorResponse returns an error report if an error response is detected from the API.
func (c *Client) IsErrorResponse(response *http.Response, responseErr *error, expectedReturnCode int) (bool, error) {
	if *responseErr != nil {
		return true, *responseErr
	}

	if response.StatusCode != expectedReturnCode {
		errorResponse := &ErrorResponse{Response: response}
		data, err := io.ReadAll(response.Body)
		if err == nil && len(data) > 0 {
			err := json.Unmarshal(data, errorResponse)
			if err != nil {
				errorResponse.Message = string(data)
			}
		}

		if errorResponse.RequestID == "" {
			errorResponse.RequestID = response.Header.Get(headerRequestID)
		}

		return true, errorResponse
	}

	return false, nil
}

// ArgError is an error that represents an error with an input to godo. It
// identifies the argument and the cause (if possible).
type ArgError struct {
	arg    string
	reason string
}

var _ error = &ArgError{}

// NewArgError creates an InputError.
func NewArgError(arg, reason string) *ArgError {
	return &ArgError{
		arg:    arg,
		reason: reason,
	}
}

func (e *ArgError) Error() string {
	return fmt.Sprintf("%s is invalid because %s", e.arg, e.reason)
}

func (r *ErrorResponse) Error() string {
	if r.RequestID != "" {
		return fmt.Sprintf("%v %v: %d (request %q) %s",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.RequestID, r.Message)
	}
	return fmt.Sprintf("%v %v: %d %s",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message)
}
