package megaport

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

// NATGatewayClientTestSuite tests the NAT Gateway Service Client.
type NATGatewayClientTestSuite struct {
	ClientTestSuite
}

func TestNATGatewayClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(NATGatewayClientTestSuite))
}

func (suite *NATGatewayClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *NATGatewayClientTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewaySessions() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"sessionCount": [1000, 2000, 4000],
				"speedMbps": 100
			},
			{
				"sessionCount": [8000, 16000],
				"speedMbps": 1000
			}
		]
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways/sessions", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	sessions, err := natSvc.ListNATGatewaySessions(ctx)
	suite.NoError(err)
	suite.Len(sessions, 2)

	suite.Equal(100, sessions[0].SpeedMbps)
	suite.Equal([]int{1000, 2000, 4000}, sessions[0].SessionCount)

	suite.Equal(1000, sessions[1].SpeedMbps)
	suite.Equal([]int{8000, 16000}, sessions[1].SessionCount)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewaySessionsEmpty() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Success",
		"terms": "https://www.megaport.com/legal/acceptable-use-policy",
		"data": []
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways/sessions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	sessions, err := natSvc.ListNATGatewaySessions(ctx)
	suite.NoError(err)
	suite.Empty(sessions)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayTelemetry() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

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

	path := fmt.Sprintf("/v3/products/nat_gateways/%s/telemetry", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal("7", r.URL.Query().Get("days"))
		suite.Equal([]string{"BITS"}, r.URL.Query()["type"])
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	resp, err := natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: productUID,
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](7),
	})
	suite.NoError(err)
	suite.Equal(productUID, resp.ServiceUID)
	suite.Equal("BITS", resp.Type)
	suite.Equal(int64(1608516536000), resp.TimeFrame.From)
	suite.Equal(int64(1608603936000), resp.TimeFrame.To)
	suite.Len(resp.Data, 1)
	suite.Equal("BITS", resp.Data[0].Type)
	suite.Equal("IN", resp.Data[0].Subtype)
	suite.Len(resp.Data[0].Samples, 2)
	suite.Equal(int64(1608516536000), resp.Data[0].Samples[0].Timestamp)
	suite.Equal(125.5, resp.Data[0].Samples[0].Value)
	suite.Equal(int64(1608517536000), resp.Data[0].Samples[1].Timestamp)
	suite.Equal(130.2, resp.Data[0].Samples[1].Value)
	suite.Equal("Mbps", resp.Data[0].Unit.Name)
	suite.Equal("Megabits per second", resp.Data[0].Unit.FullName)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayTelemetryWithFromTo() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"

	jblob := `{
		"serviceUid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		"type": "BITS",
		"timeFrame": {"from": 1608516536000, "to": 1608603936000},
		"data": []
	}`

	path := fmt.Sprintf("/v3/products/nat_gateways/%s/telemetry", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		suite.Equal("1608516536000", r.URL.Query().Get("from"))
		suite.Equal("1608603936000", r.URL.Query().Get("to"))
		suite.Equal([]string{"BITS", "PACKETS"}, r.URL.Query()["type"])
		suite.Empty(r.URL.Query().Get("days"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	resp, err := natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: productUID,
		Types:      []string{"BITS", "PACKETS"},
		From:       PtrTo[int64](1608516536000),
		To:         PtrTo[int64](1608603936000),
	})
	suite.NoError(err)
	suite.Equal(productUID, resp.ServiceUID)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayTelemetryValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	// Missing ProductUID
	_, err := natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		Types: []string{"BITS"},
		Days:  PtrTo[int32](7),
	})
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	// Missing Types
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Days:       PtrTo[int32](7),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryTypesRequired)

	// Days and From/To mutually exclusive
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](7),
		From:       PtrTo[int64](1608516536000),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryTimeExclusive)
}
