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

// ModifyProduct(ctx context.Context, req *ModifyProductRequest) (*ModifyProductResponse, error)
// DeleteProduct(ctx context.Context, req *DeleteProductRequest) (*DeleteProductResponse, error)
// RestoreProduct(ctx context.Context, req *RestoreProductRequest) (*RestoreProductResponse, error)
// ManageProductLock(ctx context.Context, req *ManageProductLockRequest) (*ManageProductLockResponse, error)

func TestExecuteOrder(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()
	productSvc := client.ProductService

	jblob := `{
			"message": "test-message",
			"terms": "test-terms",
			"data": [
			{"technicalServiceUid": "36b3f68e-2f54-4331-bf94-f8984449365f"}
		]
	}`

	portOrder := []types.PortOrder{
		{
			Name:                  "test-port",
			Term:                  12,
			PortSpeed:             10000,
			LocationID:            226,
			Virtual:               false,
			Market:                "US",
			MarketplaceVisibility: false,
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
	wantRes := PtrTo([]byte(jblob))
	gotRes, err := productSvc.ExecuteOrder(ctx, portOrder)
	require.NoError(t, err)
	require.Equal(t, wantRes, gotRes)
}

func TestModifyProduct(t *testing.T) {
	setup()
	defer teardown()

	ctx := context.Background()
	productSvc := client.ProductService
	productUid := "36b3f68e-2f54-4331-bf94-f8984449365f"
	jblob := `{
    "message": "Product [36b3f68e-2f54-4331-bf94-f8984449365f] has been updated",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": {
        "serviceName": "updated-test-product",
        "name": "updated-test-product",
        "secondaryName": null,
        "technicalServiceId": 185927,
        "technicalServiceUid": "ef60d544-00e1-4ccc-bcff-3e2050bface5",
        "requestedDate": 1706202200307,
        "configuredDate": null,
        "currentEstimatedDelivery": null,
        "companyName": "test-company",
        "companyId": 1153,
        "billingContactName": null,
        "billingContactId": null,
        "adminContactName": null,
        "adminContactId": null,
        "technicalContactName": null,
        "technicalContactId": null,
        "salesName": null,
        "salesId": null,
        "billableId": 177726,
        "billableUsageAlgorithm": null,
        "productType": "MEGAPORT",
        "provisioningStatus": "DEPLOYABLE",
        "failedReason": null,
        "inAdvanceBillingStatus": null,
        "provisioningItems": [],
        "tags": [],
        "vxcDistanceBand": null,
        "intercapPath": null,
        "marketplaceVisibility": false,
        "vxcPermitted": true,
        "vxcAutoApproval": false,
        "createDate": 1706202200307,
        "terminationDate": null,
        "contractStartDate": null,
        "contractTermMonths": 1,
        "rateType": "MONTHLY",
        "trialAgreement": false,
        "payerCompanyId": null,
        "nonPayerCompanyId": null,
        "minimumSpeed": null,
        "maximumSpeed": null,
        "rateLimit": 10000,
        "errorMessage": null,
        "lagId": null,
        "aggregationId": null,
        "lagPrimary": null,
        "market": "USA",
        "accountManager": null,
        "promptUid": null,
        "components": [],
        "attributes": [],
        "aLocation": null,
        "bLocation": null,
        "aMetro": null,
        "aCountry": null,
        "aLocationId": null,
        "bLocationId": null,
        "bMetro": null,
        "bCountry": null,
        "attributeTags": {},
        "createdBy": "32df7107-fdca-4c2a-8ccb-c6867813b3f2",
        "buyoutPort": false,
        "virtual": false,
        "locked": false,
        "adminLocked": false,
        "bgpShutdownDefault": false,
        "originDomain": null
    	}
	}`
	productType := types.PRODUCT_MEGAPORT
	wantReq := &ModifyProductRequest{
		ProductID:             productUid,
		ProductType:           productType,
		Name:                  "updated-test-product",
		CostCentre:            "US",
		MarketplaceVisibility: false,
	}
	wantUpdate := &types.ProductUpdate{
		Name:                 wantReq.Name,
		CostCentre:           wantReq.CostCentre,
		MarketplaceVisbility: wantReq.MarketplaceVisibility,
	}
	path := fmt.Sprintf("/v2/product/%s/%s", productType, productUid)
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		v := new(types.ProductUpdate)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, jblob)
		require.Equal(t, wantUpdate, v)
	})
	wantRes := &ModifyProductResponse{
		IsUpdated: true,
	}
	gotRes, err := productSvc.ModifyProduct(ctx, wantReq)
	require.NoError(t, err)
	require.Equal(t, wantRes, gotRes)
}

func TestDeleteProduct(t *testing.T) {

}

func RestoreProduct(t *testing.T) {

}

func ManageProductLuck(t *testing.T) {

}
