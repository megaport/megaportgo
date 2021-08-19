package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/shared"
	"github.com/megaport/megaportgo/types"
)

type Config struct {
	Log          Logger
	Endpoint     string
	SessionToken string
}

// MakeAPICall
func (c *Config) MakeAPICall(verb string, endpoint string, body []byte) (*http.Response, error) {
	client := &http.Client{}
	var request *http.Request
	var reqErr error

	url := c.Endpoint + endpoint

	if body == nil {
		request, reqErr = http.NewRequest(verb, url, nil)
	} else {
		request, reqErr = http.NewRequest(verb, url, bytes.NewBuffer(body))
	}

	if reqErr != nil {
		return nil, reqErr
	}

	request.Header.Set("X-Auth-Token", c.SessionToken)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Go-Megaport-Library/0.1")

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, resErr := client.Do(request)

	if resErr != nil {
		return nil, resErr
	} else {
		return response, nil
	}
}

// IsErrorResponse returns an error report if an error response is detected from the API.
func (c *Config) IsErrorResponse(response *http.Response, responseErr *error, expectedReturnCode int) (bool, error) {
	if *responseErr != nil {
		return true, *responseErr
	}

	if response.StatusCode != expectedReturnCode {
		body, fileErr := ioutil.ReadAll(response.Body)

		if fileErr != nil {
			return false, fileErr
		}

		errResponse := types.ErrorResponse{}
		errParseErr := json.Unmarshal([]byte(body), &errResponse)

		if errParseErr != nil {
			errorReport := fmt.Sprintf(mega_err.ERR_PARSING_ERR_RESPONSE, response.StatusCode, errParseErr.Error(), string(body))
			return true, errors.New(errorReport)
		}

		return true, errors.New(errResponse.Message + ": " + errResponse.Data)
	}

	return false, nil
}

// PurchaseError prints out details about a failed purchase.
func (c *Config) PurchaseError(productID string, err error) {
	if !shared.IsGuid(productID) {
		c.Log.Infoln("Returned product ID is empty.")
	}

	if err != nil {
		c.Log.Infoln("Error purchasing Product:", err)
	}
}
