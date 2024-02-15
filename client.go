package megaport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	libraryVersion = "1.0"
	defaultBaseURL = "https://api-staging.megaport.com/"
	userAgent      = "Go-Megaport-Library/" + libraryVersion
	mediaType      = "application/json"
	headerTraceId  = "Trace-Id"
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
	// Token Expiration
	TokenExpiry time.Time

	// Optional function called after every successful request made to the API
	onRequestCompleted RequestCompletionCallback

	// Services used for communicating with the Megaport API
	AuthenticationService AuthenticationService
	PortService           PortService
	PartnerService        PartnerService
	ProductService        ProductService
	LocationService       LocationService
	VXCService            VXCService
	MCRService            MCRService
	MVEService            MVEService

	// Optional extra HTTP headers to set on every request to the API.
	headers map[string]string
}

// RequestCompletionCallback defines the type of the request callback function
type RequestCompletionCallback func(*http.Request, *http.Response)

// NewFromToken returns a new Megaport API client with the given API
// token.
func NewFromToken(token string) *Client {
	cleanToken := strings.Trim(strings.TrimSpace(token), "'")
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cleanToken})

	oauthClient := oauth2.NewClient(ctx, ts)
	client, err := New(oauthClient)
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

	c.AuthenticationService = NewAuthenticationService(c)
	c.ProductService = NewProductService(c)
	c.PortService = NewPortService(c)
	c.LocationService = NewLocationService(c)
	c.MCRService = NewMCRService(c)
	c.MVEService = NewMVEService(c)
	c.VXCService = NewVXCService(c)
	c.PartnerService = NewPartnerService(c)

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
		req, err = http.NewRequestWithContext(ctx, method, u.String(), nil)
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

		req, err = http.NewRequestWithContext(ctx, method, u.String(), buf)
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
		return nil, err
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

	return resp, nil
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

	if errorResponse.TraceID == "" {
		errorResponse.TraceID = r.Header.Get(headerTraceId)
	}

	return errorResponse
}

// PtrTo returns a pointer to the provided input.
func PtrTo[T any](v T) *T {
	return &v
}
