package megaport

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

var (
	mux *http.ServeMux

	ctx = context.TODO()

	client *Client

	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = NewClient(nil, nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url
}

func teardown() {
	server.Close()
}

func testMethod(t *testing.T, r *http.Request, expected string) {
	if expected != r.Method {
		t.Errorf("Request method = %v, expected %v", r.Method, expected)
	}
}

type values map[string]string

func testFormValues(t *testing.T, r *http.Request, values values) {
	expected := url.Values{}
	for k, v := range values {
		expected.Add(k, v)
	}

	err := r.ParseForm()
	if err != nil {
		t.Fatalf("parseForm(): %v", err)
	}

	if !reflect.DeepEqual(expected, r.Form) {
		t.Errorf("Request parameters = %v, expected %v", r.Form, expected)
	}
}

func testURLParseError(t *testing.T, err error) {
	if err == nil {
		t.Errorf("Expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("Expected URL parse error, got %+v", err)
	}
}

func testClientDefaultBaseURL(t *testing.T, c *Client) {
	if c.BaseURL == nil || c.BaseURL.String() != defaultBaseURL {
		t.Errorf("NewClient BaseURL = %v, expected %v", c.BaseURL, defaultBaseURL)
	}
}

func testClientDefaultUserAgent(t *testing.T, c *Client) {
	if c.UserAgent != userAgent {
		t.Errorf("NewClient UserAgent = %v, expected %v", c.UserAgent, userAgent)
	}
}

func testClientDefaults(t *testing.T, c *Client) {
	testClientDefaultBaseURL(t, c)
	testClientDefaultUserAgent(t, c)
}

func TestNewClient(t *testing.T) {
	c := NewClient(nil, nil)
	testClientDefaults(t, c)
}

func TestNewFromToken(t *testing.T) {
	c := NewFromToken("myToken")
	testClientDefaults(t, c)
}

func TestNewFromToken_cleaned(t *testing.T) {
	testTokens := []string{"myToken ", " myToken", " myToken ", "'myToken'", " 'myToken' "}
	expected := "Bearer myToken"

	setup()
	defer teardown()

	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, tt := range testTokens {
		t.Run(tt, func(t *testing.T) {
			c := NewFromToken(tt)
			req, _ := c.NewRequest(ctx, http.MethodGet, server.URL+"/foo", nil)
			resp, err := c.Do(ctx, req, nil)
			if err != nil {
				t.Fatalf("Do(): %v", err)
			}

			authHeader := resp.Request.Header.Get("Authorization")
			if authHeader != expected {
				t.Errorf("Authorization header = %v, expected %v", authHeader, expected)
			}
		})
	}
}

func TestNew(t *testing.T) {
	c, err := New(nil)

	if err != nil {
		t.Fatalf("New(): %v", err)
	}
	testClientDefaults(t, c)
}

func TestNewRequest_get(t *testing.T) {
	c := NewClient(nil, nil)

	inURL, outURL := "/foo", defaultBaseURL+"foo"
	req, _ := c.NewRequest(ctx, http.MethodGet, inURL, nil)

	// test relative URL was expanded
	if req.URL.String() != outURL {
		t.Errorf("NewRequest(%v) URL = %v, expected %v", inURL, req.URL, outURL)
	}

	// test the content-type header is not set
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		t.Errorf("NewRequest() Content-Type = %v, expected empty string", contentType)
	}

	// test default user-agent is attached to the request
	userAgent := req.Header.Get("User-Agent")
	if c.UserAgent != userAgent {
		t.Errorf("NewRequest() User-Agent = %v, expected %v", userAgent, c.UserAgent)
	}
}

func TestNewRequest_badURL(t *testing.T) {
	c := NewClient(nil, nil)
	_, err := c.NewRequest(ctx, http.MethodGet, ":", nil)
	testURLParseError(t, err)
}

func TestNewRequest_withCustomUserAgent(t *testing.T) {
	ua := "testing/0.0.1"
	c, err := New(nil, SetUserAgent(ua))

	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	req, _ := c.NewRequest(ctx, http.MethodGet, "/foo", nil)

	expected := fmt.Sprintf("%s %s", ua, userAgent)
	if got := req.Header.Get("User-Agent"); got != expected {
		t.Errorf("New() UserAgent = %s; expected %s", got, expected)
	}
}

func TestNewRequest_withCustomHeaders(t *testing.T) {
	expectedIdentity := "identity"
	expectedCustom := "x_test_header"

	c, err := New(nil, SetRequestHeaders(map[string]string{
		"Accept-Encoding": expectedIdentity,
		"X-Test-Header":   expectedCustom,
	}))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	req, _ := c.NewRequest(ctx, http.MethodGet, "/foo", nil)

	if got := req.Header.Get("Accept-Encoding"); got != expectedIdentity {
		t.Errorf("New() Custom Accept Encoding Header = %s; expected %s", got, expectedIdentity)
	}
	if got := req.Header.Get("X-Test-Header"); got != expectedCustom {
		t.Errorf("New() Custom Accept Encoding Header = %s; expected %s", got, expectedCustom)
	}
}

func TestDo(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			t.Errorf("Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	body := new(foo)
	_, err := client.Do(context.Background(), req, body)
	if err != nil {
		t.Fatalf("Do(): %v", err)
	}

	expected := &foo{"a"}
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("Response body = %v, expected %v", body, expected)
	}
}

func TestDo_httpError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected HTTP 400 error.")
	}
}

// Test handling of an error caused by the internal http client's Do()
// function.
func TestDo_redirectLoop(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	_, err := client.Do(context.Background(), req, nil)

	if err == nil {
		t.Error("Expected error to be returned.")
	}
	if err, ok := err.(*url.Error); !ok {
		t.Errorf("Expected a URL error; got %#v.", err)
	}
}

func TestErrorResponse_Error(t *testing.T) {
	res := &http.Response{Request: &http.Request{}}
	err := ErrorResponse{Message: "m", Response: res}
	if err.Error() == "" {
		t.Errorf("Expected non-empty ErrorResponse.Error()")
	}
}

// TestWithRetryAndBackoffs tests the retryablehttp client's default retry policy.
func TestWithRetryAndBackoffs(t *testing.T) {
	// Mock server which always responds 500.
	setup()
	defer teardown()

	url, _ := url.Parse(server.URL)
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"id": "bad_request", "message": "broken"}`))
	})

	tokenSrc := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "new_token",
	})

	oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSrc)

	waitMax := PtrTo(6.0)
	waitMin := PtrTo(3.0)

	retryConfig := RetryConfig{
		RetryMax:     3,
		RetryWaitMin: waitMin,
		RetryWaitMax: waitMax,
	}

	// Create the client. Use short retry windows so we fail faster.
	client, err := New(oauthClient, WithRetryAndBackoffs(retryConfig))
	client.BaseURL = url
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Create the request
	req, err := client.NewRequest(ctx, http.MethodGet, "/foo", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	expectingErr := fmt.Sprintf("GET %s/foo: 500 broken; giving up after 4 attempt(s)", url)
	// Send the request.
	_, err = client.Do(context.Background(), req, nil)
	if err == nil || (err.Error() != expectingErr) {
		t.Fatalf("expected giving up error, got: %#v", err)
	}
}

func TestWithRetryAndBackoffsLogger(t *testing.T) {
	// Mock server which always responds 500.
	setup()
	defer teardown()

	url, _ := url.Parse(server.URL)
	mux.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})

	tokenSrc := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: "new_token",
	})

	oauth_client := oauth2.NewClient(oauth2.NoContext, tokenSrc)

	var buf bytes.Buffer
	retryConfig := RetryConfig{
		RetryMax: 3,
		Logger:   log.New(&buf, "", 0),
	}

	// Create the client. Use short retry windows so we fail faster.
	client, err := New(oauth_client, WithRetryAndBackoffs(retryConfig))
	client.BaseURL = url
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Create the request
	req, err := client.NewRequest(ctx, http.MethodGet, "/foo", nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	_, err = client.Do(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	got := buf.String()
	expected := fmt.Sprintf("[DEBUG] GET %s/foo\n", url)
	if expected != got {
		t.Fatalf("expected: %s; got: %s", expected, got)
	}
}

func TestDo_completion_callback(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := http.MethodGet; m != r.Method {
			t.Errorf("Request method = %v, expected %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	req, _ := client.NewRequest(ctx, http.MethodGet, "/", nil)
	body := new(foo)
	var completedReq *http.Request
	var completedResp string
	client.OnRequestCompleted(func(req *http.Request, resp *http.Response) {
		completedReq = req
		b, err := httputil.DumpResponse(resp, true)
		if err != nil {
			t.Errorf("Failed to dump response: %s", err)
		}
		completedResp = string(b)
	})
	_, err := client.Do(context.Background(), req, body)
	if err != nil {
		t.Fatalf("Do(): %v", err)
	}
	if !reflect.DeepEqual(req, completedReq) {
		t.Errorf("Completed request = %v, expected %v", completedReq, req)
	}
	expected := `{"A":"a"}`
	if !strings.Contains(completedResp, expected) {
		t.Errorf("expected response to contain %v, Response = %v", expected, completedResp)
	}
}

func TestCustomUserAgent(t *testing.T) {
	ua := "testing/0.0.1"
	c, err := New(nil, SetUserAgent(ua))

	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	expected := fmt.Sprintf("%s %s", ua, userAgent)
	if got := c.UserAgent; got != expected {
		t.Errorf("New() UserAgent = %s; expected %s", got, expected)
	}
}

func TestCustomBaseURL(t *testing.T) {
	baseURL := "http://localhost/foo"
	c, err := New(nil, SetBaseURL(baseURL))

	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	expected := baseURL
	if got := c.BaseURL.String(); got != expected {
		t.Errorf("New() BaseURL = %s; expected %s", got, expected)
	}
}

func TestCustomBaseURL_badURL(t *testing.T) {
	baseURL := ":"
	_, err := New(nil, SetBaseURL(baseURL))

	testURLParseError(t, err)
}
