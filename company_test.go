package megaport

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

// CompanyClientTestSuite tests the Company service.
type CompanyClientTestSuite struct {
	ClientTestSuite
}

func TestCompanyClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(CompanyClientTestSuite))
}

func (suite *CompanyClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *CompanyClientTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *CompanyClientTestSuite) TestGetCurrentCompanyM2M() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal(http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"message": "Found M2M access token",
			"terms": "terms",
			"data": {
				"companyUid": "160208ae-01e4-4cb9-8d57-03a197bd47a8",
				"companyName": "Example Pty Ltd",
				"accountType": "MANAGED_ACCOUNT",
				"parentCompanyUid": "3f2504e0-4f89-11d3-9a0c-0305e82c3301",
				"parentCompanyName": "Example Partner Pty Ltd"
			}
		}`)
	})

	got, err := suite.client.CompanyService.GetCurrentCompany(context.Background())
	suite.NoError(err)
	suite.Equal(&CurrentCompany{
		CompanyUID:        "160208ae-01e4-4cb9-8d57-03a197bd47a8",
		CompanyName:       "Example Pty Ltd",
		AccountType:       AccountTypeManagedAccount,
		ParentCompanyUID:  "3f2504e0-4f89-11d3-9a0c-0305e82c3301",
		ParentCompanyName: "Example Partner Pty Ltd",
	}, got)
}

// A company with no parent omits the parent fields entirely.
func (suite *CompanyClientTestSuite) TestGetCurrentCompanyNoParent() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","data":{"companyUid":"uid-1","companyName":"Co","accountType":"ADMIN"}}`)
	})

	got, err := suite.client.CompanyService.GetCurrentCompany(context.Background())
	suite.NoError(err)
	suite.Equal("uid-1", got.CompanyUID)
	suite.Equal(AccountTypeAdmin, got.AccountType)
	suite.Empty(got.ParentCompanyUID)
	suite.Empty(got.ParentCompanyName)
}

func (suite *CompanyClientTestSuite) TestGetCurrentCompanySessionUser() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"message": "Found User",
			"data": {
				"username": "someone@example.com",
				"companyUid": "uid-1",
				"companyId": 1153,
				"companyName": "Co",
				"accountType": "DIRECT_CUSTOMER",
				"rootAccountType": "DIRECT_CUSTOMER",
				"isManagedContext": true,
				"permissions": {"uid-1": ["place_order"]},
				"authorities": {"uid-1": {"place_order": true}},
				"loginDate": 1792165887000,
				"companies": [
					{"companyUid": "uid-1", "name": "Co", "accountType": "DIRECT_CUSTOMER"},
					{"companyUid": "uid-2", "name": "Other", "accountType": "PARTNER"}
				]
			}
		}`)
	})

	got, err := suite.client.CompanyService.GetCurrentCompany(context.Background())
	suite.NoError(err)
	suite.Equal(&CurrentCompany{
		CompanyUID:       "uid-1",
		CompanyID:        1153,
		CompanyName:      "Co",
		AccountType:      AccountTypeDirectCustomer,
		RootAccountType:  AccountTypeDirectCustomer,
		IsManagedContext: true,
		Companies: []MemberCompany{
			{CompanyUID: "uid-1", Name: "Co", AccountType: AccountTypeDirectCustomer},
			{CompanyUID: "uid-2", Name: "Other", AccountType: AccountTypePartner},
		},
	}, got)
}

func (suite *CompanyClientTestSuite) TestGetCurrentCompanyUnknownAccountType() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","data":{"companyUid":"uid-1","companyName":"Co","accountType":"SOMETHING_NEW"}}`)
	})

	got, err := suite.client.CompanyService.GetCurrentCompany(context.Background())
	suite.NoError(err)
	suite.Equal(AccountType("SOMETHING_NEW"), got.AccountType)
}

func (suite *CompanyClientTestSuite) TestGetCurrentCompanyNullData() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","data":null}`)
	})

	got, err := suite.client.CompanyService.GetCurrentCompany(context.Background())
	suite.Nil(got)
	suite.ErrorIs(err, ErrCurrentCompanyResponseEmpty)
}

func (suite *CompanyClientTestSuite) TestGetCurrentCompanyMissingCompanyUID() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","data":{"companyName":"Co"}}`)
	})

	got, err := suite.client.CompanyService.GetCurrentCompany(context.Background())
	suite.Nil(got)
	suite.ErrorIs(err, ErrCurrentCompanyResponseEmpty)
}

func (suite *CompanyClientTestSuite) TestGetCurrentCompanyUnauthorized() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"message":"Unauthorized"}`)
	})

	got, err := suite.client.CompanyService.GetCurrentCompany(context.Background())
	suite.Nil(got)
	suite.Error(err)

	var errResp *ErrorResponse
	suite.True(errors.As(err, &errResp))
	suite.Equal(http.StatusUnauthorized, errResp.Response.StatusCode)
}

func (suite *CompanyClientTestSuite) TestGetCurrentCompanyDecodeError() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"test","data":{`)
	})

	got, err := suite.client.CompanyService.GetCurrentCompany(context.Background())
	suite.Nil(got)
	suite.Error(err)
}

func (suite *CompanyClientTestSuite) TestGetCurrentCompanySendsCallContext() {
	suite.mux.HandleFunc("/v2/me", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal("managed-uid", r.Header.Get("X-Call-Context"))
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"message":"ok","data":{"companyUid":"managed-uid","companyName":"Managed"}}`)
	})

	c, err := New(nil, WithBaseURL(suite.server.URL), WithCallContext("managed-uid"))
	suite.NoError(err)

	got, err := c.CompanyService.GetCurrentCompany(context.Background())
	suite.NoError(err)
	suite.Equal("managed-uid", got.CompanyUID)
}

func (suite *CompanyClientTestSuite) TestCompanyServiceIsWired() {
	c, err := New(nil)
	suite.NoError(err)
	suite.NotNil(c.CompanyService)
}
