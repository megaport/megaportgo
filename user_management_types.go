package megaport

type UserEmail struct {
	EmailAddressId int     `json:"emailAddressId"`
	Email          string  `json:"email"`
	Primary        bool    `json:"primary"`
	BadEmail       bool    `json:"badEmail"`
	BadEmailType   *string `json:"badEmailType"`
	BadEmailReason *string `json:"badEmailReason"`
}

type User struct {
	Salutation                 string      `json:"salutation"`
	Position                   string      `json:"position"`
	FirstName                  string      `json:"firstName"`
	LastName                   string      `json:"lastName"`
	Phone                      string      `json:"phone"`
	Mobile                     string      `json:"mobile"`
	Email                      string      `json:"email"`
	PartyId                    int         `json:"partyId"`
	Username                   string      `json:"username"`
	Description                string      `json:"description"`
	Active                     bool        `json:"active"`
	UID                        string      `json:"uid"`
	Emails                     []UserEmail `json:"emails"`
	SalesforceId               string      `json:"salesforceId"`
	ChannelManager             bool        `json:"channelManager"`
	RequireTotp                bool        `json:"requireTotp"`
	NotificationEnabled        bool        `json:"notificationEnabled"`
	SecurityRoles              []string    `json:"securityRoles"`
	FeatureFlags               []string    `json:"featureFlags"`
	Newsletter                 bool        `json:"newsletter"`
	Promotions                 bool        `json:"promotions"`
	MfaEnabled                 bool        `json:"mfaEnabled"`
	ConfirmationPending        bool        `json:"confirmationPending"`
	Name                       string      `json:"name"`
	ReceivesChildNotifications bool        `json:"receivesChildNotifications"`
}

type UserPosition string

// UserPosition constants for known Megaport user roles.

// USER_POSITION_COMPANY_ADMIN represents a Company Admin user.
// Company Admin users have access to all user privileges.
// We recommend limiting the number of Company Admin users to only those who require full access, but defining at least two for redundancy.
const USER_POSITION_COMPANY_ADMIN UserPosition = "Company Admin"

// USER_POSITION_TECHNICAL_ADMIN represents a Technical Admin user.
// This role is for technical users who know how to create and approve orders.
const USER_POSITION_TECHNICAL_ADMIN UserPosition = "Technical Admin"

// USER_POSITION_TECHNICAL_CONTACT represents a Technical Contact user.
// This role is for technical users who know how to design and modify services but don’t have the authority to approve orders.
const USER_POSITION_TECHNICAL_CONTACT UserPosition = "Technical Contact"

// USER_POSITION_FINANCE represents a Finance user.
// Finance users should have a financial responsibility within the organization while also understanding the consequences of their actions if they delete or approve services.
const USER_POSITION_FINANCE UserPosition = "Finance"

// USER_POSITION_FINANCIAL_CONTACT represents a Financial Contact user.
// This user role is similar to the Finance role without the ability to place and approve orders, delete services, or administer service keys.
const USER_POSITION_FINANCIAL_CONTACT UserPosition = "Financial Contact"

// USER_POSITION_READ_ONLY represents a Read Only user.
// Read Only is the most restrictive role. Note that a Read Only user can view service details which you may want to keep secure and private.
const USER_POSITION_READ_ONLY UserPosition = "Read Only"

type UserActivity struct {
	// LoginName is the display name of the user who performed the activity.
	LoginName string `json:"loginName"`
	// PersonId is the unique identifier for the user (person).
	PersonId int `json:"personId"`
	// Description provides details about the activity performed.
	Description string `json:"description"`
	// Name is the type or name of the activity (e.g., "Login").
	Name string `json:"name"`
	// CreateDate is the timestamp when the activity occurred, parsed as a Time type.
	CreateDate Time `json:"createDate"`
	// UserType indicates the type of user (e.g., "USER").
	UserType string `json:"userType"`
}
