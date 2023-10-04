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
	Client       httpClientInterface
}

// http interface to cover mocking
type httpClientInterface interface {
	Do(req *http.Request) (retres *http.Response, reterr error)
}

// Wrap http.RoundTripper to append the user-agent header.
type roundTripper struct {
	T http.RoundTripper
}

func (t *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Go-Megaport-Library/0.1")

	return t.T.RoundTrip(req)
}

func NewHttpClient() *http.Client {
	return &http.Client{Transport: &roundTripper{http.DefaultTransport}}
}

// MakeAPICall
func (c *Config) MakeAPICall(verb string, endpoint string, body []byte) (*http.Response, error) {
	var request *http.Request
	var reqErr error

	url := c.Endpoint + endpoint

	c.Log.Debugln("Making call to: ", string(url))

	if body == nil {
		request, reqErr = http.NewRequest(verb, url, nil)
	} else {
		request, reqErr = http.NewRequest(verb, url, bytes.NewBuffer(body))
	}

	if reqErr != nil {
		return nil, reqErr
	}

	// Set the bearer token in the request header
	request.Header.Set("Authorization", "Bearer "+c.SessionToken)

	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	// check config for http client, create a new client if one was not provided at instantiation
	if c.Client == nil {
		c.Client = NewHttpClient()

	}

	response, resErr := c.Client.Do(request)

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

func (c *Config) GetProductType(productId string) (string, error) {
	url := "/v2/product/" + productId
	verb := "GET"

	detailsResponse, err := c.MakeAPICall(verb, url, nil)
	defer detailsResponse.Body.Close()

	isResErr, compiledResErr := c.IsErrorResponse(detailsResponse, &err, 200)
	if isResErr {
		return "", compiledResErr
	}

	body, fileErr := ioutil.ReadAll(detailsResponse.Body)

	if fileErr != nil {
		return "", fileErr
	}

	obj := map[string]interface{}{}
	if err := json.Unmarshal([]byte(body), &obj); err != nil {
		c.Log.Error(err)
	}

	return obj["data"].(map[string]interface{})["productType"].(string), nil
}
