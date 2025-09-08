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

// TestUserCRD tests the lifecycle of user management: Create, Read, Delete
// Note: Update operations are skipped as they require email confirmation which is not suitable for automated testing
func (suite *UserManagementIntegrationTestSuite) TestUserCRD() {
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
		if u.PersonId == employeeID {
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
		if u.PersonId == employeeID {
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

	// Skip Update operation - requires email confirmation which is not suitable for automated testing

	// Test Delete operation
	suite.testDeleteUser(suite.client, ctx, employeeID)
}

// TestUpdateUserPendingConfirmation tests that updating a user with confirmationPending=true returns an error
func (suite *UserManagementIntegrationTestSuite) TestUpdateUserPendingConfirmation() {
	ctx := context.Background()

	// Create a user first (newly created users have confirmationPending=true)
	createdUser, err := suite.testCreateUser(suite.client, ctx)
	suite.NoError(err)
	suite.NotNil(createdUser)

	employeeID := createdUser.EmployeeID
	suite.True(employeeID > 0, "Employee ID should be greater than 0")

	// Verify the user has confirmationPending=true (newly created users should have this)
	user, err := suite.client.UserManagementService.GetUser(ctx, employeeID)
	suite.NoError(err)
	suite.NotNil(user)

	// Try to update the user while confirmation is pending - this should fail
	newFirstName := "UpdatedName"
	updateReq := &UpdateUserRequest{
		FirstName: &newFirstName,
	}

	suite.client.Logger.DebugContext(ctx, "Attempting to update user with pending confirmation",
		slog.Int("employee_id", employeeID),
		slog.Bool("confirmation_pending", user.ConfirmationPending))

	err = suite.client.UserManagementService.UpdateUser(ctx, employeeID, updateReq)

	// This should return an error because the user has confirmationPending=true
	suite.Error(err, "Updating user with pending confirmation should return an error")
	suite.Contains(err.Error(), "pending confirmation", "Error should mention pending confirmation")

	suite.client.Logger.DebugContext(ctx, "Update correctly failed for user with pending confirmation",
		slog.String("error", err.Error()))

	// Clean up - delete the test user
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
		slog.String("first_name", createReq.FirstName),
		slog.String("last_name", createReq.LastName),
		slog.String("email", createReq.Email),
		slog.Bool("active", createReq.Active),
		slog.String("position", string(createReq.Position)))

	createRes, err := c.UserManagementService.CreateUser(ctx, createReq)
	if err != nil {
		suite.client.Logger.ErrorContext(ctx, "Failed to create user", slog.String("error", err.Error()))
		return nil, err
	}

	suite.client.Logger.DebugContext(ctx, "User created successfully",
		slog.Int("employee_id", createRes.EmployeeID),
		slog.Int("employment_id", createRes.EmploymentID),
		slog.Int("company_id", createRes.CompanyID))

	return createRes, nil
}

func (suite *UserManagementIntegrationTestSuite) testReadUser(c *Client, ctx context.Context, employeeID int) {
	suite.client.Logger.DebugContext(ctx, "Reading User", slog.Int("employee_id", employeeID))

	user, err := c.UserManagementService.GetUser(ctx, employeeID)
	suite.NoError(err)
	suite.NotNil(user)

	suite.client.Logger.DebugContext(ctx, "Retrieved user details",
		slog.Int("party_id", user.PartyId),
		slog.Int("person_id", user.PersonId),
		slog.String("first_name", user.FirstName),
		slog.String("last_name", user.LastName),
		slog.String("email", user.Email),
		slog.Bool("active", user.Active),
		slog.String("position", user.Position))

	// Verify user data matches what we created
	suite.Equal("test", user.FirstName)
	suite.Equal("user staging", user.LastName)
	suite.Equal("megaport.testuser@abcd.com", user.Email)
	suite.True(user.Active)
	suite.Equal("Company Admin", user.Position)
	suite.Equal(employeeID, user.PartyId)
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
