package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// UserManagementService is an interface that defines methods for managing users in the Megaport API.
type UserManagementService interface {
	// CreateUser creates a new user in the Megaport system.
	CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error)
	// GetUser retrieves a user by their employee ID.
	GetUser(ctx context.Context, employeeID int) (*User, error)
	// ListCompanyUsers retrieves a list of all users in the company.
	ListCompanyUsers(ctx context.Context) ([]*User, error)
	// UpdateUser updates an existing user in the Megaport system.
	UpdateUser(ctx context.Context, employeeID int, req *UpdateUserRequest) error
	// DeleteUser deletes a user by their employee ID.
	DeleteUser(ctx context.Context, employeeID int) error
	// DeactivateUser deactivates a user by their employee ID.
	DeactivateUser(ctx context.Context, employeeID int) error
	// GetUserActivity retrieves the activity of a user based on their person ID or UID.
	GetUserActivity(ctx context.Context, req *GetUserActivityRequest) ([]*UserActivity, error)
}

type UserManagementServiceOp struct {
	client *Client
}

func NewUserManagementService(client *Client) UserManagementService {
	return &UserManagementServiceOp{
		client: client,
	}
}

type CreateUserRequest struct {
	FirstName string       `json:"firstName"`
	LastName  string       `json:"lastName"`
	Active    bool         `json:"active"`
	Email     string       `json:"email"`
	Phone     string       `json:"phone"`
	Position  UserPosition `json:"position"`
}

// Validate validates the CreateUserRequest according to the Megaport API requirements
func (req *CreateUserRequest) Validate() error {
	// Email validation - required, valid email format, minimum 5 characters
	if req.Email == "" {
		return errors.New("email is required")
	}
	if len(req.Email) < 5 {
		return errors.New("email must be at least 5 characters long")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}

	// FirstName validation - required, non-empty
	if strings.TrimSpace(req.FirstName) == "" {
		return errors.New("firstName is required and cannot be empty")
	}

	// LastName validation - required, non-empty
	if strings.TrimSpace(req.LastName) == "" {
		return errors.New("lastName is required and cannot be empty")
	}

	// Phone validation - optional, but if provided must match international format
	if req.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(req.Phone) {
			return errors.New("phone number must be in valid international format (e.g., +1234567890)")
		}
	}

	// Position validation - required, must be one of the valid roles
	if req.Position == "" {
		return errors.New("position is required")
	}
	if !req.Position.IsValid() {
		return fmt.Errorf("invalid position: %s. Must be one of: %s", req.Position, req.Position.ValidPositions())
	}

	return nil
}

// createUserAPIResponse represents the API response when creating a new user.
// Used internally for JSON unmarshalling.
type createUserAPIResponse struct {
	Message string              `json:"message"`
	Terms   string              `json:"terms"`
	Data    *CreateUserResponse `json:"data"`
}

// CreateUserResponse represents the data returned when creating a new user.
type CreateUserResponse struct {
	CompanyID    int `json:"companyId"`
	EmploymentID int `json:"employmentId"`
	EmployeeID   int `json:"employeeId"`
}

// UpdateUserRequest represents the request body for updating an existing user.
// Only include the attributes that you want to update.
type UpdateUserRequest struct {
	// NotificationEnabled indicates whether the user should receive notifications.
	NotificationEnabled *bool `json:"notificationEnabled,omitempty"`

	// Position defines the role of the user within the organization.
	// Use the UserPosition type constants for standard roles (e.g., USER_POSITION_COMPANY_ADMIN).
	// For more information on roles, see Add / invite user to company.
	Position *string `json:"position,omitempty"`

	// CompanyId is the unique identifier of the company to which the user belongs.
	// This ID is used to associate the user with a specific company.
	CompanyId *int `json:"companyId,omitempty"`

	// Newsletter indicates whether the user has opted into receiving the newsletter.
	Newsletter *bool `json:"newsletter,omitempty"`

	// Promotions indicates whether the user has opted into receiving promotional communications.
	Promotions *bool `json:"promotions,omitempty"`

	// FirstName is the user's first name.
	FirstName *string `json:"firstName,omitempty"`

	// LastName is the user's last name.
	LastName *string `json:"lastName,omitempty"`

	// Phone is the user's primary phone number.
	Phone *string `json:"phone,omitempty"`

	// ChannelManager indicates whether the user is a channel manager.
	// This field is used to specify the channel manager associated with the user.
	ChannelManager *bool `json:"channelManager,omitempty"`

	// Active indicates whether the user account is active.
	// Set to true if the user is active or false to define the user as inactive.
	Active *bool `json:"active,omitempty"`

	// SecurityRoles defines the array of security roles assigned to the user.
	// Examples include "companyAdmin", "technicalContact".
	SecurityRoles *[]string `json:"securityRoles,omitempty"`

	// Email is the email address of the user.
	Email *string `json:"email,omitempty"`
}

// Validate validates the UpdateUserRequest according to the Megaport API requirements
func (req *UpdateUserRequest) Validate() error {
	// FirstName validation - if provided, must be non-empty
	if req.FirstName != nil && strings.TrimSpace(*req.FirstName) == "" {
		return errors.New("firstName cannot be empty if provided")
	}

	// LastName validation - if provided, must be non-empty
	if req.LastName != nil && strings.TrimSpace(*req.LastName) == "" {
		return errors.New("lastName cannot be empty if provided")
	}

	// Phone validation - if provided, must match international format
	if req.Phone != nil && *req.Phone != "" {
		phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
		if !phoneRegex.MatchString(*req.Phone) {
			return errors.New("phone number must be in valid international format (e.g., +1234567890)")
		}
	}

	// Email validation - if provided, must be valid email format with minimum 5 characters
	if req.Email != nil {
		if *req.Email != "" {
			if len(*req.Email) < 5 {
				return errors.New("email must be at least 5 characters long")
			}
			if _, err := mail.ParseAddress(*req.Email); err != nil {
				return fmt.Errorf("invalid email format: %w", err)
			}
		}
	}

	// Position validation - if provided, must be one of the valid roles
	if req.Position != nil && *req.Position != "" {
		// Convert string to UserPosition for validation
		pos := UserPosition(*req.Position)
		if !pos.IsValid() {
			return fmt.Errorf("invalid position: %s. Must be one of: %s", *req.Position, pos.ValidPositions())
		}
	}

	// CompanyId validation - if provided, must be positive
	if req.CompanyId != nil && *req.CompanyId <= 0 {
		return errors.New("companyId must be a positive integer")
	}

	return nil
}

// GetUserActivityRequest represents the query parameters for retrieving user activity.
// If a field is not an empty string, it will be included as a query parameter.
type GetUserActivityRequest struct {
	// PersonIdOrUid is the user's person ID or UID to filter activity for a specific user.
	PersonIdOrUid string
	// CompanyIdOrUid is the company ID or UID to filter activity for a specific company.
	CompanyIdOrUid string
}

// CreateUser creates a new user in the Megaport system.
func (svc *UserManagementServiceOp) CreateUser(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
	// Validate the request according to API requirements
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	path := "/v2/employment"
	clientReq, err := svc.client.NewRequest(ctx, "POST", path, req)
	if err != nil {
		return nil, err
	}

	response, err := svc.client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 201 {
		svc.client.Logger.ErrorContext(ctx, "CreateUser failed",
			slog.Int("status_code", response.StatusCode),
			slog.String("response", string(body)))
		return nil, fmt.Errorf("failed to create user: HTTP %d - %s", response.StatusCode, string(body))
	}

	var apiResponse *createUserAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, err
	}

	return apiResponse.Data, nil
}

func (svc *UserManagementServiceOp) GetUser(ctx context.Context, employeeID int) (*User, error) {
	path := "/v2/employee/" + strconv.Itoa(employeeID)
	clientReq, err := svc.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		svc.client.Logger.ErrorContext(ctx, "GetUser failed",
			slog.Int("employee_id", employeeID),
			slog.Int("status_code", response.StatusCode),
			slog.String("response", string(body)))
		return nil, fmt.Errorf("failed to get user %d: HTTP %d - %s", employeeID, response.StatusCode, string(body))
	}

	// The API response is wrapped in an object with message, terms, and data fields
	var apiResponse struct {
		Message string `json:"message"`
		Terms   string `json:"terms"`
		Data    User   `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		svc.client.Logger.ErrorContext(ctx, "Failed to unmarshal GetUser response",
			slog.Int("employee_id", employeeID),
			slog.String("error", err.Error()),
			slog.String("body", string(body)))
		return nil, err
	}

	user := &apiResponse.Data

	return user, nil
}

func (svc *UserManagementServiceOp) ListCompanyUsers(ctx context.Context) ([]*User, error) {
	path := "/v2/employment"
	clientReq, err := svc.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		svc.client.Logger.ErrorContext(ctx, "ListCompanyUsers failed",
			slog.Int("status_code", response.StatusCode),
			slog.String("response", string(body)))
		return nil, fmt.Errorf("failed to list company users: HTTP %d - %s", response.StatusCode, string(body))
	}

	svc.client.Logger.DebugContext(ctx, "ListCompanyUsers API response received",
		slog.Int("response_size", len(body)))

	// The API response is wrapped in an object with message, terms, and data fields
	var apiResponse struct {
		Message string  `json:"message"`
		Terms   string  `json:"terms"`
		Data    []*User `json:"data"`
	}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		svc.client.Logger.ErrorContext(ctx, "Failed to unmarshal ListCompanyUsers response",
			slog.String("error", err.Error()),
			slog.String("body", string(body)))
		return nil, err
	}

	svc.client.Logger.DebugContext(ctx, "ListCompanyUsers parsed successfully",
		slog.Int("user_count", len(apiResponse.Data)))

	return apiResponse.Data, nil
}

func (svc *UserManagementServiceOp) UpdateUser(ctx context.Context, employeeID int, req *UpdateUserRequest) error {
	// Validate the request according to API requirements
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get User Details First Before Updating
	existingUser, err := svc.GetUser(ctx, employeeID)
	if err != nil {
		return err
	}
	if existingUser.InvitationPending {
		return fmt.Errorf("cannot update user %d: user has not accepted invitation yet (invitation pending)", employeeID)
	}

	path := "/v2/employee/" + strconv.Itoa(employeeID)

	svc.client.Logger.DebugContext(ctx, "Updating user",
		slog.Int("employee_id", employeeID))

	clientReq, err := svc.client.NewRequest(ctx, "PUT", path, req)
	if err != nil {
		return err
	}

	response, err := svc.client.Do(ctx, clientReq, nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		return errors.New("failed to update user: " + string(body))
	}

	return nil
}

func (svc *UserManagementServiceOp) DeactivateUser(ctx context.Context, employeeID int) error {
	// Make direct HTTP request to deactivate user, bypassing UpdateUser validation
	path := "/v2/employee/" + strconv.Itoa(employeeID)

	// Create request payload to set active to false
	active := false
	req := &UpdateUserRequest{
		Active: &active,
	}

	svc.client.Logger.DebugContext(ctx, "Deactivating user via direct HTTP request",
		slog.Int("employee_id", employeeID))

	clientReq, err := svc.client.NewRequest(ctx, "PUT", path, req)
	if err != nil {
		return err
	}

	response, err := svc.client.Do(ctx, clientReq, nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		svc.client.Logger.ErrorContext(ctx, "DeactivateUser failed",
			slog.Int("employee_id", employeeID),
			slog.Int("status_code", response.StatusCode),
			slog.String("response", string(body)))
		return fmt.Errorf("failed to deactivate user %d: HTTP %d - %s", employeeID, response.StatusCode, string(body))
	}

	return nil
}

func (svc *UserManagementServiceOp) DeleteUser(ctx context.Context, employeeID int) error {
	// If user has logged in, we cannot delete them, only deactivate.
	existingUser, err := svc.GetUser(ctx, employeeID)
	if err != nil {
		return err
	}
	if !existingUser.InvitationPending {
		return fmt.Errorf("user %d has already logged in and cannot be deleted, only deactivated", employeeID)
	}
	path := "/v2/employee/" + strconv.Itoa(employeeID)
	clientReq, err := svc.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	response, err := svc.client.Do(ctx, clientReq, nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		body, _ := io.ReadAll(response.Body)
		return errors.New("failed to delete user: " + string(body))
	}

	return nil
}

// GetUserActivity retrieves a log of user activity in the Megaport Portal.
// It can be filtered by person ID/UID and/or company ID/UID using the request parameters.
func (svc *UserManagementServiceOp) GetUserActivity(ctx context.Context, req *GetUserActivityRequest) ([]*UserActivity, error) {
	path := "/v3/activity"

	// Append query parameters if provided using proper URL encoding
	queryParams := url.Values{}
	if req.PersonIdOrUid != "" {
		queryParams.Add("personIdOrUid", req.PersonIdOrUid)
	}
	if req.CompanyIdOrUid != "" {
		queryParams.Add("companyIdOrUid", req.CompanyIdOrUid)
	}

	if len(queryParams) > 0 {
		path += "?" + queryParams.Encode()
	}

	clientReq, err := svc.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := svc.client.Do(ctx, clientReq, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		svc.client.Logger.ErrorContext(ctx, "GetUserActivity failed",
			slog.Int("status_code", response.StatusCode),
			slog.String("response", string(body)))
		return nil, fmt.Errorf("failed to get user activity: HTTP %d - %s", response.StatusCode, string(body))
	}

	var activities []*UserActivity
	if err := json.Unmarshal(body, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}
