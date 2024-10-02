package megaport

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	// Mock HTTP client and server response
	mockResponse := `{"message": "success"}`
	mockServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(mockResponse))
		if err != nil {
			suite.FailNowf("Write() unexpected error: %v", err.Error())
		}
	})
	server := httptest.NewServer(mockServer)
	defer server.Close()

	// Create a new client with the mock server URL
	c := NewClient(nil, nil)
	url, _ := url.Parse(server.URL)
	c.BaseURL = url

	// Create a new request
	req, err := c.NewRequest(ctx, http.MethodGet, "/foo", nil)
	if err != nil {
		suite.FailNowf("New() unexpected error: %v", err.Error())
	}

	// Perform the request
	resp, err := c.Do(ctx, req, nil)
	if err != nil {
		suite.FailNowf("Do() unexpected error: %v", err.Error())
	}
	defer resp.Body.Close()

	// Read and log the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		suite.FailNowf("ReadAll() unexpected error: %v", err.Error())
	}

	// Log the response body
	encodedBody := base64.StdEncoding.EncodeToString(body)
	c.Logger.DebugContext(ctx, "response_body", slog.String("response_body_base_64", encodedBody))

	// Check the response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		suite.FailNowf("Unmarshal() unexpected error: %v", err.Error())
	}

	expectedMessage := "success"
	resultMsg, ok := result["message"].(string)
	if !ok {
		suite.FailNow("Response message is not a string")
	}

	if result["message"] != expectedMessage {
		suite.FailNowf("Response message = %s; expected %s", resultMsg, expectedMessage)
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
