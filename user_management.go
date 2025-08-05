package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strconv"
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

// CreateUserAPIResponse represents the API response when creating a new user.
type CreateUserAPIResponse struct {
	Message string              `json:"message"`
	Terms   string              `json:"terms"`
	Data    *CreateUserResponse `json:"data"`
}

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
	Position *UserPosition `json:"position,omitempty"`

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

	// RequireTotp indicates whether multi-factor authentication using
	// time-based one-time passwords is required for this user.
	RequireTotp *bool `json:"requireTotp,omitempty"`

	// Active indicates whether the user account is active.
	// Set to false to deactivate a user without deleting their account.
	Active *bool `json:"active,omitempty"`

	// SecurityRoles defines the array of security roles assigned to the user.
	// Examples include "companyAdmin", "technicalContact".
	SecurityRoles *[]string `json:"securityRoles,omitempty"`

	// Mobile is the user's mobile phone number.
	Mobile *string `json:"mobile,omitempty"`

	// Email is the primary email address for the user.
	Email *string `json:"email,omitempty"`

	// PartyId is the employee ID of the user, equivalent to the personId.
	PartyId *string `json:"partyId,omitempty"`

	// AltId is the Google ID associated with the user.
	AltId *string `json:"altId,omitempty"`

	// Username is the login username for the user.
	Username *string `json:"username,omitempty"`

	// UID is the unique identifier for the user.
	UID *string `json:"uid,omitempty"`

	// GooglePlus is the Google social profile link for the user.
	GooglePlus *string `json:"googlePlus,omitempty"`

	// SalesforceId is the ID of the user in Salesforce CRM.
	SalesforceId *string `json:"salesforceId,omitempty"`

	// Name is the full name of the user.
	Name *string `json:"name,omitempty"`

	// Emails is an array of email objects associated with the user.
	// The first email in the array is considered the primary email.
	Emails *[]UserEmail `json:"emails,omitempty"`
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

	var apiResponse *CreateUserAPIResponse
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

	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, err
	}

	return &user, nil
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

	var users []*User
	if err := json.Unmarshal(body, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (svc *UserManagementServiceOp) UpdateUser(ctx context.Context, employeeID int, req *UpdateUserRequest) error {
	path := "/v2/employee/" + strconv.Itoa(employeeID)
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

func (svc *UserManagementServiceOp) DeleteUser(ctx context.Context, employeeID int) error {
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

	// Append query parameters if provided
	first := true
	if req.PersonIdOrUid != "" {
		path += "?personIdOrUid=" + req.PersonIdOrUid
		first = false
	}

	if req.CompanyIdOrUid != "" {
		if first {
			path += "?"
		} else {
			path += "&"
		}
		path += "companyIdOrUid=" + req.CompanyIdOrUid
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

	var activities []*UserActivity
	if err := json.Unmarshal(body, &activities); err != nil {
		return nil, err
	}

	return activities, nil
}
