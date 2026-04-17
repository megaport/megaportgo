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

	"github.com/stretchr/testify/suite"
)

// OrderServiceTestSuite tests the Order service.
type OrderServiceTestSuite struct {
	ClientTestSuite
}

func TestOrderServiceTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(OrderServiceTestSuite))
}

func (suite *OrderServiceTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	u, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = u
}

func (suite *OrderServiceTestSuite) TearDownTest() {
	suite.server.Close()
}

const orderResponseJSON = `{
	"message": "Data returned successfully",
	"terms": "This data is subject to the Acceptable Use Policy",
	"data": {
		"companyUid": "e900d0d5-1030-4e29-b2d8-816ad4263190",
		"createdBy": "user-name",
		"items": ["e900d0d5-1030-4e29-b2d8-816ad4263190"],
		"reference": "ORD-123456",
		"state": "DESIGN",
		"uid": "11111111-2222-3333-4444-555555555555"
	}
}`

func (suite *OrderServiceTestSuite) TestCreateOrder() {
	ctx := context.Background()

	suite.mux.HandleFunc("/v3/orders", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)

		body, err := io.ReadAll(r.Body)
		suite.NoError(err)
		var payload CreateOrderRequest
		suite.NoError(json.Unmarshal(body, &payload))
		suite.Equal([]string{"nat-gw-uid"}, payload.Items)
		suite.Equal("ORD-TEST", payload.Reference)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, orderResponseJSON)
	})

	order, err := suite.client.OrderService.CreateOrder(ctx, &CreateOrderRequest{
		Items:     []string{"nat-gw-uid"},
		Reference: "ORD-TEST",
	})
	suite.NoError(err)
	suite.Equal("11111111-2222-3333-4444-555555555555", order.UID)
	suite.Equal("DESIGN", order.State)
}

func (suite *OrderServiceTestSuite) TestCreateOrder_MissingItems() {
	_, err := suite.client.OrderService.CreateOrder(context.Background(), &CreateOrderRequest{})
	suite.ErrorIs(err, ErrOrderItemsRequired)
}

func (suite *OrderServiceTestSuite) TestGetOrder() {
	ctx := context.Background()
	uid := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/orders/"+uid, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, orderResponseJSON)
	})

	order, err := suite.client.OrderService.GetOrder(ctx, uid)
	suite.NoError(err)
	suite.Equal(uid, order.UID)
	suite.Equal("DESIGN", order.State)
}

func (suite *OrderServiceTestSuite) TestGetOrder_MissingUID() {
	_, err := suite.client.OrderService.GetOrder(context.Background(), "")
	suite.ErrorIs(err, ErrOrderUIDRequired)
}

func (suite *OrderServiceTestSuite) TestUpdateOrder() {
	ctx := context.Background()
	uid := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/orders/"+uid, func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPut, r.Method)

		body, err := io.ReadAll(r.Body)
		suite.NoError(err)
		var payload UpdateOrderRequest
		suite.NoError(json.Unmarshal(body, &payload))
		suite.Equal("ORD-UPDATED", payload.Reference)

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, orderResponseJSON)
	})

	order, err := suite.client.OrderService.UpdateOrder(ctx, uid, &UpdateOrderRequest{
		Items:     []string{"nat-gw-uid"},
		Reference: "ORD-UPDATED",
	})
	suite.NoError(err)
	suite.Equal(uid, order.UID)
}

func (suite *OrderServiceTestSuite) TestDeleteOrder() {
	ctx := context.Background()
	uid := "11111111-2222-3333-4444-555555555555"

	called := false
	suite.mux.HandleFunc("/v3/orders/"+uid, func(w http.ResponseWriter, r *http.Request) {
		called = true
		suite.Equal(http.MethodDelete, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"Order deleted successfully","terms":""}`)
	})

	err := suite.client.OrderService.DeleteOrder(ctx, uid)
	suite.NoError(err)
	suite.True(called)
}

func (suite *OrderServiceTestSuite) TestDeleteOrder_MissingUID() {
	err := suite.client.OrderService.DeleteOrder(context.Background(), "")
	suite.ErrorIs(err, ErrOrderUIDRequired)
}

func (suite *OrderServiceTestSuite) TestValidateOrder() {
	ctx := context.Background()
	uid := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/orders/"+uid+"/validate", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, orderResponseJSON)
	})

	order, err := suite.client.OrderService.ValidateOrder(ctx, uid)
	suite.NoError(err)
	suite.Equal(uid, order.UID)
}

func (suite *OrderServiceTestSuite) TestBuyOrder() {
	ctx := context.Background()
	uid := "11111111-2222-3333-4444-555555555555"

	suite.mux.HandleFunc("/v3/orders/"+uid+"/buy", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodPost, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, orderResponseJSON)
	})

	order, err := suite.client.OrderService.BuyOrder(ctx, uid)
	suite.NoError(err)
	suite.Equal(uid, order.UID)
}

func (suite *OrderServiceTestSuite) TestBuyOrder_MissingUID() {
	_, err := suite.client.OrderService.BuyOrder(context.Background(), "")
	suite.ErrorIs(err, ErrOrderUIDRequired)
}
