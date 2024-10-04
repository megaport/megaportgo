package megaport

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type Environment string

const (
	EnvironmentStaging    Environment = "https://api-staging.megaport.com/"
	EnvironmentProduction Environment = "https://api.megaport.com/"
)

const (
	libraryVersion = "1.0"
	defaultBaseURL = EnvironmentStaging
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

	// The access key for the API token
	AccessKey string

	// The secret key for the API token
	SecretKey string

	// Services used for communicating with the Megaport API

	// PortService provides methods for interacting with the Ports API
	PortService PortService
	// PartnerService provides methods for interacting with the Partners API
	PartnerService PartnerService
	// ProductService provides methods for interacting with the Products API
	ProductService ProductService
	// LocationService provides methods for interacting with the Locations API
	LocationService LocationService
	// VXCService provides methods for interacting with the VXCs API
	VXCService VXCService
	// MCRService provides methods for interacting with the MCRs API
	MCRService MCRService
	// MVEService provides methods for interacting with the MVEs API
	MVEService MVEService
	// ServiceKeyService provides methods for interacting with the Service Keys API
	ServiceKeyService ServiceKeyService
	// ManagedAccountService provides methods for interacting with the Managed Accounts API
	ManagedAccountService ManagedAccountService

	accessToken string    // Access Token for client
	tokenExpiry time.Time // Token Expiration

	LogResponseBody bool // Log Response Body of HTTP Requests
	LogCapture      *LogCapture

	// Optional function called after every successful request made to the API
	onRequestCompleted RequestCompletionCallback

	// Optional extra HTTP headers to set on every request to the API.
	headers map[string]string

	authMux sync.Mutex
}

// AccessTokenResponse is the response structure for the Login method containing the access token and expiration time.
type AccessTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error"`
}

// LogCapture is a simple struct to capture logs
type LogCapture struct {
	Buffer bytes.Buffer
}

// Write writes to the buffer
func (lc *LogCapture) Write(p []byte) (n int, err error) {
	return lc.Buffer.Write(p)
}

// String returns the buffer as a string
func (lc *LogCapture) String() string {
	return lc.Buffer.String()
}

type LevelFilterHandler struct {
	level   slog.Level
	handler slog.Handler
}

func NewLevelFilterHandler(level slog.Level, handler slog.Handler) *LevelFilterHandler {
	return &LevelFilterHandler{level: level, handler: handler}
}

func (h *LevelFilterHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level >= h.level {
		return h.handler.Handle(ctx, r)
	}
	return nil
}

func (h *LevelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LevelFilterHandler{
		level:   h.level,
		handler: h.handler.WithAttrs(attrs),
	}
}

func (h *LevelFilterHandler) WithGroup(name string) slog.Handler {
	return &LevelFilterHandler{
		level:   h.level,
		handler: h.handler.WithGroup(name),
	}
}

func (h *LevelFilterHandler) Enabled(context context.Context, level slog.Level) bool {
	return level >= h.level
}

// RequestCompletionCallback defines the type of the request callback function
type RequestCompletionCallback func(*http.Request, *http.Response)

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
		baseURL, _ = url.Parse(string(defaultBaseURL))
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	c := &Client{
		HTTPClient: httpClient,
		BaseURL:    baseURL,
		UserAgent:  userAgent,
		Logger:     logger,
		LogCapture: &LogCapture{},
	}

	c.ProductService = NewProductService(c)
	c.PortService = NewPortService(c)
	c.LocationService = NewLocationService(c)
	c.MCRService = NewMCRService(c)
	c.MVEService = NewMVEService(c)
	c.VXCService = NewVXCService(c)
	c.PartnerService = NewPartnerService(c)
	c.ServiceKeyService = NewServiceKeyService(c)
	c.ManagedAccountService = NewManagedAccountService(c)

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

// WithBaseURL is a client option for setting the base URL.
func WithBaseURL(bu string) ClientOpt {
	return func(c *Client) error {
		u, err := url.Parse(bu)
		if err != nil {
			return err
		}

		c.BaseURL = u
		return nil
	}
}

// WithLogHandler is an option to pass in a custom slog handler
func WithLogHandler(h slog.Handler) ClientOpt {
	return func(c *Client) error {
		c.Logger = slog.New(h)
		return nil
	}
}

// WithUserAgent is a client option for setting the user agent.
func WithUserAgent(ua string) ClientOpt {
	return func(c *Client) error {
		c.UserAgent = fmt.Sprintf("%s %s", ua, c.UserAgent)
		return nil
	}
}

// WithCustomHeaders sets optional HTTP headers on the client that are
// sent on each HTTP request.
func WithCustomHeaders(headers map[string]string) ClientOpt {
	return func(c *Client) error {
		for k, v := range headers {
			c.headers[k] = v
		}
		return nil
	}
}

// WithCredentials sets the client's API credentials
func WithCredentials(accessKey, secretKey string) ClientOpt {
	return func(c *Client) error {
		c.AccessKey = accessKey
		c.SecretKey = secretKey
		return nil
	}
}

// WithEnvironment is a helper for setting a BaseURL by environment
func WithEnvironment(e Environment) ClientOpt {
	return func(c *Client) error {
		u, err := url.Parse(string(e))
		if err != nil {
			return err
		}

		c.BaseURL = u
		return nil
	}
}

func WithLogLevel(level slog.Level) ClientOpt {
	return func(c *Client) error {
		handler := NewLevelFilterHandler(level, slog.NewJSONHandler(io.MultiWriter(os.Stdout, c.LogCapture), nil))
		c.Logger = slog.New(handler)
		return nil
	}
}

// WithLogResponseBody is a client option for setting the log response body flag
func WithLogResponseBody() ClientOpt {
	return func(c *Client) error {
		// for debugging - capture logs
		logCapture := &LogCapture{}
		multiWriter := io.MultiWriter(os.Stdout, logCapture)
		logger := slog.New(slog.NewJSONHandler(multiWriter, nil))
		c.Logger = logger
		c.LogCapture = logCapture
		c.LogResponseBody = true
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

	c.authMux.Lock()
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	c.authMux.Unlock()

	return req, nil
}

// SetOnRequestCompleted sets the Megaport API request completion callback
func (c *Client) SetOnRequestCompleted(rc RequestCompletionCallback) {
	c.onRequestCompleted = rc
}

// Do sends an API request and returns the API response. The API response is JSON decoded and stored in the value
// pointed to by v, or returned as an error if an API error has occurred. If v implements the io.Writer interface,
// the raw response will be written to v, without attempting to decode it.
func (c *Client) Do(ctx context.Context, req *http.Request, v interface{}) (*http.Response, error) {
	reqStart := time.Now()
	resp, err := DoRequestWithClient(ctx, c.HTTPClient, req)
	if err != nil {
		return nil, err
	}
	if c.onRequestCompleted != nil {
		c.onRequestCompleted(req, resp)
	}
	reqTime := time.Since(reqStart)

	respBody := resp.Body

	attrs := []slog.Attr{slog.Duration("duration", reqTime),
		slog.Int("status_code", resp.StatusCode),
		slog.String("path", req.URL.EscapedPath()),
		slog.String("api_host", c.BaseURL.Host),
		slog.String("method", req.Method),
		slog.String("trace_id", resp.Header.Get(headerTraceId))}

	if c.LogResponseBody {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		// Base64 encode the response body
		encodedBody := base64.StdEncoding.EncodeToString(b)

		// Create new reader for the later code
		respBody = io.NopCloser(bytes.NewReader(b))

		attrs = append(attrs, slog.String("response_body_base_64", encodedBody))
	}

	c.Logger.DebugContext(ctx, "completed apirequest", slog.Any("api_request", attrs))

	err = CheckResponse(resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusNoContent && v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, respBody)
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

type AuthInfo struct {
	Expiration  time.Time
	AccessToken string
}

// Authorize performs an OAuth-style login using the client's AccessKey and SecretKey and updates the client's access token on a successful response.
func (c *Client) Authorize(ctx context.Context) (*AuthInfo, error) {
	c.Logger.DebugContext(ctx, "authorizing client using access key and secret key", slog.String("access_key", c.AccessKey))

	// Shortcut if we've already authenticated.
	if time.Now().Before(c.tokenExpiry) {
		return &AuthInfo{Expiration: c.tokenExpiry, AccessToken: c.accessToken}, nil
	}

	if c.AccessKey == "" {
		return nil, errors.New("client has no AccessKey configured")
	}

	if c.SecretKey == "" {
		return nil, errors.New("client has no SecretKey configured")
	}

	// Encode the client ID and client secret to create Basic Authentication
	authHeader := base64.StdEncoding.EncodeToString([]byte(c.AccessKey + ":" + c.SecretKey))

	// Set the URL for the token endpoint
	var tokenURL string

	if c.BaseURL.Host == "api.megaport.com" {
		tokenURL = "https://auth-m2m.megaport.com/oauth2/token"
	} else if c.BaseURL.Host == "api-staging.megaport.com" {
		tokenURL = "https://oauth-m2m-staging.auth.ap-southeast-2.amazoncognito.com/oauth2/token"
	} else {
		return nil, errors.New("unknown API environment")
	}

	// Create form data for the request body
	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	// Create an HTTP request
	clientReq, err := c.NewRequest(ctx, "POST", tokenURL, nil)
	if err != nil {
		return nil, err
	}

	clientReq.URL.RawQuery = data.Encode()

	// Set the request headers
	clientReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	clientReq.Header.Set("Authorization", "Basic "+authHeader)

	// Create an HTTP client and send the request
	c.Logger.DebugContext(ctx, "login request", slog.String("token_url", tokenURL), slog.String("authorization_header", clientReq.Header.Get("Authorization")), slog.String("content_type", clientReq.Header.Get("Content_Type")))
	resp, err := c.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse the response JSON to extract the access token and expiration time
	authResponse := AccessTokenResponse{}
	if err := json.Unmarshal(body, &authResponse); err != nil {
		return nil, err
	}

	if authResponse.Error != "" {
		return nil, errors.New("authentication error: " + authResponse.Error)
	}

	c.authMux.Lock()
	// Store the access token and expiration in the client
	c.tokenExpiry = time.Now().Add(time.Duration(authResponse.ExpiresIn) * time.Second)
	c.accessToken = authResponse.AccessToken
	c.authMux.Unlock()

	c.Logger.DebugContext(ctx, "successful login")

	return &AuthInfo{Expiration: c.tokenExpiry, AccessToken: c.accessToken}, nil
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
