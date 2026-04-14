package megaport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	ctx = context.TODO()
)

// ClientTestSuite tests the Megaport SDK Client.
type ClientTestSuite struct {
	suite.Suite
	client *Client
	server *httptest.Server
	mux    *http.ServeMux
}

func TestClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *ClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// testURLParseError tests if the error is a URL parse error.
func (suite *ClientTestSuite) testURLParseError(err error) {
	if err == nil {
		suite.FailNow("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		suite.FailNowf("Expected URL parse error, got %+v", err.Error())
	}
}

// testClientDefaultBaseURL tests if the client's default base URL is set to the defaultBaseURL.
func (suite *ClientTestSuite) testClientDefaultBaseURL(c *Client) {
	if c.BaseURL == nil || c.BaseURL.String() != string(defaultBaseURL) {
		suite.FailNowf("NewClient BaseURL = %v, expected %v", c.BaseURL.String(), defaultBaseURL)
	}
}

// testClientDefaultUserAgent tests if the client's default user agent is set to the userAgent.
func (suite *ClientTestSuite) testClientDefaultUserAgent(c *Client) {
	if c.UserAgent != userAgent {
		suite.FailNowf("NewClient UserAgent = %v, expected %v", c.UserAgent, userAgent)
	}
}

// testClientDefaults tests if the client's default base URL and user agent are set to the defaultBaseURL and userAgent.
func (suite *ClientTestSuite) testClientDefaults(c *Client) {
	suite.testClientDefaultBaseURL(c)
	suite.testClientDefaultUserAgent(c)
}

// TestNewClient tests if the NewClient function returns a client with the default base URL and user agent.
func (suite *ClientTestSuite) TestNewClient() {
	c := NewClient(nil, nil)
	suite.testClientDefaults(c)
}

// TestNew tests if the New function returns a client with the default base URL and user agent.
func (suite *ClientTestSuite) TestNew() {
	c, err := New(nil)

	if err != nil {
		suite.FailNowf("New(): %v", err.Error())
	}
	suite.testClientDefaults(c)
}

// TestNewRequest_get tests if the NewRequest function returns a GET request with the default base URL and user agent.
func (suite *ClientTestSuite) TestNewRequest_get() {
	c := NewClient(nil, nil)

	inURL, outURL := "/foo", string(defaultBaseURL)+"foo"
	req, _ := c.NewRequest(ctx, http.MethodGet, inURL, nil)

	// test relative URL was expanded
	if req.URL.String() != outURL {
		suite.FailNowf("NewRequest(%v) URL = %v, expected %v", inURL, req.URL, outURL)
	}

	// test the content-type header is not set
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		suite.FailNowf("NewRequest() Content-Type = %v, expected empty string", contentType)
	}

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		suite.FailNowf("NewRequest() User-Agent = %v, expected %v", userAgent, c.UserAgent)
	}
}

// TestNewRequest_badURL tests if the NewRequest function returns a URL parse error.
func (suite *ClientTestSuite) TestNewRequest_badURL() {
	c := NewClient(nil, nil)
	_, err := c.NewRequest(ctx, http.MethodGet, ":", nil)
	suite.testURLParseError(err)
}

// TestNewRequest_withCustomUserAgent tests if the NewRequest function returns a request with a custom user agent.
func (suite *ClientTestSuite) TestNewRequest_withCustomUserAgent() {
	ua := "testing/0.0.1"
	c, err := New(nil, WithUserAgent(ua))

	if err != nil {
		suite.FailNowf("New() unexpected error: %v", err.Error())
	}

	req, _ := c.NewRequest(ctx, http.MethodGet, "/foo", nil)

	expected := fmt.Sprintf("%s %s", ua, userAgent)
	if got := req.Header.Get("User-Agent"); got != expected {
		suite.FailNowf("New() UserAgent = %s; expected %s", got, expected)
	}
}

// TestNewRequest_withResponseLogging tests if the NewRequest function returns a request with response logging.
func (suite *ClientTestSuite) TestNewRequest_withResponseLogging() {
	// for debugging - capture logs
	logCapture := &bytes.Buffer{}
	levelFilterHandler := NewLevelFilterHandler(slog.LevelDebug, slog.NewJSONHandler(io.Writer(logCapture), nil))

	c, err := New(nil, WithLogResponseBody(), WithLogHandler(levelFilterHandler))
	if err != nil {
		suite.FailNowf("unexpected error", "New() unexpected error: %v", err.Error())
	}
	suite.client = c
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url

	suite.mux.HandleFunc("/a", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			suite.FailNowf("Incorrect request method", "Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := suite.client.NewRequest(ctx, http.MethodGet, "/a", nil)
	_, err = suite.client.Do(ctx, req, nil)
	if err != nil {
		suite.FailNowf("Unexpected error: Do()", "Unexpected error: Do(): %v", err.Error())
	}

	// Check the log output for the expected base64 encoded response body
	expectedBase64 := "eyJBIjoiYSJ9" // base64 encoded {"A":"a"}
	logOutput := logCapture.String()
	if !strings.Contains(logOutput, expectedBase64) {
		suite.FailNowf("Log output does not contain expected base64", "Log output: %s", logOutput)
	}
}

// TestNewRequest_withCustomHeaders tests if the NewRequest function returns a request with custom headers.
func (suite *ClientTestSuite) TestNewRequest_withCustomHeaders() {
	expectedIdentity := "identity"
	expectedCustom := "x_test_header"

	c, err := New(nil, WithCustomHeaders(map[string]string{
		"Accept-Encoding": expectedIdentity,
		"X-Test-Header":   expectedCustom,
	}))
	if err != nil {
		suite.FailNowf("New() unexpected error: %v", err.Error())
	}

	req, _ := c.NewRequest(ctx, http.MethodGet, "/foo", nil)

	if got := req.Header.Get("Accept-Encoding"); got != expectedIdentity {
		suite.FailNowf("New() Custom Accept Encoding Header = %s; expected %s", got, expectedIdentity)
	}
	if got := req.Header.Get("X-Test-Header"); got != expectedCustom {
		suite.FailNowf("New() Custom Accept Encoding Header = %s; expected %s", got, expectedCustom)
	}
}

// TestDo_get tests if the Do function returns a GET request with the default base URL and user agent.
func (suite *ClientTestSuite) TestDo() {

	type foo struct {
		A string
	}

	suite.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			suite.FailNowf("Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := suite.client.NewRequest(ctx, http.MethodGet, "/", nil)
	body := new(foo)
	_, err := suite.client.Do(context.Background(), req, body)
	if err != nil {
		suite.FailNowf("", "Do(): %v", err.Error())
	}

	expected := &foo{"a"}
	if !reflect.DeepEqual(body, expected) {
		suite.FailNowf("", "Response body = %v, expected %v", *body, expected)
	}
}

// TestDo_httpError tests if the Do function returns an HTTP 400 error.
func (suite *ClientTestSuite) TestDo_httpError() {
	suite.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := suite.client.NewRequest(ctx, http.MethodGet, "/", nil)
	_, err := suite.client.Do(context.Background(), req, nil)

	if err == nil {
		suite.FailNow("Expected HTTP 400 error.")
	}
}

// TestErrorResponse_Error tests if the ErrorResponse function returns a non-empty error message.
func (suite *ClientTestSuite) TestErrorResponse_Error() {
	res := &http.Response{Request: &http.Request{}}
	err := ErrorResponse{Message: "m", Response: res}
	if err.Error() == "" {
		suite.FailNow("Expected non-empty ErrorResponse.Error()")
	}
}

// TestDo_completion_callback tests if the Do function calls the completion callback.
func (suite *ClientTestSuite) TestDo_completion_callback() {

	type foo struct {
		A string
	}

	suite.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			suite.FailNowf("", "Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := suite.client.NewRequest(ctx, http.MethodGet, "/", nil)
	body := new(foo)
	var completedReq *http.Request
	var completedResp string
	suite.client.SetOnRequestCompleted(func(req *http.Request, resp *http.Response) {
		completedReq = req
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			suite.FailNowf("Failed to dump response", "Failed to dump response: %s", err)
		}
		completedResp = string(b)
	})
	_, err := suite.client.Do(context.Background(), req, body)
	if err != nil {
		suite.FailNowf("", "Do(): %v", err)
	}
	if !reflect.DeepEqual(req, completedReq) {
		suite.FailNowf("", "Completed request = %v, expected %v", completedReq, req)
	}
	expected := `{"A":"a"}`
	if !strings.Contains(completedResp, expected) {
		suite.FailNowf("", "expected response to contain %v, Response = %v", expected, completedResp)
	}
}

// TestCustomUserAgent tests if the New function returns a client with a custom user agent.
func (suite *ClientTestSuite) TestCustomUserAgent() {
	ua := "testing/0.0.1"
	c, err := New(nil, WithUserAgent(ua))

	if err != nil {
		suite.FailNowf("", "New() unexpected error: %v", err)
	}

	expected := fmt.Sprintf("%s %s", ua, userAgent)
	if got := c.UserAgent; got != expected {
		suite.FailNowf("", "New() UserAgent = %s; expected %s", got, expected)
	}
}

// TestCustomBaseURL tests if the New function returns a client with a custom base URL.
func (suite *ClientTestSuite) TestCustomBaseURL() {
	baseURL := "http://localhost/foo"
	c, err := New(nil, WithBaseURL(baseURL))

	if err != nil {
		suite.FailNowf("", "New() unexpected error: %v", err)
	}

	expected := baseURL
	if got := c.BaseURL.String(); got != expected {
		suite.FailNowf("", "New() BaseURL = %s; expected %s", got, expected)
	}
}

// TestCustomBaseURL_badURL tests if the New function returns a URL parse error.
func (suite *ClientTestSuite) TestCustomBaseURL_badURL() {
	baseURL := ":"
	_, err := New(nil, WithBaseURL(baseURL))

	suite.testURLParseError(err)
}

// testMethod tests if the request method is the expected method.
func (suite *ClientTestSuite) testMethod(r *http.Request, expected string) {
	if expected != r.Method {
		suite.FailNowf("Request method = %v, expected %v", r.Method, expected)
	}
}

// MockTokenProvider is a mock implementation of TokenProvider for testing.
type MockTokenProvider struct {
	Token string
	Err   error
	Calls int
}

func (m *MockTokenProvider) GetToken(ctx context.Context) (string, error) {
	m.Calls++
	return m.Token, m.Err
}

// TestWithAccessToken tests if WithAccessToken sets the access token correctly.
func (suite *ClientTestSuite) TestWithAccessToken() {
	token := "test-token-12345"
	c, err := New(nil, WithAccessToken(token, time.Time{}))

	suite.Require().NoError(err)
	suite.Equal(token, c.accessToken)
	suite.False(c.tokenExpiry.IsZero(), "Token expiry should be set to a future time when zero is passed")
}

// TestWithAccessToken_withExpiry tests if WithAccessToken sets the access token with a specific expiry.
func (suite *ClientTestSuite) TestWithAccessToken_withExpiry() {
	token := "test-token-12345"
	expiry := time.Now().Add(1 * time.Hour)
	c, err := New(nil, WithAccessToken(token, expiry))

	suite.Require().NoError(err)
	suite.Equal(token, c.accessToken)
	suite.Equal(expiry.Unix(), c.tokenExpiry.Unix())
}

// TestWithTokenProvider tests if WithTokenProvider sets the token provider correctly.
func (suite *ClientTestSuite) TestWithTokenProvider() {
	provider := &MockTokenProvider{Token: "provider-token"}
	c, err := New(nil, WithTokenProvider(provider))

	suite.Require().NoError(err)
	suite.NotNil(c.tokenProvider)
}

// TestSetAccessToken tests if SetAccessToken sets the access token correctly.
func (suite *ClientTestSuite) TestSetAccessToken() {
	c := NewClient(nil, nil)
	token := "test-token-12345"

	c.SetAccessToken(token, time.Time{})

	suite.Equal(token, c.accessToken)
	suite.False(c.tokenExpiry.IsZero(), "Token expiry should be set to a future time when zero is passed")
}

// TestSetAccessToken_withExpiry tests if SetAccessToken sets the access token with a specific expiry.
func (suite *ClientTestSuite) TestSetAccessToken_withExpiry() {
	c := NewClient(nil, nil)
	token := "test-token-12345"
	expiry := time.Now().Add(1 * time.Hour)

	c.SetAccessToken(token, expiry)

	suite.Equal(token, c.accessToken)
	suite.Equal(expiry.Unix(), c.tokenExpiry.Unix())
}

// TestSetTokenProvider tests if SetTokenProvider sets the token provider correctly.
func (suite *ClientTestSuite) TestSetTokenProvider() {
	c := NewClient(nil, nil)
	provider := &MockTokenProvider{Token: "provider-token"}

	c.SetTokenProvider(provider)

	suite.NotNil(c.tokenProvider)
}

// TestNewRequest_withAccessToken tests if NewRequest includes the Authorization header when access token is set.
func (suite *ClientTestSuite) TestNewRequest_withAccessToken() {
	token := "my-access-token"
	c, err := New(nil, WithAccessToken(token, time.Time{}))
	suite.Require().NoError(err)

	req, err := c.NewRequest(ctx, http.MethodGet, "/test", nil)
	suite.Require().NoError(err)

	authHeader := req.Header.Get("Authorization")
	suite.Equal("Bearer "+token, authHeader)
}

// TestNewRequest_withTokenProvider tests if NewRequest uses the TokenProvider for authorization.
func (suite *ClientTestSuite) TestNewRequest_withTokenProvider() {
	provider := &MockTokenProvider{Token: "provider-token-xyz"}
	c, err := New(nil, WithTokenProvider(provider))
	suite.Require().NoError(err)

	req, err := c.NewRequest(ctx, http.MethodGet, "/test", nil)
	suite.Require().NoError(err)

	authHeader := req.Header.Get("Authorization")
	suite.Equal("Bearer provider-token-xyz", authHeader)
	suite.Equal(1, provider.Calls, "TokenProvider.GetToken should be called once")
}

// TestNewRequest_tokenProviderError tests if NewRequest returns an error when TokenProvider fails.
func (suite *ClientTestSuite) TestNewRequest_tokenProviderError() {
	provider := &MockTokenProvider{Err: fmt.Errorf("token fetch failed")}
	c, err := New(nil, WithTokenProvider(provider))
	suite.Require().NoError(err)

	_, err = c.NewRequest(ctx, http.MethodGet, "/test", nil)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to get token from provider")
}

// TestNewRequest_tokenProviderPrecedence tests that TokenProvider takes precedence over stored accessToken.
func (suite *ClientTestSuite) TestNewRequest_tokenProviderPrecedence() {
	provider := &MockTokenProvider{Token: "provider-token"}
	c, err := New(nil,
		WithAccessToken("stored-token", time.Time{}),
		WithTokenProvider(provider),
	)
	suite.Require().NoError(err)

	req, err := c.NewRequest(ctx, http.MethodGet, "/test", nil)
	suite.Require().NoError(err)

	authHeader := req.Header.Get("Authorization")
	suite.Equal("Bearer provider-token", authHeader, "TokenProvider should take precedence over stored token")
}

// TestAuthorize_withTokenProvider tests if Authorize returns the token from TokenProvider.
func (suite *ClientTestSuite) TestAuthorize_withTokenProvider() {
	provider := &MockTokenProvider{Token: "auth-provider-token"}
	c, err := New(nil, WithTokenProvider(provider))
	suite.Require().NoError(err)

	authInfo, err := c.Authorize(ctx)
	suite.Require().NoError(err)
	suite.Equal("auth-provider-token", authInfo.AccessToken)
}

// TestAuthorize_tokenProviderError tests if Authorize returns an error when TokenProvider fails.
func (suite *ClientTestSuite) TestAuthorize_tokenProviderError() {
	provider := &MockTokenProvider{Err: fmt.Errorf("auth failed")}
	c, err := New(nil, WithTokenProvider(provider))
	suite.Require().NoError(err)

	_, err = c.Authorize(ctx)
	suite.Error(err)
	suite.Contains(err.Error(), "failed to get token from provider")
}

// TestDo_withTokenProvider tests a full request cycle using TokenProvider.
func (suite *ClientTestSuite) TestDo_withTokenProvider() {
	type response struct {
		Status string `json:"status"`
	}

	// Set up handler that validates the auth header
	var receivedAuthHeader string
	suite.mux.HandleFunc("/test-endpoint", func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	provider := &MockTokenProvider{Token: "integration-test-token"}
	suite.client.SetTokenProvider(provider)

	req, err := suite.client.NewRequest(ctx, http.MethodGet, "/test-endpoint", nil)
	suite.Require().NoError(err)

	body := new(response)
	_, err = suite.client.Do(ctx, req, body)
	suite.Require().NoError(err)

	suite.Equal("Bearer integration-test-token", receivedAuthHeader)
	suite.Equal("ok", body.Status)
}

// TestTokenProvider_emptyToken tests that an empty token from provider doesn't set Authorization header.
func (suite *ClientTestSuite) TestTokenProvider_emptyToken() {
	provider := &MockTokenProvider{Token: ""}
	c, err := New(nil, WithTokenProvider(provider))
	suite.Require().NoError(err)

	req, err := c.NewRequest(ctx, http.MethodGet, "/test", nil)
	suite.Require().NoError(err)

	authHeader := req.Header.Get("Authorization")
	suite.Empty(authHeader, "Authorization header should be empty when provider returns empty token")
}
