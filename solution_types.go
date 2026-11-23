package megaport

import "time"

// SolutionTypeDCI is the Solution type of a Data Centre Interconnect (DCI): two Ports joined by one VXC,
// ordered, priced and billed as one item.
const SolutionTypeDCI = "DCI"

// The roles a member service can have in a Solution. The API may add roles for new Solution types.
const (
	SolutionMemberRoleAEndPort = "A_END_PORT"
	SolutionMemberRoleBEndPort = "B_END_PORT"
	SolutionMemberRoleVXC      = "VXC"
)

// Solution represents a Solution in the Megaport API: a named group of services, such as a DCI, created at buy time.
type Solution struct {
	UID                   string `json:"uid"`
	Type                  string `json:"type"`
	Name                  string `json:"name"`
	Active                bool   `json:"active"`
	Locked                bool   `json:"locked"`
	ServiceLevelReference string `json:"serviceLevelReference"`
	ProjectUID            string `json:"projectUid"`

	// Members are the Solution's services, ordered A-End Port, B-End Port, then VXC.
	Members []SolutionMember `json:"members"`
	// ProvisioningStatus is the Solution's overall status, which the API rolls up from its members.
	ProvisioningStatus string `json:"provisioningStatus"`
	// ALocationID and BLocationID are 0 until the Solution's VXC is created.
	ALocationID int `json:"aLocationId"`
	BLocationID int `json:"bLocationId"`
	// SpeedMbps is taken from the Solution's VXC.
	SpeedMbps   int    `json:"speedMbps"`
	CreateDate  *Time  `json:"createDate"`
	CreatedBy   string `json:"createdBy"`
	LiveDate    *Time  `json:"liveDate"`
	CompanyUID  string `json:"companyUid"`
	CompanyName string `json:"companyName"`
}

// MemberUID returns the product UID of the member with the given role, or "" when the Solution has none.
func (s *Solution) MemberUID(role string) string {
	for _, m := range s.Members {
		if m.Role == role {
			return m.ServiceUID
		}
	}
	return ""
}

// SolutionMember is a service in a Solution.
type SolutionMember struct {
	ServiceUID         string `json:"serviceUid"`
	ServiceName        string `json:"serviceName"`
	ServiceType        string `json:"serviceType"`
	ProvisioningStatus string `json:"provisioningStatus"`
	// Role is one of the SolutionMemberRole constants, or "" for a Port at neither end of the Solution's VXC.
	Role string `json:"role"`
	// Position is the member's place in the Solution, starting at 0.
	Position      int    `json:"position"`
	LocationID    int    `json:"locationId"`
	DiversityZone string `json:"diversityZone"`
}

// SolutionBillable is a Solution's current term, as returned by GET /v1/solutions/{uid}/billable.
// A Solution is billed as one product, so its member services have no billing of their own.
type SolutionBillable struct {
	ContractTermMonths int  `json:"contractTermMonths"`
	AutoRenewTerm      bool `json:"autoRenewTerm"`
	// CanReterm reports whether the Solution can be moved to a different term.
	CanReterm bool `json:"canReterm"`
	// ContractStartDate and ContractEndDate are nil before billing starts.
	ContractStartDate *Time `json:"contractStartDate"`
	ContractEndDate   *Time `json:"contractEndDate"`
}

// BuyDCIRequest represents a request to buy a DCI: a Port at each end and a VXC between them, placed as one order.
// The API accepts a Term of 12, 24 or 36 months. Both Ports and the VXC take the same PortSpeed, Term and AutoRenewTerm.
type BuyDCIRequest struct {
	Name                  string // The name of the DCI. Each member service is ordered with this name.
	AEndLocationID        int
	BEndLocationID        int
	PortSpeed             int // The speed of both Ports and of the VXC, in Mbps.
	Term                  int
	AutoRenewTerm         bool
	ServiceLevelReference string
	AEndDiversityZone     string
	BEndDiversityZone     string

	WaitForProvision bool          // Wait until the DCI provisions before returning
	WaitForTime      time.Duration // How long to wait for the DCI to provision if WaitForProvision is true (default is 5 minutes)
}

// BuyDCIResponse represents a response from buying a DCI. Read the member services with GetSolution.
type BuyDCIResponse struct {
	SolutionUID string
}

// UpdateSolutionRequest represents a request to change a Solution's name and service level reference.
// The API sets both on every update: it cannot leave one as it is.
type UpdateSolutionRequest struct {
	SolutionUID string `json:"-"`
	// Name is the new name of the Solution. It is required.
	Name string `json:"name"`
	// ServiceLevelReference replaces the reference on the Solution and on every member service.
	// An empty value clears it. The API accepts at most 255 characters.
	ServiceLevelReference string `json:"serviceLevelReference"`
}

// CancelSolutionRequest represents a request to cancel a Solution. The API checks the reason and the comment
// before it cancels a member service.
type CancelSolutionRequest struct {
	SolutionUID string `json:"-"`
	// CancellationReason is one of the SolutionCancellationReason constants. The API rejects any other value.
	// It may be left empty.
	CancellationReason string `json:"cancellation_reason,omitempty"`
	// CancellationComment is free text recorded with the cancellation. The API accepts at most 400 characters.
	// Every reason except SolutionCancellationReasonBillingIssues requires a comment, and that one refuses it.
	CancellationComment string `json:"cancellation_comment,omitempty"`
}

// The reasons the API accepts for canceling a Solution. All but SolutionCancellationReasonBillingIssues
// require a CancellationComment.
const (
	SolutionCancellationReasonBillingIssues         = "BILLING_ISSUES"
	SolutionCancellationReasonMovedToCloudNative    = "MOVED_TO_CLOUD_NATIVE_SOL"
	SolutionCancellationReasonOther                 = "OTHER"
	SolutionCancellationReasonPricing               = "PRICING"
	SolutionCancellationReasonProductPerformance    = "PRODUCT_PERFORMANCE"
	SolutionCancellationReasonProjectEnded          = "PROJECT_ENDED"
	SolutionCancellationReasonProofOfConcept        = "PROOF_OF_CONCEPT"
	SolutionCancellationReasonServiceMoved          = "SERVICE_MOVED"
	SolutionCancellationReasonServiceOrderedInError = "SERVICE_ORDERED_IN_ERROR"
	SolutionCancellationReasonSwitchingProvider     = "SWITCHING_PROVIDER"
)

// solutionOrder declares a new Solution in the body of POST /v4/networkdesign/buy.
// The design's services point to it with solutionRef.
type solutionOrder struct {
	Ref                   string `json:"ref"`
	Type                  string `json:"type"`
	Name                  string `json:"name"`
	ServiceLevelReference string `json:"serviceLevelReference,omitempty"`
}

// dciPortOrder is a Port in a DCI order. ProductUID is a temporary UID the VXC names its ends by.
// The API takes a UID without the DESIGN status to mean an existing service, and rejects it.
type dciPortOrder struct {
	ProductUID         string          `json:"productUid"`
	Name               string          `json:"productName"`
	ProductType        string          `json:"productType"`
	ProvisioningStatus string          `json:"provisioningStatus"`
	LocationID         int             `json:"locationId"`
	PortSpeed          int             `json:"portSpeed"`
	Term               int             `json:"term"`
	AutoRenewTerm      bool            `json:"autoRenewTerm"`
	SolutionRef        string          `json:"solutionRef"`
	Config             PortOrderConfig `json:"config"`
	AssociatedVXCs     []dciVXCOrder   `json:"associatedVxcs"`
}

// dciVXCOrder is the VXC in a DCI order. It rides on the A-End Port.
type dciVXCOrder struct {
	ProductUID         string         `json:"productUid"`
	Name               string         `json:"productName"`
	ProductType        string         `json:"productType"`
	ProvisioningStatus string         `json:"provisioningStatus"`
	RateLimit          int            `json:"rateLimit"`
	Term               int            `json:"term"`
	AutoRenewTerm      bool           `json:"autoRenewTerm"`
	ConnectType        string         `json:"connectType"`
	AEnd               dciVXCOrderEnd `json:"aEnd"`
	BEnd               dciVXCOrderEnd `json:"bEnd"`
	SolutionRef        string         `json:"solutionRef"`
}

type dciVXCOrderEnd struct {
	ProductUID string `json:"productUid"`
}

// solutionOrderResponse is the part of the POST /v4/networkdesign/buy response that names the Solutions created.
// It is a sibling of data, which keeps its services-array shape.
type solutionOrderResponse struct {
	Solutions []solutionOrderConfirmation `json:"solutions"`
}

type solutionOrderConfirmation struct {
	UID  string `json:"uid"`
	Ref  string `json:"ref"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// solutionAPIResponse is the envelope returned by GET /v1/solutions/{uid}.
type solutionAPIResponse struct {
	Message string    `json:"message"`
	Terms   string    `json:"terms"`
	Data    *Solution `json:"data"`
}

// solutionBillableAPIResponse is the envelope returned by GET /v1/solutions/{uid}/billable.
type solutionBillableAPIResponse struct {
	Message string            `json:"message"`
	Terms   string            `json:"terms"`
	Data    *SolutionBillable `json:"data"`
}
