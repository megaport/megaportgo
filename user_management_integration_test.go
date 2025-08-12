package megaport

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

// UserManagementIntegrationTestSuite tests the User Management Service.
type UserManagementIntegrationTestSuite IntegrationTestSuite

func TestUserManagementIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(UserManagementIntegrationTestSuite))
	}
}

func (suite *UserManagementIntegrationTestSuite) SetupSuite() {
	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	megaportClient, err := New(nil, WithBaseURL(MEGAPORTURL), WithLogHandler(handler), WithCredentials(accessKey, secretKey))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	_, err = megaportClient.Authorize(ctx)
	if err != nil {
		suite.FailNowf("", "could not authorize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

// TestUserCRUD tests the full lifecycle of user management: Create, Read, Update, Delete
func (suite *UserManagementIntegrationTestSuite) TestUserCRUD() {
	ctx := context.Background()

	// Get initial users list to verify new user is actually created
	usersListInitial, err := suite.client.UserManagementService.ListCompanyUsers(ctx)
	suite.NoError(err)

	// Create user
	createdUser, err := suite.testCreateUser(suite.client, ctx)
	suite.NoError(err)
	suite.NotNil(createdUser)

	employeeID := createdUser.EmployeeID
	suite.True(employeeID > 0, "Employee ID should be greater than 0")

	// Verify user was actually created by checking it doesn't exist in initial list
	userIsActuallyNew := true
	for _, u := range usersListInitial {
		if u.GetUserID() == employeeID {
			userIsActuallyNew = false
			break
		}
	}
	suite.True(userIsActuallyNew, "User should be new")

	// Get users list after creation and verify new user exists
	usersListPostCreate, err := suite.client.UserManagementService.ListCompanyUsers(ctx)
	suite.NoError(err)

	foundNewUser := false
	for _, u := range usersListPostCreate {
		if u.GetUserID() == employeeID {
			foundNewUser = true
			suite.Equal("test", u.FirstName)
			suite.Equal("user staging", u.LastName)
			suite.Equal("megaport.testuser@abcd.com", u.Email)
			suite.True(u.Active)
			suite.Equal("Company Admin", u.Position)
			break
		}
	}
	suite.True(foundNewUser, "Should find the newly created user in the users list")

	// Test Read operation
	suite.testReadUser(suite.client, ctx, employeeID)

	// Test Update operation
	suite.testUpdateUser(suite.client, ctx, employeeID)

	// Test Delete operation
	suite.testDeleteUser(suite.client, ctx, employeeID)
}

func (suite *UserManagementIntegrationTestSuite) testCreateUser(c *Client, ctx context.Context) (*CreateUserResponse, error) {
	suite.client.Logger.DebugContext(ctx, "Creating User")

	createReq := &CreateUserRequest{
		FirstName: "test",
		LastName:  "user staging",
		Active:    true,
		Email:     "megaport.testuser@abcd.com",
		Phone:     "+14155552671",
		Position:  USER_POSITION_COMPANY_ADMIN,
	}

	suite.client.Logger.DebugContext(ctx, "Sending create user request",
		slog.String("firstName", createReq.FirstName),
		slog.String("lastName", createReq.LastName),
		slog.String("email", createReq.Email),
		slog.Bool("active", createReq.Active),
		slog.String("position", string(createReq.Position)))

	createRes, err := c.UserManagementService.CreateUser(ctx, createReq)
	if err != nil {
		suite.client.Logger.ErrorContext(ctx, "Failed to create user", slog.String("error", err.Error()))
		return nil, err
	}

	suite.client.Logger.DebugContext(ctx, "User created successfully",
		slog.Int("employeeID", createRes.EmployeeID),
		slog.Int("employmentID", createRes.EmploymentID),
		slog.Int("companyID", createRes.CompanyID))

	return createRes, nil
}

func (suite *UserManagementIntegrationTestSuite) testReadUser(c *Client, ctx context.Context, employeeID int) {
	suite.client.Logger.DebugContext(ctx, "Reading User", slog.Int("employee_id", employeeID))

	user, err := c.UserManagementService.GetUser(ctx, employeeID)
	suite.NoError(err)
	suite.NotNil(user)

	suite.client.Logger.DebugContext(ctx, "Retrieved user details",
		slog.Int("partyId", user.PartyId),
		slog.Int("personId", user.PersonId),
		slog.String("firstName", user.FirstName),
		slog.String("lastName", user.LastName),
		slog.String("email", user.Email),
		slog.Bool("active", user.Active),
		slog.String("position", user.Position))

	// Verify user data matches what we created
	suite.Equal("test", user.FirstName)
	suite.Equal("user staging", user.LastName)
	suite.Equal("megaport.testuser@abcd.com", user.Email)
	suite.True(user.Active)
	suite.Equal("Company Admin", user.Position)
	suite.Equal(employeeID, user.GetUserID())
}

func (suite *UserManagementIntegrationTestSuite) testUpdateUser(c *Client, ctx context.Context, employeeID int) {
	suite.client.Logger.DebugContext(ctx, "Updating User", slog.Int("employee_id", employeeID))

	// Update user's first name and active status
	newFirstName := "updated-test"
	newLastName := "updated-user"
	updateReq := &UpdateUserRequest{
		FirstName: &newFirstName,
		LastName:  &newLastName,
	}

	suite.client.Logger.DebugContext(ctx, "Sending update user request",
		slog.String("newFirstName", newFirstName),
		slog.String("newLastName", newLastName))

	err := c.UserManagementService.UpdateUser(ctx, employeeID, updateReq)
	suite.NoError(err)

	// Verify the update by reading the user again
	updatedUser, err := c.UserManagementService.GetUser(ctx, employeeID)
	suite.NoError(err)
	suite.NotNil(updatedUser)

	suite.client.Logger.DebugContext(ctx, "Verified updated user details",
		slog.String("firstName", updatedUser.FirstName),
		slog.Bool("active", updatedUser.Active))

	suite.Equal(newFirstName, updatedUser.FirstName)
	suite.Equal(newLastName, updatedUser.LastName)
	suite.Equal("megaport.testuser@abcd.com", updatedUser.Email)
	suite.Equal(true, updatedUser.Active)
}

func (suite *UserManagementIntegrationTestSuite) testDeleteUser(c *Client, ctx context.Context, employeeID int) {
	suite.client.Logger.DebugContext(ctx, "Deleting User", slog.Int("employee_id", employeeID))

	err := c.UserManagementService.DeleteUser(ctx, employeeID)
	suite.NoError(err)

	suite.client.Logger.DebugContext(ctx, "User deleted successfully")

	// Verify user is deleted by trying to get it (should fail)
	_, err = c.UserManagementService.GetUser(ctx, employeeID)
	suite.Error(err, "Getting deleted user should return an error")

	suite.client.Logger.DebugContext(ctx, "Verified user deletion - GetUser returned error as expected")
}
