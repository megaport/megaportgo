package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
	"golang.org/x/oauth2"
)

const (
	libraryVersion = "1.0"
	defaultBaseURL = "https://api-staging.megaport.com/"
	userAgent      = "Go-Megaport-Library/" + libraryVersion
	mediaType      = "application/json"

	defaultRetryMax     = 4
	defaultRetryWaitMax = 30
	defaultRetryWaitMin = 1

	headerRequestID             = "x-request-id"
	internalHeaderRetryAttempts = "X-Megaport-Retry-Attempts"
)

// Client manges communication with the Megaport API
type Client struct {
	// HTTP Client used to communicate with the Megaport API
	HTTPClient *http.Client

	// Logger for client
	Logger *slog.Logger

	// Base URL
	BaseURL *url.URL

	// User agent for client
	UserAgent string

	// Session Token for client
	SessionToken string

	// Optional function called after every successful request made to the DO APIs
	onRequestCompleted RequestCompletionCallback

	// Services used for communicating with the Megaport API
	AuthenticationService AuthenticationService
	PortService           PortService
	ProductService        ProductService
	LocationService       LocationService

	// Optional extra HTTP headers to set on every request to the API.
	headers map[string]string

	// Optional retry values. Setting the RetryConfig.RetryMax value enables automatically retrying requests
	// that fail with 429 or 500-level response codes using the go-retryablehttp client
	RetryConfig RetryConfig
}

// RetryConfig sets the values used for enabling retries and backoffs for
// requests that fail with 429 or 500-level response codes using the go-retryablehttp client.
// RetryConfig.RetryMax must be configured to enable this behavior. RetryConfig.RetryWaitMin and
// RetryConfig.RetryWaitMax are optional, with the default values being 1.0 and 30.0, respectively.
//
// You can use
//
//	megaport.PtrTo(1.0)
//
// to explicitly set the RetryWaitMin and RetryWaitMax values.
//
// Note: Opting to use the go-retryablehttp client will overwrite any custom HTTP client passed into New().
// Only the oauth2.TokenSource and Timeout will be maintained.
type RetryConfig struct {
	RetryMax     int
	RetryWaitMin *float64    // Minimum time to wait
	RetryWaitMax *float64    // Maximum time to wait
	Logger       interface{} // Customer logger instance. Must implement either go-retryablehttp.Logger or go-retryablehttp.LeveledLogger
}

// Wrap http.RoundTripper to append the user-agent header.
type roundTripper struct {
	T http.RoundTripper
}

func (t *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Go-Megaport-Library/0.2.2")

	return t.T.RoundTrip(req)
}

func NewHttpClient() *http.Client {
	return &http.Client{Transport: &roundTripper{http.DefaultTransport}}
}

// RequestCompletionCallback defines the type of the request callback function
type RequestCompletionCallback func(*http.Request, *http.Response)

// An ErrorResponse reports the error caused by an API request
type ErrorResponse struct {
	// HTTP response that caused this error
	Response *http.Response

	// Error message
	Message string `json:"message"`

	// RequestID returned from the API, useful to contact support.
	RequestID string `json:"request_id"`

	// Attempts is the number of times the request was attempted when retries are enabled.
	Attempts int
}

func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)

	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	origURL, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	origValues := origURL.Query()

	newValues, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	for k, v := range newValues {
		origValues[k] = v
	}

	origURL.RawQuery = origValues.Encode()
	return origURL.String(), nil
}

// NewFromToken returns a new Megaport API client with the given API
// token.
func NewFromToken(token string) *Client {
	cleanToken := strings.Trim(strings.TrimSpace(token), "'")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cleanToken})

	oauthClient := oauth2.NewClient(ctx, ts)
	client, err := New(oauthClient, WithRetryAndBackoffs(
		RetryConfig{
			RetryMax:     defaultRetryMax,
			RetryWaitMin: PtrTo(float64(defaultRetryWaitMin)),
			RetryWaitMax: PtrTo(float64(defaultRetryWaitMax)),
		},
	))
	if err != nil {
		panic(err)
	}

	return client
}

// NewClient returns a new Megaport API client, using the given
// http.Client to perform all requests.

func NewClient(httpClient *http.Client, base *url.URL) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	var baseURL *url.URL
	if base != nil {
		baseURL = base
	} else {
		baseURL, _ = url.Parse(defaultBaseURL)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	c := &Client{
		HTTPClient: httpClient,
		BaseURL:    baseURL,
		UserAgent:  userAgent,
		Logger:     logger,
	}

	c.AuthenticationService = NewAuthenticationServiceOp(c)
	c.ProductService = NewProductServiceOp(c)
	c.PortService = NewPortServiceOp(c)
	c.LocationService = NewLocationServiceOp(c)

	c.headers = make(map[string]string)

	return c
}

// ClientOpt are options for New.
type ClientOpt func(*Client) error

// New returns a new Megaport API client instance.
func New(httpClient *http.Client, opts ...ClientOpt) (*Client, error) {
	c := NewClient(httpClient, nil)
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// if retryMax is set it will use the retryablehttp client.
	if c.RetryConfig.RetryMax > 0 {
		retryableClient := retryablehttp.NewClient()
		retryableClient.RetryMax = c.RetryConfig.RetryMax

		if c.RetryConfig.RetryWaitMin != nil {
			retryableClient.RetryWaitMin = time.Duration(*c.RetryConfig.RetryWaitMin * float64(time.Second))
		}
		if c.RetryConfig.RetryWaitMax != nil {
			retryableClient.RetryWaitMax = time.Duration(*c.RetryConfig.RetryWaitMax * float64(time.Second))
		}

		// By default this is nil and does not log.
		retryableClient.Logger = c.RetryConfig.Logger

		// if timeout is set, it is maintained before overwriting client with StandardClient()
		retryableClient.HTTPClient.Timeout = c.HTTPClient.Timeout

		// This custom ErrorHandler is required to provide errors that are consistent
		// with a *megaport.ErrorResponse and a non-nil *megaport.Response while providing
		// insight into retries using an internal header.
		retryableClient.ErrorHandler = func(resp *http.Response, err error, numTries int) (*http.Response, error) {
			if resp != nil {
				resp.Header.Add(internalHeaderRetryAttempts, strconv.Itoa(numTries))

				return resp, err
			}

			return resp, err
		}

		retryableClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			// In addition to the default retry policy, we also retry HTTP/2 INTERNAL_ERROR errors.
			// See: https://github.com/golang/go/issues/51323
			if err != nil && strings.Contains(err.Error(), "INTERNAL_ERROR") && strings.Contains(reflect.TypeOf(err).String(), "http2") {
				return true, nil
			}

			return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
		}

		var source *oauth2.Transport
		if _, ok := c.HTTPClient.Transport.(*oauth2.Transport); ok {
			source = c.HTTPClient.Transport.(*oauth2.Transport)
		}
		c.HTTPClient = retryableClient.StandardClient()
		c.HTTPClient.Transport = &oauth2.Transport{
			Base:   c.HTTPClient.Transport,
			Source: source.Source,
		}

	}

	return c, nil
}

// SetBaseURL is a client option for setting the base URL.
func SetBaseURL(bu string) ClientOpt {
	return func(c *Client) error {
		u, err := url.Parse(bu)
		if err != nil {
			return err
		}

		c.BaseURL = u
		return nil
	}
}

// SetLogHandler is an option to pass in a custom slog handler
func SetLogHandler(h slog.Handler) ClientOpt {
	return func(c *Client) error {
		c.Logger = slog.New(h)
		return nil
	}
}

// SetUserAgent is a client option for setting the user agent.
func SetUserAgent(ua string) ClientOpt {
	return func(c *Client) error {
		c.UserAgent = fmt.Sprintf("%s %s", ua, c.UserAgent)
		return nil
	}
}

// SetRequestHeaders sets optional HTTP headers on the client that are
// sent on each HTTP request.
func SetRequestHeaders(headers map[string]string) ClientOpt {
	return func(c *Client) error {
		for k, v := range headers {
			c.headers[k] = v
		}
		return nil
	}
}

// WithRetryAndBackoffs sets retry values. Setting the RetryConfig.RetryMax value enables automatically retrying requests
// that fail with 429 or 500-level response codes using the go-retryablehttp client
func WithRetryAndBackoffs(retryConfig RetryConfig) ClientOpt {
	return func(c *Client) error {
		c.RetryConfig.RetryMax = retryConfig.RetryMax
		c.RetryConfig.RetryWaitMax = retryConfig.RetryWaitMax
		c.RetryConfig.RetryWaitMin = retryConfig.RetryWaitMin
		c.RetryConfig.Logger = retryConfig.Logger
		return nil
	}
}

// NewRequest creates an API request. A relative URL can be provided in urlStr, which will be resolved to the
// BaseURL of the Client. Relative URLS should always be specified without a preceding slash. If specified, the
// value pointed to by body is JSON encoded and included in as the request body.
func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	u, err := c.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		req, err = http.NewRequest(method, u.String(), nil)
		if err != nil {
			return nil, err
		}

	default:
		buf := new(bytes.Buffer)
		if body != nil {
			err = json.NewEncoder(buf).Encode(body)
			if err != nil {
				return nil, err
			}
		}

		req, err = http.NewRequest(method, u.String(), buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", mediaType)
	}

	for k, v := range c.headers {
		req.Header.Add(k, v)
	}

	req.Header.Set("Accept", mediaType)
	req.Header.Set("User-Agent", c.UserAgent)

	if c.SessionToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.SessionToken)
	}

	return req, nil
}

// OnRequestCompleted sets the Megaport API request completion callback
func (c *Client) OnRequestCompleted(rc RequestCompletionCallback) {
	c.onRequestCompleted = rc
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := DoRequestWithClient(ctx, c.HTTPClient, req)
	if err != nil {
		return nil, err
	}
	if c.onRequestCompleted != nil {
		c.onRequestCompleted(req, resp)
	}

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if resp.StatusCode != http.StatusNoContent && v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				return nil, err
			}
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
			if err != nil {
				return nil, err
			}
		}
	}

	return resp, err
}

// DoRequest submits an HTTP request.
func DoRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return DoRequestWithClient(ctx, http.DefaultClient, req)
}

// DoRequestWithClient submits an HTTP request using the specified client.
func DoRequestWithClient(
	ctx context.Context,
	client *http.Client,
	req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return client.Do(req)
}

func (r *ErrorResponse) Error() string {
	var attempted string
	if r.Attempts > 0 {
		attempted = fmt.Sprintf("; giving up after %d attempt(s)", r.Attempts)
	}

	if r.RequestID != "" {
		return fmt.Sprintf("%v %v: %d (request %q) %v%s",
			r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.RequestID, r.Message, attempted)
	}
	return fmt.Sprintf("%v %v: %d %v%s",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode, r.Message, attempted)
}

// IsErrorResponse returns an error report if an error response is detected from the API.
func (c *Client) IsErrorResponse(response *http.Response, responseErr *error, expectedReturnCode int) (bool, error) {
	if *responseErr != nil {
		return true, *responseErr
	}

	if response.StatusCode != expectedReturnCode {
		body, fileErr := io.ReadAll(response.Body)

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

// CheckResponse checks the API response for errors, and returns them if present. A response is considered an
// error if it has a status code outside the 200 range. API error responses are expected to have either no response
// body, or a JSON response body that maps to ErrorResponse. Any other response body will be silently ignored.
// If the API error response does not include the request ID in its body, the one from its header will be used.
func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			errorResponse.Message = string(data)
		}
	}

	if errorResponse.RequestID == "" {
		errorResponse.RequestID = r.Header.Get(headerRequestID)
	}

	attempts, strconvErr := strconv.Atoi(r.Header.Get(internalHeaderRetryAttempts))
	if strconvErr == nil {
		errorResponse.Attempts = attempts
	}

	return errorResponse
}

// PtrTo returns a pointer to the provided input.
func PtrTo[T any](v T) *T {
	return &v
}
