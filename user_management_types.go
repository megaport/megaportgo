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
	PersonId                   int         `json:"personId"` // Used in list responses instead of partyId
	Username                   string      `json:"username"`
	Description                string      `json:"description"`
	Active                     bool        `json:"active"`
	UID                        string      `json:"uid"`
	PersonUid                  string      `json:"personUid"` // Alternative UID field in list responses
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

	// Additional fields from list API response
	CompanyId      int    `json:"companyId"`
	EmploymentId   int    `json:"employmentId"`
	PositionId     int    `json:"positionId"`
	PersonAltId    string `json:"personAltId"`
	EmploymentType string `json:"employmentType"`
	CompanyName    string `json:"companyName"`
}

type UserPosition string

// UserPosition constants for known Megaport user roles.
const (
	// USER_POSITION_COMPANY_ADMIN represents a Company Admin user.
	// Company Admin users have access to all user privileges.
	// We recommend limiting the number of Company Admin users to only those who require full access, but defining at least two for redundancy.
	USER_POSITION_COMPANY_ADMIN UserPosition = "Company Admin"

	// USER_POSITION_TECHNICAL_ADMIN represents a Technical Admin user.
	// This role is for technical users who know how to create and approve orders.
	USER_POSITION_TECHNICAL_ADMIN UserPosition = "Technical Admin"

	// USER_POSITION_TECHNICAL_CONTACT represents a Technical Contact user.
	// This role is for technical users who know how to design and modify services but don't have the authority to approve orders.
	USER_POSITION_TECHNICAL_CONTACT UserPosition = "Technical Contact"

	// USER_POSITION_FINANCE represents a Finance user.
	// Finance users should have a financial responsibility within the organization while also understanding the consequences of their actions if they delete or approve services.
	USER_POSITION_FINANCE UserPosition = "Finance"

	// USER_POSITION_FINANCIAL_CONTACT represents a Financial Contact user.
	// This user role is similar to the Finance role without the ability to place and approve orders, delete services, or administer service keys.
	USER_POSITION_FINANCIAL_CONTACT UserPosition = "Financial Contact"

	// USER_POSITION_READ_ONLY represents a Read Only user.
	// Read Only is the most restrictive role. Note that a Read Only user can view service details which you may want to keep secure and private.
	USER_POSITION_READ_ONLY UserPosition = "Read Only"
)

// IsValid checks if the UserPosition is one of the valid predefined positions
func (p UserPosition) IsValid() bool {
	switch p {
	case USER_POSITION_COMPANY_ADMIN,
		USER_POSITION_TECHNICAL_ADMIN,
		USER_POSITION_TECHNICAL_CONTACT,
		USER_POSITION_FINANCE,
		USER_POSITION_FINANCIAL_CONTACT,
		USER_POSITION_READ_ONLY:
		return true
	default:
		return false
	}
}

// ValidPositions returns a string listing all valid UserPosition values
func (p UserPosition) ValidPositions() string {
	return string(USER_POSITION_COMPANY_ADMIN) + ", " +
		string(USER_POSITION_TECHNICAL_ADMIN) + ", " +
		string(USER_POSITION_TECHNICAL_CONTACT) + ", " +
		string(USER_POSITION_FINANCE) + ", " +
		string(USER_POSITION_FINANCIAL_CONTACT) + ", " +
		string(USER_POSITION_READ_ONLY)
}

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
