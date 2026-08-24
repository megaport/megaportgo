package megaport

import "errors"

// LookingGlassRouteVXCRef identifies the VXC a diagnostics next hop resolves through.
type LookingGlassRouteVXCRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LookingGlassRouteNextHop is the next hop for a route in the MCR routing table.
type LookingGlassRouteNextHop struct {
	IP  string                  `json:"ip"`
	VXC LookingGlassRouteVXCRef `json:"vxc"`
}

// LookingGlassIPRoute is one entry in the MCR IP routing table.
type LookingGlassIPRoute struct {
	Prefix   string                   `json:"prefix"`
	Protocol string                   `json:"protocol"`
	Distance int                      `json:"distance,omitempty"`
	Metric   int                      `json:"metric,omitempty"`
	NextHop  LookingGlassRouteNextHop `json:"nextHop"`
}

// LookingGlassBGPRoute is one BGP route. The BGP routes and the BGP neighbor
// routes endpoints both return this shape.
type LookingGlassBGPRoute struct {
	Prefix       string                   `json:"prefix"`
	ASPath       string                   `json:"asPath,omitempty"`
	Origin       string                   `json:"origin,omitempty"`
	Source       string                   `json:"source,omitempty"`
	LocalPref    int                      `json:"localPref,omitempty"`
	MED          int                      `json:"med,omitempty"`
	Weight       int                      `json:"weight,omitempty"`
	Best         bool                     `json:"best,omitempty"`
	External     bool                     `json:"external,omitempty"`
	Valid        bool                     `json:"valid,omitempty"`
	Since        string                   `json:"since,omitempty"`
	Communities  []string                 `json:"communities,omitempty"`
	AdvertisedTo []string                 `json:"advertisedTo,omitempty"`
	NextHop      LookingGlassRouteNextHop `json:"nextHop"`
}

// ListIPRoutesRequest is a request for the MCR IP routing table.
type ListIPRoutesRequest struct {
	MCRID    string // The MCR UID
	IPFilter string // Optional: only return routes for this IP address or prefix
}

// ListBGPRoutesRequest is a request for the BGP routes the MCR knows about.
type ListBGPRoutesRequest struct {
	MCRID    string // The MCR UID
	IPFilter string // Optional: only return routes for this IP address or prefix
}

// ListBGPNeighborRoutesRequest is a request for the routes exchanged with one BGP peer.
type ListBGPNeighborRoutesRequest struct {
	MCRID         string // The MCR UID
	PeerIPAddress string // The BGP peer IP address
	Direction     string // BGPRouteDirectionReceived or BGPRouteDirectionAdvertised
}

// mcrDiagnosticsRoutesResponse is the API envelope for a route diagnostics result.
type mcrDiagnosticsRoutesResponse[T any] struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    []*T   `json:"data"`
}

// MCRPingRequest represents a request to ping a destination from an MCR.
type MCRPingRequest struct {
	MCRID              string
	DestinationAddress string // required
	SourceAddress      string // optional
	PacketCount        *int32 // optional, 1-60
	PacketSize         *int32 // optional, 1-9186
}

// MCRTracerouteRequest represents a request to traceroute from an MCR.
type MCRTracerouteRequest struct {
	MCRID              string
	DestinationAddress string // required
	SourceAddress      string // optional
}

// LookingGlassPingStatistics holds RTT and packet stats from a ping.
type LookingGlassPingStatistics struct {
	Duplicates         int     `json:"duplicates"`
	Errors             int     `json:"errors"`
	PacketLossPct      float64 `json:"packetLossPct"`
	PacketsReceived    int     `json:"packetsReceived"`
	PacketsTransmitted int     `json:"packetsTransmitted"`
	RTTAvgMs           float64 `json:"rttAvgMs"`
	RTTMaxMs           float64 `json:"rttMaxMs"`
	RTTMdevMs          float64 `json:"rttMdevMs"`
	RTTMinMs           float64 `json:"rttMinMs"`
	TotalTimeMs        float64 `json:"totalTimeMs"`
}

// LookingGlassPingResult is the result of a ping operation.
type LookingGlassPingResult struct {
	RawOutput  string                      `json:"rawOutput,omitempty"`
	Statistics *LookingGlassPingStatistics `json:"statistics,omitempty"`
}

// LookingGlassTracerouteProbe is a single probe result within a traceroute hop.
type LookingGlassTracerouteProbe struct {
	Host  string  `json:"host,omitempty"`
	RTTMs float64 `json:"rttMs,omitempty"`
}

// LookingGlassTracerouteHop is one hop in a traceroute result.
type LookingGlassTracerouteHop struct {
	Hop    string                         `json:"hop"`
	Probes []*LookingGlassTracerouteProbe `json:"probes"`
}

// LookingGlassTracerouteResult is the result of a traceroute operation.
type LookingGlassTracerouteResult struct {
	RawOutput string                       `json:"rawOutput,omitempty"`
	Hops      []*LookingGlassTracerouteHop `json:"hops"`
}

// mcrDiagnosticsStringResponse is the API envelope for diagnostic submit responses.
type mcrDiagnosticsStringResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    string `json:"data"`
}

// mcrDiagnosticsPingResultResponse is the API envelope for ping operation poll responses.
type mcrDiagnosticsPingResultResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    *LookingGlassPingResult `json:"data"`
}

// mcrDiagnosticsTracerouteResultResponse is the API envelope for traceroute operation poll responses.
type mcrDiagnosticsTracerouteResultResponse struct {
	Message string                        `json:"message"`
	Terms   string                        `json:"terms"`
	Data    *LookingGlassTracerouteResult `json:"data"`
}

// Errors for MCR diagnostics operations.
var (
	ErrMCRPingDestinationRequired       = errors.New("destination address is required")
	ErrMCRPingPacketCountOutOfRange     = errors.New("packet_count must be between 1 and 60")
	ErrMCRPingPacketSizeOutOfRange      = errors.New("packet_size must be between 1 and 9186")
	ErrMCRTracerouteDestinationRequired = errors.New("destination address is required")
	ErrMCRDiagnosticsMCRUIDRequired     = errors.New("MCR UID is required")
	ErrMCRDiagnosticsOperationEmpty     = errors.New("operation ID is required")
	ErrMCRDiagnosticsPeerIPRequired     = errors.New("BGP neighbor diagnostics require a peer IP address")
	ErrMCRDiagnosticsDirectionInvalid   = errors.New("BGP neighbor diagnostics require a direction (RECEIVED or ADVERTISED)")
	ErrMCRDiagnosticsTimeout            = errors.New("timed out waiting for diagnostics operation to complete")
)
