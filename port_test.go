package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/megaport/megaportgo/mega_err"
	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/require"
)

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

func TestBuySinglePort(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	portSvc := client.PortService

	want := &types.PortOrderConfirmation{
		TechnicalServiceUID: "36b3f68e-2f54-4331-bf94-f8984449365f",
	}
	req := &BuySinglePortRequest{
		Name:       "test-port",
		Term:       12,
		PortSpeed:  10000,
		LocationId: 226,
		Market:     "US",
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
	got, err := portSvc.BuySinglePort(ctx, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestBuyLAGPort(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	portSvc := client.PortService

	want := &types.PortOrderConfirmation{
		TechnicalServiceUID: "36b3f68e-2f54-4331-bf94-f8984449365f",
	}
	req := &BuyLAGPortRequest{
		Name:       "test-port",
		Term:       12,
		PortSpeed:  10000,
		LocationId: 226,
		Market:     "US",
		IsPrivate:  true,
		LagCount:   2,
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
			LagPortCount:          req.LagCount,
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
		require.Equal(t, wantOrder.LagPortCount, gotOrder.LagPortCount)
	})
	got, err := portSvc.BuyLAGPort(ctx, req)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestBuyPortInvalidTerm(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	portSvc := client.PortService

	req := &BuyPortRequest{
		Name:       "test-port-bad-term",
		Term:       37,
		PortSpeed:  10000,
		LocationId: 226,
		Market:     "US",
		IsLag:      false,
		LagCount:   0,
		IsPrivate:  true,
	}
	_, err := portSvc.BuyPort(ctx, req)
	require.Equal(t, errors.New(mega_err.ERR_TERM_NOT_VALID), err)
}

func TestListPorts(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()

	portSvc := client.PortService

	companyUid := "32df7107-fdca-4c2a-8ccb-c6867813b3f2"
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	productUid2 := "9b1c46c7-1e8d-4035-bf38-1bc60d346d57"
	productUid3 := "91ededc2-473f-4a30-ad24-0703c7f35e50"

	want1 := &types.Port{
		ID:                    999999,
		UID:                   productUid,
		Name:                  "test-port",
		Type:                  "MEGAPORT",
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
	want2 := &types.Port{
		ID:                    999998,
		UID:                   productUid2,
		Name:                  "test-port2",
		Type:                  "MEGAPORT",
		SecondaryName:         "test-secondary-name2",
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
	want3 := &types.Port{
		ID:                    999997,
		UID:                   productUid3,
		Name:                  "test-port3",
		SecondaryName:         "test-secondary-name3",
		Type:                  "MEGAPORT",
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
	wantPorts := []*types.Port{want1, want2, want3}
	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [{
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}, {"productId":999998,"productUid":"9b1c46c7-1e8d-4035-bf38-1bc60d346d57","productName":"test-port2","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name2","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}, {"productId":999997,"productUid":"91ededc2-473f-4a30-ad24-0703c7f35e50","productName":"test-port3","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name3","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}}]
	}`
	mux.HandleFunc("/v2/products", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := portSvc.ListPorts(ctx)
	require.NoError(t, err)
	require.Equal(t, wantPorts, got)
}

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
		Type:                  "MEGAPORT",
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

	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": {
			"productId":999999,"productUid":"36b3f68e-2f54-4331-bf94-f8984449365f","productName":"test-port","productType":"MEGAPORT","provisioningStatus":"CONFIGURED","createDate":0,"createdBy":"","portSpeed":10000,"terminateDate":0,"liveDate":0,"market":"US","locationId":226,"usageAlgorithm":"","marketplaceVisibility":false,"vxcpermitted":true,"vxcAutoApproval":false,"secondaryName":"test-secondary-name","lagPrimary":false,"lagId":0,"aggregationId":0,"companyUid":"32df7107-fdca-4c2a-8ccb-c6867813b3f2","companyName":"test-company","contractStartDate":1706104800000,"contractEndDate":1737727200000,"contractTermMonths":12,"attributeTags":null,"virtual":false,"buyoutPort":false,"locked":false,"adminLocked":false,"cancelable":true,"resources":{"interface":{"demarcation":"","description":"","id":0,"loa_template":"","media":"","name":"","port_speed":0,"resource_name":"","resource_type":"","up":0}}
			}
			}`
	mux.HandleFunc(fmt.Sprintf("/v2/product/%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := portSvc.GetPort(ctx, &GetPortRequest{
		PortID: productUid,
	})
	require.NoError(t, err)
	require.Equal(t, want, got)
}
