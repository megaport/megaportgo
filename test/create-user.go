// usr/bin/env go run $0 "$@"; exit $?
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/megaport/megaportgo/config"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/service/authentication"
	"github.com/megaport/megaportgo/types"
)

const (
	ENDPOINTURL    = "https://api.staging.megaport.com/"
	CHARSET        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CREDENTIALFILE = ".mpt_test_credentials"
)

func main() {
	userPostfix := generateRandomStringWithCharset(10, CHARSET)
	password := generateRandomStringWithCharset(20, CHARSET+"@+=")
	username := "golib+" + userPostfix + "@sink.megaport.com"

	fmt.Println("Registering User: ", username)
	if createErr := createUser(username, password); createErr != nil {
		fmt.Println("Failed to create User", createErr)
		os.Exit(1)
	}
	fmt.Println("User has been register successfully")

	logger := config.NewDefaultLogger()
	logger.SetLevel(config.Off)

	client := config.NewHttpClient()

	cfg := config.Config{
		Log:      logger,
		Endpoint: ENDPOINTURL,
		Client:   client,
	}

	auth := authentication.New(&cfg)

	fmt.Println("Establishing Session for user")
	session, err := auth.LoginOauth(username, password)
	if err != nil {
		fmt.Println("Unable to establish session for user: ", err)
		os.Exit(1)
	}
	fmt.Println("Session Established")

	cfg.SessionToken = session

	fmt.Println("Setting up mock Market information for user")
	userConfErr := createMarket(username, cfg)
	if userConfErr != nil {
		fmt.Println("Setup failed", userConfErr)
		os.Exit(1)
	}
	fmt.Println("Mock Market information set for user")

	generateEnvironmentVaribles(username, password)
	fmt.Printf("User credentails can be found in %s. Source file to set environment vars for user\n\n", CREDENTIALFILE)
	fmt.Printf("\tsource %s\n\n", CREDENTIALFILE)
}

func createUser(username string, password string) error {
	createUserUrl := ENDPOINTURL + "/v2/social/registration"

	data := url.Values{}
	client := config.NewHttpClient()

	data.Add("firstName", "Go")
	data.Add("lastName", "Testing")
	data.Add("email", username)
	data.Add("password", password)
	data.Add("companyName", "Go Testing Company")

	loginRequest, _ := http.NewRequest("POST", createUserUrl, strings.NewReader(data.Encode()))
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRequest.Header.Set("Accept", "application/json")

	response, resErr := client.Do(loginRequest)

	if resErr != nil {
		return resErr
	}

	defer response.Body.Close()

	isError, compiledError := isErrorResponse(response, &resErr, 201)

	if isError {
		return compiledError
	}

	return nil
}

// IsErrorResponse returns an error report if an error response is detected from the API.
func isErrorResponse(response *http.Response, responseErr *error, expectedReturnCode int) (bool, error) {
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

func generateRandomStringWithCharset(length int, charset string) string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateEnvironmentVaribles(username string, password string) {
	file, err := os.Create(CREDENTIALFILE)

	if err != nil {
		return
	}
	defer file.Close()

	usrStr := fmt.Sprintf("export MEGAPORT_USERNAME=\"%s\"\n", username)
	pwdStr := fmt.Sprintf("export MEGAPORT_PASSWORD=\"%s\"\n", password)

	_, err = file.WriteString(usrStr)
	if err != nil {
		return
	}
	_, err = file.WriteString(pwdStr)
	if err != nil {
		return
	}
}

func createMarket(contactEmail string, cfg config.Config) error {
	market := types.Market{
		Currency:               "AUD",
		Language:               "en",
		CompanyLegalIdentifier: "ABN987654",
		CompanyLegalName:       "Go Testing Company",
		BillingContactName:     "Go Testing",
		BillingContactPhone:    "0730000000",
		BillingContactEmail:    contactEmail,
		AddressLine1:           "Level 3, 825 Ann St,  QLD 4006",
		City:                   "Fortitude Valley",
		State:                  "QLD",
		Postcode:               "4006",
		Country:                "AU",
		FirstPartyID:           808,
	}

	marketJSON, marketMarshalErr := json.Marshal(market)

	if marketMarshalErr != nil {
		return marketMarshalErr
	}

	marketResponse, marketErr := cfg.MakeAPICall("POST", "/v2/market/", marketJSON)

	isMarketError, parsedMarketErr := isErrorResponse(marketResponse, &marketErr, 201)
	if isMarketError {
		return parsedMarketErr
	}
	defer marketResponse.Body.Close()

	return nil
}
