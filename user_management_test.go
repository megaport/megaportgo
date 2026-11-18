package megaport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// UserManagementClientTestSuite tests the user management service client
type UserManagementClientTestSuite struct {
	ClientTestSuite
}

func TestUserManagementClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(UserManagementClientTestSuite))
}

func (suite *UserManagementClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *UserManagementClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestCreateUser tests the CreateUser method
func (suite *UserManagementClientTestSuite) TestCreateUser() {
	ctx := context.Background()
	createReq := &CreateUserRequest{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Phone:     "+14155552671",
		Active:    true,
		Position:  USER_POSITION_COMPANY_ADMIN,
	}

	jblob := `{
        "message": "User created successfully",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
            "companyId": 1234,
            "employmentId": 5678,
            "employeeId": 9012
        }
    }`

	suite.mux.HandleFunc("/v2/employment", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodPost)
		w.WriteHeader(http.StatusCreated) // Return 201 Created as expected
		fmt.Fprint(w, jblob)
	})

	resp, err := suite.client.UserManagementService.CreateUser(ctx, createReq)
	suite.NoError(err)
	suite.Equal(1234, resp.CompanyID)
	suite.Equal(5678, resp.EmploymentID)
	suite.Equal(9012, resp.EmployeeID)
}

// TestGetUser tests the GetUser method
func (suite *UserManagementClientTestSuite) TestGetUser() {
	ctx := context.Background()
	employeeID := 9012

	jblob := `{
        "message": "User retrieved successfully",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": {
            "partyId": 9012,
            "firstName": "John",
            "lastName": "Doe",
            "email": "john.doe@example.com",
            "phone": "+14155552671",
            "position": "Company Admin",
            "active": true
        }
    }`

	suite.mux.HandleFunc(fmt.Sprintf("/v2/employee/%d", employeeID), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	user, err := suite.client.UserManagementService.GetUser(ctx, employeeID)
	suite.NoError(err)
	suite.Equal("John", user.FirstName)
	suite.Equal("Doe", user.LastName)
	suite.Equal("john.doe@example.com", user.Email)
	suite.Equal(9012, user.PartyId)
	suite.Equal(true, user.Active)
}

// TestListCompanyUsers tests the ListCompanyUsers method
func (suite *UserManagementClientTestSuite) TestListCompanyUsers() {
	ctx := context.Background()

	jblob := `{
        "message": "Users retrieved successfully",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": [
            {
                "partyId": 9012,
                "firstName": "John",
                "lastName": "Doe",
                "email": "john.doe@example.com",
                "position": "Company Admin",
                "active": true
            },
            {
                "partyId": 9013,
                "firstName": "Jane",
                "lastName": "Smith",
                "email": "jane.smith@example.com",
                "position": "Technical Admin",
                "active": true
            }
        ]
    }`

	suite.mux.HandleFunc("/v2/employment", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	users, err := suite.client.UserManagementService.ListCompanyUsers(ctx)
	suite.NoError(err)
	suite.Equal(2, len(users))
	suite.Equal("John", users[0].FirstName)
	suite.Equal("Jane", users[1].FirstName)
}

// TestUpdateUser tests the UpdateUser method
func (suite *UserManagementClientTestSuite) TestUpdateUser() {
	ctx := context.Background()
	employeeID := 9012
	active := false
	position := string(USER_POSITION_TECHNICAL_ADMIN)
	firstName := "Johnny"

	updateReq := &UpdateUserRequest{
		Active:    &active,
		Position:  &position,
		FirstName: &firstName,
	}

	// Handle both GET (to fetch existing user) and PUT (to update) requests
	suite.mux.HandleFunc(fmt.Sprintf("/v2/employee/%d", employeeID), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Return existing user data for the initial GET request
			fmt.Fprint(w, `{
				"message": "User retrieved successfully",
				"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
				"data": {
					"partyId": 9012,
					"firstName": "John",
					"lastName": "Doe",
					"email": "john.doe@example.com",
					"position": "Company Admin",
					"active": true,
					"confirmationPending": false
				}
			}`)
		case http.MethodPut:
			// Handle the actual update request
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"message": "User updated successfully"}`)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	err := suite.client.UserManagementService.UpdateUser(ctx, employeeID, updateReq)
	suite.NoError(err)
}

// TestUserPhoneCheckedByAPI verifies CreateUser and UpdateUser send the phone as given and return the API's error.
func (suite *UserManagementClientTestSuite) TestUserPhoneCheckedByAPI() {
	ctx := context.Background()
	employeeID := 9012
	var sent struct {
		Phone string `json:"phone"`
	}
	suite.mux.HandleFunc("/v2/employment", func(w http.ResponseWriter, r *http.Request) {
		suite.NoError(json.NewDecoder(r.Body).Decode(&sent))
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message": "Error Creating User - The following issues prevented this user from being created: phone number is not valid. Please check the information provided and try again. If the problem persists, contact support@megaport.com."}`)
	})
	suite.mux.HandleFunc(fmt.Sprintf("/v2/employee/%d", employeeID), func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			suite.NoError(json.NewDecoder(r.Body).Decode(&sent))
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"message": "User's phone is not in international format."}`)
			return
		}
		fmt.Fprint(w, `{"data": {"partyId": 9012, "firstName": "John", "lastName": "Doe", "email": "john.doe@example.com", "position": "Company Admin", "active": true}}`)
	})

	for _, phone := range []string{"+1 415 555 2671", "(415) 555-2671"} {
		suite.Run(phone, func() {
			sent.Phone = ""
			_, err := suite.client.UserManagementService.CreateUser(ctx, &CreateUserRequest{
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
				Phone:     phone,
				Position:  USER_POSITION_COMPANY_ADMIN,
			})
			suite.ErrorContains(err, "phone number is not valid")
			suite.Equal(phone, sent.Phone)

			sent.Phone = ""
			err = suite.client.UserManagementService.UpdateUser(ctx, employeeID, &UpdateUserRequest{Phone: PtrTo(phone)})
			suite.ErrorContains(err, "User's phone is not in international format.")
			suite.Equal(phone, sent.Phone)
		})
	}
}

// TestDeleteUser tests the DeleteUser method
func (suite *UserManagementClientTestSuite) TestDeleteUser() {
	ctx := context.Background()
	employeeID := 9012

	suite.mux.HandleFunc(fmt.Sprintf("/v2/employee/%d", employeeID), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// DeleteUser first calls GetUser to check if user has logged in
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{
				"message": "success",
				"terms": "",
				"data": {
					"partyId": 9012,
					"firstName": "Test",
					"lastName": "User",
					"email": "test@example.com",
					"invitationPending": true
				}
			}`)
		case http.MethodDelete:
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"message": "User deleted successfully"}`)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	err := suite.client.UserManagementService.DeleteUser(ctx, employeeID)
	suite.NoError(err)
}

// TestGetUserActivity tests the GetUserActivity method
func (suite *UserManagementClientTestSuite) TestGetUserActivity() {
	ctx := context.Background()
	req := &GetUserActivityRequest{
		PersonIdOrUid:  "9012",
		CompanyIdOrUid: "1234",
	}

	jblob := `{
        "data": {
            "items": [
                {
                    "loginName": "John Doe",
                    "personId": 9012,
                    "description": "User logged in",
                    "name": "Login",
                    "createdAt": "2025-11-12T10:15:30.123Z",
                    "userType": "USER"
                },
                {
                    "loginName": "John Doe",
                    "personId": 9012,
                    "description": "User updated profile",
                    "name": "UpdateProfile",
                    "createdAt": "2025-11-12T10:20:00.000Z",
                    "userType": "USER"
                }
            ],
            "pageNumber": 1,
            "pageSize": 10,
            "totalItems": 2,
            "totalPages": 1
        }
    }`

	suite.mux.HandleFunc("/v4/activity", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Empty(r.URL.RawQuery)
		fmt.Fprint(w, jblob)
	})

	activities, err := suite.client.UserManagementService.GetUserActivity(ctx, req)
	suite.NoError(err)
	suite.Equal(2, len(activities))
	suite.Equal("John Doe", activities[0].LoginName)
	suite.Equal(9012, activities[0].PersonId)
	suite.Equal("User logged in", activities[0].Description)
	suite.Equal("Login", activities[0].Name)
	suite.Equal("USER", activities[0].UserType)
	suite.True(time.Date(2025, 11, 12, 10, 15, 30, 123000000, time.UTC).Equal(activities[0].CreateDate.Time))
}

// TestUserManagementNilRequestGuards verifies required-request methods reject a nil request.
func (suite *UserManagementClientTestSuite) TestUserManagementNilRequestGuards() {
	ctx := context.Background()
	tests := []struct {
		name    string
		call    func() error
		wantErr error
	}{
		{
			name:    "CreateUser",
			call:    func() error { _, err := suite.client.UserManagementService.CreateUser(ctx, nil); return err },
			wantErr: ErrCreateUserRequestNil,
		},
		{
			name:    "UpdateUser",
			call:    func() error { return suite.client.UserManagementService.UpdateUser(ctx, 9012, nil) },
			wantErr: ErrUpdateUserRequestNil,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.ErrorIs(tt.call(), tt.wantErr)
		})
	}
}

// TestGetUserActivityNilRequest verifies a nil request and an empty page return no entries.
func (suite *UserManagementClientTestSuite) TestGetUserActivityNilRequest() {
	ctx := context.Background()
	suite.mux.HandleFunc("/v4/activity", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Empty(r.URL.RawQuery)
		fmt.Fprint(w, `{"data": {"items": [], "pageNumber": 1, "pageSize": 10, "totalItems": 0, "totalPages": 0}}`)
	})

	activities, err := suite.client.UserManagementService.GetUserActivity(ctx, nil)
	suite.NoError(err)
	suite.Empty(activities)
}

// TestGetUserActivityFilters verifies the request fields reach the /v4/activity query string.
func (suite *UserManagementClientTestSuite) TestGetUserActivityFilters() {
	ctx := context.Background()
	aest := time.FixedZone("AEST", 10*60*60)
	from := time.Date(2026, 11, 11, 9, 30, 0, 123000000, aest)
	to := time.Date(2026, 11, 18, 10, 0, 0, 456000000, aest)
	tests := []struct {
		name string
		req  *GetUserActivityRequest
		want url.Values
	}{
		{
			name: "activity names",
			req:  &GetUserActivityRequest{ActivityNames: []string{"Login", "New API Order"}},
			want: url.Values{"activity_name": {"Login", "New API Order"}},
		},
		{
			name: "date range in UTC",
			req:  &GetUserActivityRequest{From: &from, To: &to},
			want: url.Values{"from": {"2026-11-10T23:30:00.123Z"}, "to": {"2026-11-18T00:00:00.456Z"}},
		},
		{
			name: "page size and sort",
			req:  &GetUserActivityRequest{PageSize: PtrTo(50), Sort: PtrTo("name"), Direction: PtrTo("DESC")},
			want: url.Values{"page_size": {"50"}, "sort": {"name,desc"}},
		},
		{
			name: "sort with no direction",
			req:  &GetUserActivityRequest{Sort: PtrTo("description")},
			want: url.Values{"sort": {"description"}},
		},
		{
			name: "direction with no sort",
			req:  &GetUserActivityRequest{Direction: PtrTo("ASC")},
			want: url.Values{"sort": {"createdAt,asc"}},
		},
		{
			name: "empty request",
			req:  &GetUserActivityRequest{},
			want: url.Values{},
		},
	}
	var got url.Values
	var uri string
	suite.mux.HandleFunc("/v4/activity", func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Query()
		uri = r.RequestURI
		fmt.Fprint(w, `{"data": {"items": []}}`)
	})
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			_, err := suite.client.UserManagementService.GetUserActivity(ctx, tt.req)
			suite.NoError(err)
			suite.Equal(tt.want, got)
			if len(tt.want) == 0 {
				suite.Equal("/v4/activity", uri)
			}
		})
	}
}
