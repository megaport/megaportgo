package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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

func (suite *NATGatewayClientTestSuite) TestCreateNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": {
			"adminLocked": false,
			"autoRenewTerm": true,
			"config": {
				"asn": 64512,
				"bgpShutdownDefault": false,
				"diversityZone": "red",
				"sessionCount": 100
			},
			"contractEndDate": "2024-10-01T14:34:56Z",
			"createDate": "2023-10-01T14:34:56Z",
			"createdBy": "user-name",
			"locationId": 123456,
			"locked": false,
			"orderApprovalStatus": "PENDING",
			"productName": "NAT Gateway",
			"productUid": "e900d0d5-1030-4e29-b2d8-816ad4263190",
			"promoCode": "PROMO123",
			"provisioningStatus": "NEW",
			"resourceTags": [{"key": "env", "value": "test"}],
			"serviceLevelReference": "SLR-1",
			"speed": 1000,
			"term": 1
		}
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gw, err := natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		AutoRenewTerm: true,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: false,
			DiversityZone:      "red",
			SessionCount:       100,
		},
		LocationID:            123456,
		ProductName:           "NAT Gateway",
		PromoCode:             "PROMO123",
		ResourceTags:          []ResourceTag{{Key: "env", Value: "test"}},
		ServiceLevelReference: "SLR-1",
		Speed:                 1000,
		Term:                  1,
	})
	suite.NoError(err)
	suite.Equal("e900d0d5-1030-4e29-b2d8-816ad4263190", gw.ProductUID)
	suite.Equal("NAT Gateway", gw.ProductName)
	suite.Equal(1000, gw.Speed)
	suite.Equal(1, gw.Term)
	suite.Equal(123456, gw.LocationID)
	suite.True(gw.AutoRenewTerm)
	suite.Equal(64512, gw.Config.ASN)
	suite.False(gw.Config.BGPShutdownDefault)
	suite.Equal("red", gw.Config.DiversityZone)
	suite.Equal(100, gw.Config.SessionCount)
	suite.Equal("NEW", gw.ProvisioningStatus)
	suite.Equal("PENDING", gw.OrderApprovalStatus)
	suite.Len(gw.ResourceTags, 1)
	suite.Equal("env", gw.ResourceTags[0].Key)
	suite.Equal("test", gw.ResourceTags[0].Value)
}

func (suite *NATGatewayClientTestSuite) TestCreateNATGatewayValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	// Missing ProductName
	_, err := natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		LocationID: 1, Speed: 1000, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewayProductNameRequired)

	// Invalid LocationID
	_, err = natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		ProductName: "test", LocationID: 0, Speed: 1000, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewayLocationIDRequired)

	// Invalid Speed
	_, err = natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		ProductName: "test", LocationID: 1, Speed: 0, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewaySpeedRequired)

	// Invalid Term
	_, err = natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		ProductName: "test", LocationID: 1, Speed: 1000, Term: 5,
	})
	suite.ErrorIs(err, ErrNATGatewayInvalidTerm)
}

func (suite *NATGatewayClientTestSuite) TestListNATGateways() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": [
			{
				"adminLocked": false,
				"autoRenewTerm": true,
				"config": {"asn": 64512, "bgpShutdownDefault": false, "diversityZone": "red", "sessionCount": 100},
				"locationId": 123456,
				"productName": "NAT Gateway 1",
				"productUid": "uid-1",
				"provisioningStatus": "LIVE",
				"speed": 1000,
				"term": 12
			},
			{
				"adminLocked": false,
				"autoRenewTerm": false,
				"config": {"asn": 64513, "bgpShutdownDefault": true, "diversityZone": "blue", "sessionCount": 200},
				"locationId": 789012,
				"productName": "NAT Gateway 2",
				"productUid": "uid-2",
				"provisioningStatus": "NEW",
				"speed": 2500,
				"term": 24
			}
		]
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gateways, err := natSvc.ListNATGateways(ctx)
	suite.NoError(err)
	suite.Len(gateways, 2)

	suite.Equal("uid-1", gateways[0].ProductUID)
	suite.Equal("NAT Gateway 1", gateways[0].ProductName)
	suite.Equal(1000, gateways[0].Speed)
	suite.Equal("LIVE", gateways[0].ProvisioningStatus)

	suite.Equal("uid-2", gateways[1].ProductUID)
	suite.Equal("NAT Gateway 2", gateways[1].ProductName)
	suite.Equal(2500, gateways[1].Speed)
	suite.True(gateways[1].Config.BGPShutdownDefault)
}

func (suite *NATGatewayClientTestSuite) TestListNATGatewaysEmpty() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": []
	}`

	suite.mux.HandleFunc("/v3/products/nat_gateways", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gateways, err := natSvc.ListNATGateways(ctx)
	suite.NoError(err)
	suite.Empty(gateways)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "e900d0d5-1030-4e29-b2d8-816ad4263190"

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": {
			"adminLocked": false,
			"autoRenewTerm": true,
			"config": {"asn": 64512, "bgpShutdownDefault": false, "diversityZone": "red", "sessionCount": 100},
			"contractEndDate": "2024-10-01T14:34:56Z",
			"createDate": "2023-10-01T14:34:56Z",
			"createdBy": "user-name",
			"locationId": 123456,
			"locked": false,
			"orderApprovalStatus": "APPROVED",
			"productName": "NAT Gateway",
			"productUid": "e900d0d5-1030-4e29-b2d8-816ad4263190",
			"provisioningStatus": "LIVE",
			"speed": 1000,
			"term": 12
		}
	}`

	path := fmt.Sprintf("/v3/products/nat_gateways/%s", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gw, err := natSvc.GetNATGateway(ctx, productUID)
	suite.NoError(err)
	suite.Equal(productUID, gw.ProductUID)
	suite.Equal("NAT Gateway", gw.ProductName)
	suite.Equal("LIVE", gw.ProvisioningStatus)
	suite.Equal("APPROVED", gw.OrderApprovalStatus)
	suite.Equal(12, gw.Term)
}

func (suite *NATGatewayClientTestSuite) TestGetNATGatewayValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	_, err := natSvc.GetNATGateway(ctx, "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestUpdateNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "e900d0d5-1030-4e29-b2d8-816ad4263190"

	jblob := `{
		"message": "Data returned successfully",
		"terms": "This data is subject to the Acceptable Use Policy",
		"data": {
			"adminLocked": false,
			"autoRenewTerm": false,
			"config": {"asn": 64512, "bgpShutdownDefault": true, "diversityZone": "blue", "sessionCount": 200},
			"locationId": 123456,
			"productName": "Updated NAT Gateway",
			"productUid": "e900d0d5-1030-4e29-b2d8-816ad4263190",
			"provisioningStatus": "LIVE",
			"speed": 1000,
			"term": 24
		}
	}`

	path := fmt.Sprintf("/v3/products/nat_gateways/%s", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPut, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, jblob)
	})

	gw, err := natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID:    productUID,
		AutoRenewTerm: false,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: true,
			DiversityZone:      "blue",
			SessionCount:       200,
		},
		LocationID:  123456,
		ProductName: "Updated NAT Gateway",
		Speed:       1000,
		Term:        24,
	})
	suite.NoError(err)
	suite.Equal("Updated NAT Gateway", gw.ProductName)
	suite.Equal(24, gw.Term)
	suite.True(gw.Config.BGPShutdownDefault)
	suite.Equal("blue", gw.Config.DiversityZone)
	suite.Equal(200, gw.Config.SessionCount)
}

func (suite *NATGatewayClientTestSuite) TestUpdateNATGatewayValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	// Missing ProductUID
	_, err := natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductName: "test", LocationID: 1, Speed: 1000, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)

	// Missing ProductName
	_, err = natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID: "uid", LocationID: 1, Speed: 1000, Term: 1,
	})
	suite.ErrorIs(err, ErrNATGatewayProductNameRequired)

	// Invalid Term
	_, err = natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID: "uid", ProductName: "test", LocationID: 1, Speed: 1000, Term: 7,
	})
	suite.ErrorIs(err, ErrNATGatewayInvalidTerm)
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "e900d0d5-1030-4e29-b2d8-816ad4263190"

	path := fmt.Sprintf("/v3/products/nat_gateways/%s", productUID)
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodDelete, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message": "Nat gateway order item deleted successfully", "terms": ""}`)
	})

	err := natSvc.DeleteNATGateway(ctx, productUID)
	suite.NoError(err)
}

func (suite *NATGatewayClientTestSuite) TestDeleteNATGatewayValidation() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService

	err := natSvc.DeleteNATGateway(ctx, "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
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

	fromTime := time.UnixMilli(1608516536000)
	toTime := time.UnixMilli(1608603936000)
	resp, err := natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: productUID,
		Types:      []string{"BITS", "PACKETS"},
		From:       &fromTime,
		To:         &toTime,
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
		From:       PtrTo(time.UnixMilli(1608516536000)),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryTimeExclusive)

	// Days out of range (too low)
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](0),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryDaysOutOfRange)

	// Days out of range (too high)
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		Days:       PtrTo[int32](181),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryDaysOutOfRange)

	// Only From without To
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		From:       PtrTo(time.UnixMilli(1608516536000)),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryFromToIncomplete)

	// Only To without From
	_, err = natSvc.GetNATGatewayTelemetry(ctx, &GetNATGatewayTelemetryRequest{
		ProductUID: "some-uid",
		Types:      []string{"BITS"},
		To:         PtrTo(time.UnixMilli(1608603936000)),
	})
	suite.ErrorIs(err, ErrNATGatewayTelemetryFromToIncomplete)
}

func (suite *NATGatewayClientTestSuite) TestValidateNATGatewayOrder() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "11111111-2222-3333-4444-555555555555"

	called := false
	suite.mux.HandleFunc("/v3/networkdesign/validate", func(w http.ResponseWriter, r *http.Request) {
		called = true
		suite.Equal(http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		suite.NoError(err)
		var payload []map[string]string
		suite.NoError(json.Unmarshal(body, &payload))
		suite.Equal([]map[string]string{{"productUid": productUID}}, payload)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"message":"Validation passed",
			"terms":"",
			"data":[{
				"productUid":%q,
				"productType":"NAT_GATEWAY",
				"string":"Sydney",
				"price":{
					"monthlyRate":600,
					"mbpsRate":0.6,
					"currency":"AUD",
					"productType":"NAT_GATEWAY",
					"monthlyRackRate":600
				}
			}]
		}`, productUID)
	})

	result, err := natSvc.ValidateNATGatewayOrder(ctx, productUID)
	suite.NoError(err)
	suite.True(called)
	suite.Equal(productUID, result.ProductUID)
	suite.Equal("NAT_GATEWAY", result.ProductType)
	suite.Equal("Sydney", result.Metro)
	suite.Equal(float64(600), result.Price.MonthlyRate)
	suite.Equal("AUD", result.Price.Currency)
}

func (suite *NATGatewayClientTestSuite) TestValidateNATGatewayOrder_EmptyData() {
	ctx := context.Background()
	productUID := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/networkdesign/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[]}`)
	})

	_, err := suite.client.NATGatewayService.ValidateNATGatewayOrder(ctx, productUID)
	suite.ErrorIs(err, ErrNATGatewayOrderResponseEmpty)
}

func (suite *NATGatewayClientTestSuite) TestValidateNATGatewayOrder_MissingUID() {
	_, err := suite.client.NATGatewayService.ValidateNATGatewayOrder(context.Background(), "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}

func (suite *NATGatewayClientTestSuite) TestBuyNATGateway() {
	ctx := context.Background()
	natSvc := suite.client.NATGatewayService
	productUID := "11111111-2222-3333-4444-555555555555"

	called := false
	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		called = true
		suite.Equal(http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		suite.NoError(err)
		var payload []map[string]string
		suite.NoError(json.Unmarshal(body, &payload))
		suite.Equal([]map[string]string{{"productUid": productUID}}, payload)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"message":"NAT_GATEWAY created",
			"terms":"",
			"data":[{
				"uid":%q,
				"name":"gw-name",
				"serviceName":"gw-name",
				"productType":"NAT_GATEWAY",
				"provisioningStatus":"DEPLOYABLE",
				"rateLimit":1000,
				"aLocationId":10,
				"contractTermMonths":1,
				"createDate":1776431685787
			}]
		}`, productUID)
	})

	result, err := natSvc.BuyNATGateway(ctx, productUID)
	suite.NoError(err)
	suite.True(called)
	suite.Equal(productUID, result.ProductUID)
	suite.Equal("DEPLOYABLE", result.ProvisioningStatus)
	suite.Equal(1000, result.RateLimit)
	suite.Equal(10, result.LocationID)
	suite.Equal(1, result.ContractTermMonths)
}

func (suite *NATGatewayClientTestSuite) TestBuyNATGateway_EmptyData() {
	ctx := context.Background()
	productUID := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","terms":"","data":[]}`)
	})

	_, err := suite.client.NATGatewayService.BuyNATGateway(ctx, productUID)
	suite.ErrorIs(err, ErrNATGatewayOrderResponseEmpty)
}

func (suite *NATGatewayClientTestSuite) TestBuyNATGateway_MissingUID() {
	_, err := suite.client.NATGatewayService.BuyNATGateway(context.Background(), "")
	suite.ErrorIs(err, ErrNATGatewayProductUIDRequired)
}
