package megaport

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// ServiceKeyClientTestSuite tests the service key service client
type ServiceKeyClientTestSuite struct {
	ClientTestSuite
}

func TestServiceClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PortClientTestSuite))
}

func (suite *ServiceKeyClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *ServiceKeyClientTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *ServiceKeyClientTestSuite) TestCreateServiceKey() {
	createReq := &CreateServiceKeyRequest{
		ProductUID:  "6ed7270f-b3d9-4539-a036-1639f8a398e5",
		Description: "Test Service Key",
		Active:      true,
		MaxSpeed:    500,
		PreApproved: true,
		VLAN:        3,
	}

	validFor := &ValidFor{}
	validFor.StartTime = &Time{
		Time: time.Now(),
	}
	validFor.EndTime = &Time{
		Time: time.Now().Add(time.Hour * 24),
	}
	createReq.ValidFor = validFor

	jblob := `{
        "message": "New service key [e19de6a6-5354-4240-b382-70aa4352dc20] generated",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
          "key": "e19de6a6-5354-4240-b382-70aa4352dc20",
          "createDate": 1608587790497,
          "companyId": 1153,
          "companyUid": "160208ae-01e4-4cb9-8d57-03a197bd47a8",
          "companyName": "Megaport Lab",
          "description": "Documentation Service Key - Single Use",
          "productId": 74177,
          "productUid": "6ed7270f-b3d9-4539-a036-1639f8a398e5",
          "productName": "My LA API Port",
          "productDto": {
            "productId": 74177,
            "productUid": "6ed7270f-b3d9-4539-a036-1639f8a398e5",
            "productName": "My LA API Port",
            "productType": "MEGAPORT",
            "provisioningStatus": "CONFIGURED",
            "createDate": 1608587655803,
            "createdBy": "52a26471-94de-4d32-b89e-0cd5d26c1f65",
            "portSpeed": 1000,
            "terminateDate": null,
            "liveDate": null,
            "market": "US",
            "locationId": 60,
            "usageAlgorithm": "NOT_POST_PAID",
            "marketplaceVisibility": true,
            "vxcpermitted": true,
            "vxcAutoApproval": false,
            "secondaryName": null,
            "lagPrimary": false,
            "lagId": null,
            "aggregationId": null,
            "companyUid": "160208ae-01e4-4cb9-8d57-03a197bd47a8",
            "companyName": "Megaport Lab",
            "contractStartDate": 1608559200000,
            "contractEndDate": 1611237600000,
            "contractTermMonths": 1,
            "associatedVxcs": [],
            "associatedIxs": [],
            "attributeTags": {},
            "virtual": false,
            "buyoutPort": false,
            "locked": false,
            "adminLocked": false,
            "cancelable": true
          },
          "vlan": 3,
          "maxSpeed": 500,
          "preApproved": true,
          "singleUse": true,
          "lastUsed": null,
          "active": true,
          "validFor": {
            "start": 1608506197135,
            "end": 1612015200000
          },
          "expired": false,
          "valid": true,
          "promoCode": null
        }
      }`

	suite.mux.HandleFunc("/v2/service/key", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})

	_, err := suite.client.ServiceKeyService.CreateServiceKey(ctx, createReq)
	suite.NoError(err)
}

func (suite *ServiceKeyClientTestSuite) TestListServiceKeys() {
	productUid := "69ddf381-06ae-4150-9690-a46c8323d2d5"
	listReq := &ListServiceKeysRequest{
		ProductUID: &productUid,
	}

	jblob := `{
        "message": "Found [1] service keys for company [1153]",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": [
          {
            "key": "d1269fa6-dde7-49d2-b643-0a9d3cc7bedd",
            "createDate": 1527815582306,
            "companyId": 1153,
            "companyUid": "160208ae-01e4-4cb9-8d57-03a197be47a8",
            "companyName": "Megaport Lab",
            "description": "pmgsk",
            "productId": 19347,
            "productUid": "69ddf381-06ae-4150-9690-a46c8323d2d5",
            "productName": "PMG1",
            "vlan": 300,
            "maxSpeed": 10,
            "preApproved": null,
            "singleUse": true,
            "lastUsed": null,
            "active": true,
            "validFor": {
              "start": 1527815662405,
              "end": 1528725600000
            },
            "expired": true,
            "valid": false,
            "promoCode": null
          }
        ]
      }`
	want := &ListServiceKeysResponse{ServiceKeys: []*ServiceKey{
		{
			Key: "d1269fa6-dde7-49d2-b643-0a9d3cc7bedd",
			// CreateDate:  &Time{Time: GetTime(1527815582306)},
			// CompanyID:   1153,
			// CompanyUID:  "160208ae-01e4-4cb9-8d57-03a197be47a8",
			// CompanyName: "Megaport Lab",
			// Description: "pmgsk",
			// ProductID:   19347,
			// ProductUID:  "69ddf381-06ae-4150-9690-a46c8323d2d5",
			// ProductName: "PMG1",
			// VLAN:        300,
			// MaxSpeed:    10,
			// PreApproved: false,
			// SingleUse:   true,
			// LastUsed:    nil,
			// Active:      true,
			// ValidFor: &ValidFor{
			// 	StartTime: &Time{
			// 		Time: GetTime(1527815662405),
			// 	},
			// 	EndTime: &Time{
			// 		Time: GetTime(1528725600000),
			// 	},
			// },
		},
	}}
	suite.mux.HandleFunc(fmt.Sprintf("/v2/service/key?productidOrUid=%s", productUid), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	listRes, err := suite.client.ServiceKeyService.ListServiceKeys(ctx, listReq)
	suite.NoError(err)
	suite.Equal(want, listRes)
}

func (suite *ServiceKeyClientTestSuite) TestUpdateServiceKey() {
	updateReq := &UpdateServiceKeyRequest{
		Key:       "e19de6a6-5354-4240-b382-70aa4352dc20",
		ProductID: 74177,
		SingleUse: false,
		Active:    true,
		ValidFor: &ValidFor{
			StartTime: &Time{
				Time: GetTime(1608506197135),
			},
			EndTime: &Time{
				Time: GetTime(1612015200000),
			},
		},
	}
	jblob := `{
        "message": "Service key [e19de6a6-5354-4240-b382-70aa4352dc20] updated",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
          "key": "e19de6a6-5354-4240-b382-70aa4352dc20",
          "createDate": 1608587790497,
          "companyId": 1153,
          "companyUid": "160208ae-01e4-4cb9-8d57-03a197bd47a8",
          "companyName": "Megaport Lab",
          "description": null,
          "productId": 74177,
          "productUid": "6ed7270f-b3d9-4539-a036-1639f8a398e5",
          "productName": "My LA API Port",
          "productDto": {
            "productId": 74177,
            "productUid": "6ed7270f-b3d9-4539-a036-1639f8a398e5",
            "productName": "My LA API Port",
            "productType": "MEGAPORT",
            "provisioningStatus": "LIVE",
            "createDate": 1608587655803,
            "createdBy": "52a26471-94de-4d32-b89e-0cd5d26c1f65",
            "portSpeed": 1000,
            "terminateDate": null,
            "liveDate": 1608588001277,
            "market": "US",
            "locationId": 60,
            "usageAlgorithm": "NOT_POST_PAID",
            "marketplaceVisibility": true,
            "vxcpermitted": true,
            "vxcAutoApproval": false,
            "secondaryName": null,
            "lagPrimary": false,
            "lagId": null,
            "aggregationId": null,
            "companyUid": "160208ae-01e4-4cb9-8d57-03a197bd47a8",
            "companyName": "Megaport Lab",
            "contractStartDate": 1608588001105,
            "contractEndDate": 1611237600000,
            "contractTermMonths": 1,
            "associatedVxcs": [],
            "associatedIxs": [],
            "attributeTags": {},
            "virtual": false,
            "buyoutPort": false,
            "locked": false,
            "adminLocked": false,
            "cancelable": true
          },
          "vlan": null,
          "maxSpeed": 1000,
          "preApproved": null,
          "singleUse": false,
          "lastUsed": null,
          "active": true,
          "validFor": {
            "start": 1608506197135,
            "end": 1612015200000
          },
          "expired": false,
          "valid": true,
          "promoCode": null
        }
      }`
	suite.mux.HandleFunc("/v2/service/key", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPut)
		fmt.Fprint(w, jblob)
	})
	_, err := suite.client.ServiceKeyService.UpdateServiceKey(ctx, updateReq)
	suite.NoError(err)
}
