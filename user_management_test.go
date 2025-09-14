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

// UserManagementClientTestSuite tests the service key service client
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

// TestDeleteUser tests the DeleteUser method
func (suite *UserManagementClientTestSuite) TestDeleteUser() {
	ctx := context.Background()
	employeeID := 9012

	suite.mux.HandleFunc(fmt.Sprintf("/v2/employee/%d", employeeID), func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodDelete)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message": "User deleted successfully"}`)
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

	jblob := `[
        {
            "loginName": "John Doe",
            "personId": 9012,
            "description": "User logged in",
            "name": "Login",
            "createDate": 1623456789000,
            "userType": "USER"
        },
        {
            "loginName": "John Doe",
            "personId": 9012,
            "description": "User updated profile",
            "name": "UpdateProfile",
            "createDate": 1623456999000,
            "userType": "USER"
        }
    ]`

	suite.mux.HandleFunc("/v3/activity", func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		suite.Equal("9012", r.URL.Query().Get("personIdOrUid"))
		suite.Equal("1234", r.URL.Query().Get("companyIdOrUid"))
		fmt.Fprint(w, jblob)
	})

	activities, err := suite.client.UserManagementService.GetUserActivity(ctx, req)
	suite.NoError(err)
	suite.Equal(2, len(activities))
	suite.Equal("John Doe", activities[0].LoginName)
	suite.Equal(9012, activities[0].PersonId)
	suite.Equal("User logged in", activities[0].Description)
	suite.Equal("Login", activities[0].Name)
}
