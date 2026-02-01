package megaport

// RouteProtocol represents the protocol type for a route.
type RouteProtocol string

const (
	RouteProtocolBGP       RouteProtocol = "BGP"
	RouteProtocolStatic    RouteProtocol = "STATIC"
	RouteProtocolConnected RouteProtocol = "CONNECTED"
	RouteProtocolLocal     RouteProtocol = "LOCAL"
)

// BGPSessionStatus represents the status of a BGP session.
type BGPSessionStatus string

const (
	BGPSessionStatusUp      BGPSessionStatus = "UP"
	BGPSessionStatusDown    BGPSessionStatus = "DOWN"
	BGPSessionStatusUnknown BGPSessionStatus = "UNKNOWN"
)

// LookingGlassIPRoute represents an IP route in the MCR routing table.
type LookingGlassIPRoute struct {
	Prefix      string        `json:"prefix"`      // The network prefix (e.g., "10.0.0.0/24")
	NextHop     string        `json:"nextHop"`     // The next hop IP address
	Protocol    RouteProtocol `json:"protocol"`    // The protocol that learned this route
	Metric      *int          `json:"metric"`      // The route metric (optional)
	LocalPref   *int          `json:"localPref"`   // BGP local preference (optional)
	ASPath      []int         `json:"asPath"`      // BGP AS path (optional)
	Age         *int          `json:"age"`         // Age of the route in seconds (optional)
	Interface   string        `json:"interface"`   // The interface for this route
	VXCId       *int          `json:"vxcId"`       // Associated VXC ID (optional)
	VXCName     string        `json:"vxcName"`     // Associated VXC name (optional)
	Communities []string      `json:"communities"` // BGP communities (optional)
	Origin      string        `json:"origin"`      // BGP origin attribute (optional)
	MED         *int          `json:"med"`         // BGP Multi-Exit Discriminator (optional)
	Best        *bool         `json:"best"`        // Whether this is the best route (optional)
}

// LookingGlassBGPRoute represents a BGP-specific route with full BGP attributes.
type LookingGlassBGPRoute struct {
	Prefix      string   `json:"prefix"`      // The network prefix
	NextHop     string   `json:"nextHop"`     // The next hop IP address
	ASPath      []int    `json:"asPath"`      // The AS path
	LocalPref   *int     `json:"localPref"`   // Local preference
	MED         *int     `json:"med"`         // Multi-Exit Discriminator
	Origin      string   `json:"origin"`      // Origin attribute (IGP, EGP, INCOMPLETE)
	Communities []string `json:"communities"` // BGP communities
	Weight      *int     `json:"weight"`      // BGP weight
	Valid       bool     `json:"valid"`       // Whether the route is valid
	Best        bool     `json:"best"`        // Whether this is the best path
	NeighborIP  string   `json:"neighborIp"`  // The BGP neighbor IP that advertised this route
	NeighborASN *int     `json:"neighborAsn"` // The BGP neighbor ASN
	Age         *int     `json:"age"`         // Age of the route in seconds
	VXCId       *int     `json:"vxcId"`       // Associated VXC ID
	VXCName     string   `json:"vxcName"`     // Associated VXC name
}

// LookingGlassBGPSession represents a BGP session on the MCR.
type LookingGlassBGPSession struct {
	SessionID       string           `json:"sessionId"`       // Unique identifier for the BGP session
	NeighborAddress string           `json:"neighborAddress"` // The BGP neighbor IP address
	NeighborASN     int              `json:"neighborAsn"`     // The BGP neighbor ASN
	LocalASN        int              `json:"localAsn"`        // The local ASN
	Status          BGPSessionStatus `json:"status"`          // Session status (UP, DOWN, UNKNOWN)
	Uptime          *int             `json:"uptime"`          // Session uptime in seconds
	PrefixesIn      *int             `json:"prefixesIn"`      // Number of prefixes received
	PrefixesOut     *int             `json:"prefixesOut"`     // Number of prefixes advertised
	VXCId           int              `json:"vxcId"`           // Associated VXC ID
	VXCName         string           `json:"vxcName"`         // Associated VXC name
	LastStateChange *int             `json:"lastStateChange"` // Seconds since last state change
	Description     string           `json:"description"`     // Session description
}

// LookingGlassBGPNeighborRoute represents a route advertised or received from a BGP neighbor.
type LookingGlassBGPNeighborRoute struct {
	Prefix      string   `json:"prefix"`      // The network prefix
	NextHop     string   `json:"nextHop"`     // The next hop IP address
	ASPath      []int    `json:"asPath"`      // The AS path
	LocalPref   *int     `json:"localPref"`   // Local preference
	MED         *int     `json:"med"`         // Multi-Exit Discriminator
	Origin      string   `json:"origin"`      // Origin attribute
	Communities []string `json:"communities"` // BGP communities
	Valid       bool     `json:"valid"`       // Whether the route is valid
	Best        bool     `json:"best"`        // Whether this is the best path
}

// LookingGlassRouteDirection represents the direction for BGP neighbor routes.
type LookingGlassRouteDirection string

const (
	LookingGlassRouteDirectionAdvertised LookingGlassRouteDirection = "advertised"
	LookingGlassRouteDirectionReceived   LookingGlassRouteDirection = "received"
)

// ListIPRoutesRequest represents a request to list IP routes from the MCR Looking Glass.
type ListIPRoutesRequest struct {
	MCRID    string        // The MCR UID
	Protocol RouteProtocol // Optional: filter by protocol
	IPFilter string        // Optional: filter by IP address or prefix
}

// ListBGPRoutesRequest represents a request to list BGP routes from the MCR Looking Glass.
type ListBGPRoutesRequest struct {
	MCRID    string // The MCR UID
	IPFilter string // Optional: filter by IP address or prefix
}

// ListBGPSessionsRequest represents a request to list BGP sessions from the MCR Looking Glass.
type ListBGPSessionsRequest struct {
	MCRID string // The MCR UID
}

// ListBGPNeighborRoutesRequest represents a request to list routes from a specific BGP neighbor.
type ListBGPNeighborRoutesRequest struct {
	MCRID     string                     // The MCR UID
	SessionID string                     // The BGP session ID
	Direction LookingGlassRouteDirection // The direction (advertised or received)
	IPFilter  string                     // Optional: filter by IP address or prefix
}

// LookingGlassAsyncStatus represents the status of an async Looking Glass operation.
type LookingGlassAsyncStatus string

const (
	LookingGlassAsyncStatusPending    LookingGlassAsyncStatus = "PENDING"
	LookingGlassAsyncStatusProcessing LookingGlassAsyncStatus = "PROCESSING"
	LookingGlassAsyncStatusComplete   LookingGlassAsyncStatus = "COMPLETE"
	LookingGlassAsyncStatusFailed     LookingGlassAsyncStatus = "FAILED"
)

// LookingGlassAsyncJob represents an async job for Looking Glass queries.
type LookingGlassAsyncJob struct {
	JobID     string                  `json:"jobId"`     // The unique job identifier
	Status    LookingGlassAsyncStatus `json:"status"`    // The job status
	CreatedAt *Time                   `json:"createdAt"` // When the job was created
	UpdatedAt *Time                   `json:"updatedAt"` // When the job was last updated
	ExpiresAt *Time                   `json:"expiresAt"` // When the job results expire
}

// API Response types for Looking Glass endpoints

// LookingGlassIPRoutesResponse represents the API response for IP routes.
type LookingGlassIPRoutesResponse struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    []*LookingGlassIPRoute `json:"data"`
}

// LookingGlassBGPRoutesResponse represents the API response for BGP routes.
type LookingGlassBGPRoutesResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    []*LookingGlassBGPRoute `json:"data"`
}

// LookingGlassBGPSessionsResponse represents the API response for BGP sessions.
type LookingGlassBGPSessionsResponse struct {
	Message string                    `json:"message"`
	Terms   string                    `json:"terms"`
	Data    []*LookingGlassBGPSession `json:"data"`
}

// LookingGlassBGPNeighborRoutesResponse represents the API response for BGP neighbor routes.
type LookingGlassBGPNeighborRoutesResponse struct {
	Message string                          `json:"message"`
	Terms   string                          `json:"terms"`
	Data    []*LookingGlassBGPNeighborRoute `json:"data"`
}

// LookingGlassAsyncJobResponse represents the API response for an async job submission.
type LookingGlassAsyncJobResponse struct {
	Message string                `json:"message"`
	Terms   string                `json:"terms"`
	Data    *LookingGlassAsyncJob `json:"data"`
}

// LookingGlassAsyncIPRoutesResponse represents the API response for async IP routes result.
type LookingGlassAsyncIPRoutesResponse struct {
	Message string             `json:"message"`
	Terms   string             `json:"terms"`
	Data    *AsyncIPRoutesData `json:"data"`
}

// AsyncIPRoutesData contains the async job metadata and routes.
type AsyncIPRoutesData struct {
	JobID  string                  `json:"jobId"`
	Status LookingGlassAsyncStatus `json:"status"`
	Routes []*LookingGlassIPRoute  `json:"routes"`
}

// LookingGlassAsyncBGPNeighborRoutesResponse represents the API response for async BGP neighbor routes result.
type LookingGlassAsyncBGPNeighborRoutesResponse struct {
	Message string                      `json:"message"`
	Terms   string                      `json:"terms"`
	Data    *AsyncBGPNeighborRoutesData `json:"data"`
}

// AsyncBGPNeighborRoutesData contains the async job metadata and neighbor routes.
type AsyncBGPNeighborRoutesData struct {
	JobID  string                          `json:"jobId"`
	Status LookingGlassAsyncStatus         `json:"status"`
	Routes []*LookingGlassBGPNeighborRoute `json:"routes"`
}
