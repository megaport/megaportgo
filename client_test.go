package megaport

import (
	"context"
	"fmt"
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

type ClientTestSuite struct {
	suite.Suite
	client *Client
	server *httptest.Server
	mux    *http.ServeMux
}

func TestClientTestSuite(t *testing.T) {
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

// type values map[string]string

// func testFormValues(t *testing.T, r *http.Request, values values) {
// 	expected := url.Values{}
// 	for k, v := range values {
// 		expected.Add(k, v)
// 	}

// 	err := r.ParseForm()
// 	if err != nil {
// 		t.Fatalf("parseForm(): %v", err)
// 	}

// 	if !reflect.DeepEqual(expected, r.Form) {
// 		t.Errorf("Request parameters = %v, expected %v", r.Form, expected)
// 	}
// }

func (suite *ClientTestSuite) testURLParseError(err error) {
	if err == nil {
		suite.FailNow("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		suite.FailNowf("Expected URL parse error, got %+v", err.Error())
	}
}

func (suite *ClientTestSuite) testClientDefaultBaseURL(c *Client) {
	if c.BaseURL == nil || c.BaseURL.String() != defaultBaseURL {
		suite.FailNowf("NewClient BaseURL = %v, expected %v", c.BaseURL.String(), defaultBaseURL)
	}
}

func (suite *ClientTestSuite) testClientDefaultUserAgent(c *Client) {
	if c.UserAgent != userAgent {
		suite.FailNowf("NewClient UserAgent = %v, expected %v", c.UserAgent, userAgent)
	}
}

func (suite *ClientTestSuite) testClientDefaults(c *Client) {
	suite.testClientDefaultBaseURL(c)
	suite.testClientDefaultUserAgent(c)
}

func (suite *ClientTestSuite) TestNewClient() {
	c := NewClient(nil, nil)
	suite.testClientDefaults(c)
}

func (suite *ClientTestSuite) TestNewFromToken() {
	c := NewFromToken("myToken")
	suite.testClientDefaults(c)
}

func (suite *ClientTestSuite) TestNewFromToken_cleaned() {
	testTokens := []string{"myToken ", " myToken", " myToken ", "'myToken'", " 'myToken' "}
	expected := "Bearer myToken"

	suite.mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, tt := range testTokens {
		c := NewFromToken(tt)
		req, _ := c.NewRequest(ctx, http.MethodGet, suite.server.URL+"/foo", nil)
		resp, err := c.Do(ctx, req, nil)
		if err != nil {
			suite.FailNowf("Do(): %v", err.Error())
		}

		authHeader := resp.Request.Header.Get("Authorization")
		if authHeader != expected {
			suite.FailNowf("Authorization header = %v, expected %v", authHeader, expected)
		}
	}
}

func (suite *ClientTestSuite) TestNew() {
	c, err := New(nil)

	if err != nil {
		suite.FailNowf("New(): %v", err.Error())
	}
	suite.testClientDefaults(c)
}

func (suite *ClientTestSuite) TestNewRequest_get() {
	c := NewClient(nil, nil)

	inURL, outURL := "/foo", defaultBaseURL+"foo"
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

func (suite *ClientTestSuite) TestNewRequest_badURL() {
	c := NewClient(nil, nil)
	_, err := c.NewRequest(ctx, http.MethodGet, ":", nil)
	suite.testURLParseError(err)
}

func (suite *ClientTestSuite) TestNewRequest_withCustomUserAgent() {
	ua := "testing/0.0.1"
	c, err := New(nil, SetUserAgent(ua))

	if err != nil {
		suite.FailNowf("New() unexpected error: %v", err.Error())
	}

	req, _ := c.NewRequest(ctx, http.MethodGet, "/foo", nil)

	expected := fmt.Sprintf("%s %s", ua, userAgent)
	if got := req.Header.Get("User-Agent"); got != expected {
		suite.FailNowf("New() UserAgent = %s; expected %s", got, expected)
	}
}

func (suite *ClientTestSuite) TestNewRequest_withCustomHeaders() {
	expectedIdentity := "identity"
	expectedCustom := "x_test_header"

	c, err := New(nil, SetRequestHeaders(map[string]string{
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

func (suite *ClientTestSuite) TestErrorResponse_Error() {
	res := &http.Response{Request: &http.Request{}}
	err := ErrorResponse{Message: "m", Response: res}
	if err.Error() == "" {
		suite.FailNow("Expected non-empty ErrorResponse.Error()")
	}
}

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
	suite.client.OnRequestCompleted(func(req *http.Request, resp *http.Response) {
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

func (suite *ClientTestSuite) TestCustomUserAgent() {
	ua := "testing/0.0.1"
	c, err := New(nil, SetUserAgent(ua))

	if err != nil {
		suite.FailNowf("", "New() unexpected error: %v", err)
	}

	expected := fmt.Sprintf("%s %s", ua, userAgent)
	if got := c.UserAgent; got != expected {
		suite.FailNowf("", "New() UserAgent = %s; expected %s", got, expected)
	}
}

func (suite *ClientTestSuite) TestCustomBaseURL() {
	baseURL := "http://localhost/foo"
	c, err := New(nil, SetBaseURL(baseURL))

	if err != nil {
		suite.FailNowf("", "New() unexpected error: %v", err)
	}

	expected := baseURL
	if got := c.BaseURL.String(); got != expected {
		suite.FailNowf("", "New() BaseURL = %s; expected %s", got, expected)
	}
}

func (suite *ClientTestSuite) TestCustomBaseURL_badURL() {
	baseURL := ":"
	_, err := New(nil, SetBaseURL(baseURL))

	suite.testURLParseError(err)
}

func (suite *ClientTestSuite) testMethod(r *http.Request, expected string) {
	if expected != r.Method {
		suite.FailNowf("Request method = %v, expected %v", r.Method, expected)
	}
}
