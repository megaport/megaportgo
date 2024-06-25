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

// ManagedAccountClientTestSuite tests the Managed Account service client
type ManagedAccountClientTestSuite struct {
	ClientTestSuite
}

func TestManagedAccountClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(PortClientTestSuite))
}

func (suite *ManagedAccountClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *ManagedAccountClientTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *ManagedAccountClientTestSuite) TestCreateManagedAccount() {
	ctx := context.Background()
	createReq := &ManagedAccountRequest{
		AccountName: "Test Account",
		AccountRef:  "test-account",
	}
	path := "/v2/managedCompanies"
	jblob := `{
        "message": "New managed company has been successfully created.",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
          "accountRef": "555-1212-0317-1967",
          "accountName": "Best Company Ever",
          "companyUid": "fd404dc9-9efd-43c1-9e0b-a58a9d250130"
        }
      }`
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		fmt.Fprint(w, jblob)
	})
	createRes, err := suite.client.ManagedAccountService.CreateManagedAccount(ctx, createReq)
	suite.NoError(err)
	suite.NotNil(createRes)
	want := &ManagedAccount{
		AccountRef:  "555-1212-0317-1967",
		AccountName: "Best Company Ever",
		CompanyUID:  "fd404dc9-9efd-43c1-9e0b-a58a9d250130",
	}
	suite.Equal(want, createRes)
}

func (suite *ManagedAccountClientTestSuite) TestListManagedAccounts() {
	ctx := context.Background()
	path := "/v2/managedCompanies"
	jblob := `{
        "message": "Managed Accounts.",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": [
          {
            "accountRef": "AUS-BNE-4101-KK",
            "accountName": "Demo Company",
            "companyUid": "ccfcf1dc-cf38-4526-9f2f-13d36a2441db"
          },
          {
            "accountRef": "555-1212-0317-1967",
            "accountName": "Best Company Ever",
            "companyUid": "fd404dc9-9efd-43c1-9e0b-a58a9d250130"
          }
        ]
      }`
	want := []*ManagedAccount{
		{
			AccountRef:  "AUS-BNE-4101-KK",
			AccountName: "Demo Company",
			CompanyUID:  "ccfcf1dc-cf38-4526-9f2f-13d36a2441db",
		},
		{
			AccountRef:  "555-1212-0317-1967",
			AccountName: "Best Company Ever",
			CompanyUID:  "fd404dc9-9efd-43c1-9e0b-a58a9d250130",
		},
	}
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	listRes, err := suite.client.ManagedAccountService.ListManagedAccounts(ctx)
	suite.NoError(err)
	suite.NotNil(listRes)
	suite.Equal(want, listRes)
}

func (suite *ManagedAccountClientTestSuite) TestUpdateManagedAccount() {
	ctx := context.Background()
	updateReq := &ManagedAccountRequest{
		AccountName: "Test Account",
		AccountRef:  "test-account",
	}
	path := "/v2/managedCompanies/fd404dc9-9efd-43c1-9e0b-a58a9d250130"
	jblob := `{
        "message": "Managed company has been successfully updated.",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
          "accountRef": "555-1212-0317-1967",
          "accountName": "Best Company Ever",
          "companyUid": "fd404dc9-9efd-43c1-9e0b-a58a9d250130"
        }
      }`
	want := &ManagedAccount{
		AccountRef:  "555-1212-0317-1967",
		AccountName: "Best Company Ever",
		CompanyUID:  "fd404dc9-9efd-43c1-9e0b-a58a9d250130",
	}
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPatch)
		fmt.Fprint(w, jblob)
	})
	updateRes, err := suite.client.ManagedAccountService.UpdateManagedAccount(ctx, "fd404dc9-9efd-43c1-9e0b-a58a9d250130", updateReq)
	suite.NoError(err)
	suite.NotNil(updateRes)
	suite.Equal(want, updateRes)
}
