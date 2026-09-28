package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TelemetryClientTestSuite struct {
	ClientTestSuite
}

func TestTelemetryClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(TelemetryClientTestSuite))
}

func (suite *TelemetryClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *TelemetryClientTestSuite) TearDownTest() {
	suite.server.Close()
}

const telemetryTestUID = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

type telemetryMethod func(context.Context, *GetTelemetryRequest) (*ServiceTelemetryResponse, error)

func (suite *TelemetryClientTestSuite) telemetryMethods() map[string]telemetryMethod {
	return map[string]telemetryMethod{
		"megaport": suite.client.PortService.GetPortTelemetry,
		"mcr2":     suite.client.MCRService.GetMCRTelemetry,
		"mve":      suite.client.MVEService.GetMVETelemetry,
		"vxc":      suite.client.VXCService.GetVXCTelemetry,
		"ix":       suite.client.IXService.GetIXTelemetry,
	}
}

// Each method must call its own product type path and decode the response.
func (suite *TelemetryClientTestSuite) TestGetTelemetryPaths() {
	jblob := `{
		"serviceUid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"type": "BITS",
		"timeFrame": {"from": 1608516536000, "to": 1608603936000},
		"data": [
			{
				"type": "BITS",
				"subtype": "IN",
				"samples": [[1608516536000, 125.5], [1608517536000, 130.2]],
				"unit": {"name": "Mbps", "fullName": "Megabits per second"}
			}
		]
	}`
	var gotPath string
	suite.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal(url.Values{"type": {"BITS"}, "days": {"7"}}, r.URL.Query())
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	for productType, get := range suite.telemetryMethods() {
		resp, err := get(context.Background(), &GetTelemetryRequest{
			ProductUID: telemetryTestUID,
			Types:      []string{"BITS"},
			Days:       PtrTo[int32](7),
		})
		suite.Require().NoError(err, productType)
		suite.Equal(fmt.Sprintf("/v2/product/%s/%s/telemetry", productType, telemetryTestUID), gotPath)
		suite.Equal(&ServiceTelemetryResponse{
			ServiceUID: telemetryTestUID,
			Type:       "BITS",
			TimeFrame:  TelemetryTimeFrame{From: 1608516536000, To: 1608603936000},
			Data: []*TelemetryMetricData{{
				Type:    "BITS",
				Subtype: "IN",
				Samples: []TelemetrySample{{Timestamp: 1608516536000, Value: 125.5}, {Timestamp: 1608517536000, Value: 130.2}},
				Unit:    TelemetryUnit{Name: "Mbps", FullName: "Megabits per second"},
			}},
		}, resp, productType)
	}
}

// The SDK forwards the request fields as query parameters and leaves range checks to the API.
func (suite *TelemetryClientTestSuite) TestGetTelemetryQuery() {
	from := time.UnixMilli(1608516536000)
	tests := []struct {
		name string
		req  *GetTelemetryRequest
		want url.Values
	}{
		{
			name: "from and to in epoch milliseconds",
			req:  &GetTelemetryRequest{Types: []string{"BITS", "PACKETS"}, From: &from, To: PtrTo(from.Add(24 * time.Hour))},
			want: url.Values{"type": {"BITS", "PACKETS"}, "from": {"1608516536000"}, "to": {"1608602936000"}},
		},
		{
			name: "days",
			req:  &GetTelemetryRequest{Types: []string{"SPEED"}, Days: PtrTo[int32](180)},
			want: url.Values{"type": {"SPEED"}, "days": {"180"}},
		},
		{
			name: "no types or range",
			req:  &GetTelemetryRequest{},
			want: url.Values{},
		},
	}

	var got url.Values
	suite.mux.HandleFunc("/v2/product/mcr2/"+telemetryTestUID+"/telemetry", func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		fmt.Fprint(w, `{"serviceUid":"a1b2c3d4-e5f6-7890-abcd-ef1234567890","data":[]}`)
	})
	for _, tt := range tests {
		tt.req.ProductUID = telemetryTestUID
		_, err := suite.client.MCRService.GetMCRTelemetry(context.Background(), tt.req)
		suite.Require().NoError(err, tt.name)
		suite.Equal(tt.want, got, tt.name)
	}
}

func (suite *TelemetryClientTestSuite) TestGetTelemetryRequestErrors() {
	for productType, get := range suite.telemetryMethods() {
		_, err := get(context.Background(), nil)
		suite.ErrorIs(err, ErrTelemetryRequestNil, productType)

		_, err = get(context.Background(), &GetTelemetryRequest{Types: []string{"BITS"}})
		suite.ErrorIs(err, ErrTelemetryProductUIDRequired, productType)
	}
}

func (suite *TelemetryClientTestSuite) TestGetTelemetryAPIError() {
	suite.mux.HandleFunc("/v2/product/mcr2/"+telemetryTestUID+"/telemetry", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"'from' must be before 'to'","data":null}`)
	})

	resp, err := suite.client.MCRService.GetMCRTelemetry(context.Background(), &GetTelemetryRequest{
		ProductUID: telemetryTestUID,
		Types:      []string{"BITS"},
		From:       PtrTo(time.UnixMilli(1608516536000)),
		To:         PtrTo(time.UnixMilli(1608516536000)),
	})
	suite.Nil(resp)
	var apiErr *ErrorResponse
	suite.Require().True(errors.As(err, &apiErr))
	suite.Equal(http.StatusBadRequest, apiErr.Response.StatusCode)
	suite.Equal("'from' must be before 'to'", apiErr.Message)
}

func (suite *TelemetryClientTestSuite) TestGetTelemetryDecodeErrorHasTraceID() {
	suite.mux.HandleFunc("/v2/product/mcr2/"+telemetryTestUID+"/telemetry", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Trace-Id", "abc123")
		fmt.Fprint(w, `{"data":[{"samples":[[1608516536000]]}]}`)
	})

	_, err := suite.client.MCRService.GetMCRTelemetry(context.Background(), &GetTelemetryRequest{
		ProductUID: telemetryTestUID,
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](1),
	})
	suite.ErrorContains(err, "trace ID abc123")
}

func TestTelemetrySampleUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    TelemetrySample
		wantErr bool
	}{
		{
			name:  "timestamp and value",
			input: `[1700000000000, 12.5]`,
			want:  TelemetrySample{Timestamp: 1700000000000, Value: 12.5},
		},
		{
			name:  "integer value",
			input: `[1700000000000, 12]`,
			want:  TelemetrySample{Timestamp: 1700000000000, Value: 12},
		},
		{
			// The API sends null when a metric had no reading in the interval.
			name:  "null value decodes as zero",
			input: `[1700000000000, null]`,
			want:  TelemetrySample{Timestamp: 1700000000000, Value: 0},
		},
		{name: "null timestamp", input: `[null, 12.5]`, wantErr: true},
		{name: "too few elements", input: `[1700000000000]`, wantErr: true},
		{name: "too many elements", input: `[1700000000000, 12.5, 1]`, wantErr: true},
		{name: "not an array", input: `{"timestamp":1700000000000}`, wantErr: true},
		{name: "non-numeric value", input: `[1700000000000, "high"]`, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got TelemetrySample
			err := json.Unmarshal([]byte(tt.input), &got)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// A single null value must not fail the decode of the surrounding series.
func TestTelemetrySampleNullWithinSeries(t *testing.T) {
	var metric TelemetryMetricData
	require.NoError(t, json.Unmarshal([]byte(`{
		"type": "Bits",
		"subtype": "in",
		"samples": [[1700000000000, 1.5], [1700000060000, null], [1700000120000, 2.5]],
		"unit": {"name": "bps", "fullName": "bits per second"}
	}`), &metric))

	require.Len(t, metric.Samples, 3)
	assert.Equal(t, 1.5, metric.Samples[0].Value)
	assert.Equal(t, 0.0, metric.Samples[1].Value)
	assert.Equal(t, int64(1700000060000), metric.Samples[1].Timestamp)
	assert.Equal(t, 2.5, metric.Samples[2].Value)
}
