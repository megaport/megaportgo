package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/require"
)

func TestGetPort(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	portSvc := client.PortService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"

	want := &types.Port{
		ID:                    999999,
		UID:                   productUid,
		Name:                  "test-port",
		SecondaryName:         "test-secondary-name",
		ProvisioningStatus:    "CONFIGURED",
		PortSpeed:             10000,
		LocationID:            226,
		LAGPrimary:            false,
		Market:                "US",
		MarketplaceVisibility: false,
		VXCPermitted:          true,
		VXCAutoApproval:       false,
		Virtual:               false,
		BuyoutPort:            false,
		CompanyName:           "test-company",
		CompanyUID:            companyUid,
		ContractStartDate:     1706104800000,
		ContractEndDate:       1737727200000,
		ContractTermMonths:    12,
		Locked:                false,
		AdminLocked:           false,
		Cancelable:            true,
	}
	mux.HandleFunc(fmt.Sprintf("/v2/product/%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"message": "test-message",
			"terms": "test-terms",
			"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}
			}
			}`)
	})
	got, err := portSvc.GetPort(ctx, &GetPortRequest{
		PortID: productUid,
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestBuyPort(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	portSvc := client.PortService

	want := &types.PortOrderConfirmation{
		TechnicalServiceUID: "36b3f68e-2f54-4331-bf94-f8984449365f",
	}
	req := &BuyPortRequest{
		Name:       "test-port",
		Term:       12,
		PortSpeed:  10000,
		LocationId: 226,
		Market:     "US",
		IsLag:      false,
		LagCount:   0,
		IsPrivate:  true,
	}

	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
			]
			}`
	portOrder := []types.PortOrder{
		{
			Name:                  req.Name,
			Term:                  req.Term,
			PortSpeed:             req.PortSpeed,
			LocationID:            req.LocationId,
			Virtual:               false,
			Market:                req.Market,
			MarketplaceVisibility: !req.IsPrivate,
		},
	}
	mux.HandleFunc("/v3/networkdesign/buy", func(w http.ResponseWriter, r *http.Request) {
		v := new([]types.PortOrder)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}
		orders := *v
		wantOrder := portOrder[0]
		gotOrder := orders[0]
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, jblob)
		require.Equal(t, wantOrder.Name, gotOrder.Name)
		require.Equal(t, wantOrder.Term, gotOrder.Term)
		require.Equal(t, wantOrder.PortSpeed, gotOrder.PortSpeed)
		require.Equal(t, wantOrder.LocationID, gotOrder.LocationID)
		require.Equal(t, wantOrder.Virtual, gotOrder.Virtual)
		require.Equal(t, wantOrder.MarketplaceVisibility, gotOrder.MarketplaceVisibility)
	})
	got, err := portSvc.BuyPort(ctx, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}
