package megaport

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// NATGatewaySession represents a speed/session-count availability entry for NAT Gateways.
type NATGatewaySession struct {
	SessionCount []int `json:"sessionCount"`
	SpeedMbps    int   `json:"speedMbps"`
}

// NATGatewaySessionsResponse is the API response for listing NAT Gateway sessions.
type NATGatewaySessionsResponse struct {
	Message string               `json:"message"`
	Terms   string               `json:"terms"`
	Data    []*NATGatewaySession `json:"data"`
}

// ServiceTelemetryResponse is the API response for service telemetry data.
// This response is NOT wrapped in the standard message/terms/data envelope.
type ServiceTelemetryResponse struct {
	ServiceUID string                 `json:"serviceUid"`
	Type       string                 `json:"type"`
	TimeFrame  TelemetryTimeFrame     `json:"timeFrame"`
	Data       []*TelemetryMetricData `json:"data"`
}

// TelemetryTimeFrame represents the time range of a telemetry response.
type TelemetryTimeFrame struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

// TelemetryMetricData represents a single metric series in a telemetry response.
type TelemetryMetricData struct {
	Type    string            `json:"type"`
	Subtype string            `json:"subtype"`
	Samples []TelemetrySample `json:"samples"`
	Unit    TelemetryUnit     `json:"unit"`
}

// TelemetrySample represents a single data point in a telemetry series.
// The API returns samples as [timestamp, value] tuples.
type TelemetrySample struct {
	Timestamp int64
	Value     float64
}

// UnmarshalJSON handles the [int64, float64] tuple format from the API.
func (s *TelemetrySample) UnmarshalJSON(data []byte) error {
	var tuple []json.Number
	if err := json.Unmarshal(data, &tuple); err != nil {
		return fmt.Errorf("telemetry sample must be a JSON array: %w", err)
	}
	if len(tuple) != 2 {
		return fmt.Errorf("telemetry sample must have exactly 2 elements, got %d", len(tuple))
	}
	ts, err := tuple[0].Int64()
	if err != nil {
		return fmt.Errorf("telemetry sample timestamp: %w", err)
	}
	val, err := tuple[1].Float64()
	if err != nil {
		return fmt.Errorf("telemetry sample value: %w", err)
	}
	s.Timestamp = ts
	s.Value = val
	return nil
}

// TelemetryUnit describes the unit of measurement for a telemetry metric.
type TelemetryUnit struct {
	Name     string `json:"name"`
	FullName string `json:"fullName"`
}

// NATGateway represents a NAT Gateway product from the Megaport API.
type NATGateway struct {
	AdminLocked           bool                    `json:"adminLocked"`
	AutoRenewTerm         bool                    `json:"autoRenewTerm"`
	Config                NATGatewayNetworkConfig `json:"config"`
	ContractEndDate       string                  `json:"contractEndDate"`
	CreateDate            string                  `json:"createDate"`
	CreatedBy             string                  `json:"createdBy"`
	LocationID            int                     `json:"locationId"`
	Locked                bool                    `json:"locked"`
	OrderApprovalStatus   string                  `json:"orderApprovalStatus"`
	ProductName           string                  `json:"productName"`
	ProductUID            string                  `json:"productUid"`
	PromoCode             string                  `json:"promoCode"`
	ProvisioningStatus    string                  `json:"provisioningStatus"`
	ResourceTags          []ResourceTag           `json:"resourceTags"`
	ServiceLevelReference string                  `json:"serviceLevelReference"`
	Speed                 int                     `json:"speed"`
	Term                  int                     `json:"term"`
}

// NATGatewayNetworkConfig represents the network configuration for a NAT Gateway.
type NATGatewayNetworkConfig struct {
	ASN                int    `json:"asn"`
	BGPShutdownDefault bool   `json:"bgpShutdownDefault"`
	DiversityZone      string `json:"diversityZone"`
	SessionCount       int    `json:"sessionCount"`
}

// CreateNATGatewayRequest represents a request to create a NAT Gateway.
type CreateNATGatewayRequest struct {
	AutoRenewTerm         bool                    `json:"autoRenewTerm"`
	Config                NATGatewayNetworkConfig `json:"config"`
	LocationID            int                     `json:"locationId"`
	ProductName           string                  `json:"productName"`
	PromoCode             string                  `json:"promoCode,omitempty"`
	ResourceTags          []ResourceTag           `json:"resourceTags,omitempty"`
	ServiceLevelReference string                  `json:"serviceLevelReference,omitempty"`
	Speed                 int                     `json:"speed"`
	Term                  int                     `json:"term"`
}

// UpdateNATGatewayRequest represents a request to update a NAT Gateway.
type UpdateNATGatewayRequest struct {
	ProductUID            string                  `json:"-"` // path parameter, not serialized
	AutoRenewTerm         bool                    `json:"autoRenewTerm"`
	Config                NATGatewayNetworkConfig `json:"config"`
	LocationID            int                     `json:"locationId"`
	ProductName           string                  `json:"productName"`
	PromoCode             string                  `json:"promoCode,omitempty"`
	ResourceTags          []ResourceTag           `json:"resourceTags,omitempty"`
	ServiceLevelReference string                  `json:"serviceLevelReference,omitempty"`
	Speed                 int                     `json:"speed"`
	Term                  int                     `json:"term"`
}

// NATGatewayResponse is the API response for a single NAT Gateway.
type NATGatewayResponse struct {
	Message string     `json:"message"`
	Terms   string     `json:"terms"`
	Data    NATGateway `json:"data"`
}

// NATGatewayListResponse is the API response for listing NAT Gateways.
type NATGatewayListResponse struct {
	Message string        `json:"message"`
	Terms   string        `json:"terms"`
	Data    []*NATGateway `json:"data"`
}

// DeleteNATGatewayResponse is the API response for deleting a NAT Gateway.
type DeleteNATGatewayResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
}

// NATGatewayOrderPrice captures the pricing preview returned by
// POST /v3/networkdesign/validate for a NAT Gateway order.
type NATGatewayOrderPrice struct {
	HourlySetup          float64 `json:"hourlySetup"`
	DailySetup           float64 `json:"dailySetup"`
	MonthlySetup         float64 `json:"monthlySetup"`
	HourlyRate           float64 `json:"hourlyRate"`
	DailyRate            float64 `json:"dailyRate"`
	MonthlyRate          float64 `json:"monthlyRate"`
	FixedRecurringCharge float64 `json:"fixedRecurringCharge"`
	LongHaulMbpsRate     float64 `json:"longHaulMbpsRate"`
	MbpsRate             float64 `json:"mbpsRate"`
	Currency             string  `json:"currency"`
	ProductType          string  `json:"productType"`
	MonthlyRackRate      float64 `json:"monthlyRackRate"`
}

// NATGatewayValidateResult is a single entry returned by
// POST /v3/networkdesign/validate.
type NATGatewayValidateResult struct {
	ProductUID  string `json:"productUid"`
	ProductType string `json:"productType"`
	// Metro is the metro/city name returned by the API for the gateway's
	// location (e.g. "Sydney"). The API ships this as a field literally
	// named "string" in the JSON response, hence the unusual json tag.
	Metro string               `json:"string"`
	Price NATGatewayOrderPrice `json:"price"`
}

// NATGatewayBuyResult is a single entry returned by
// POST /v3/networkdesign/buy after a NAT Gateway design is purchased.
type NATGatewayBuyResult struct {
	ProductUID         string `json:"uid"`
	ProductName        string `json:"name"`
	ServiceName        string `json:"serviceName"`
	ProductType        string `json:"productType"`
	ProvisioningStatus string `json:"provisioningStatus"`
	RateLimit          int    `json:"rateLimit"`
	LocationID         int    `json:"aLocationId"`
	ContractTermMonths int    `json:"contractTermMonths"`
	CreateDate         int64  `json:"createDate"`
}

// natGatewayValidateEnvelope is the API response envelope for validate.
type natGatewayValidateEnvelope struct {
	Message string                      `json:"message"`
	Terms   string                      `json:"terms"`
	Data    []*NATGatewayValidateResult `json:"data"`
}

// natGatewayBuyEnvelope is the API response envelope for buy.
type natGatewayBuyEnvelope struct {
	Message string                 `json:"message"`
	Terms   string                 `json:"terms"`
	Data    []*NATGatewayBuyResult `json:"data"`
}

// --- Packet filters -------------------------------------------------------

// Packet filter action values accepted by the API.
const (
	PacketFilterActionPermit = "permit"
	PacketFilterActionDeny   = "deny"
)

// NATGatewayPacketFilterRequest is the create/update payload for a packet
// filter on a NAT Gateway.
type NATGatewayPacketFilterRequest struct {
	Description string                        `json:"description"`
	Entries     []NATGatewayPacketFilterEntry `json:"entries"`
}

// NATGatewayPacketFilterEntry is a single rule inside a packet filter.
// Entries are evaluated in order; the first matching entry determines the
// action taken on the packet.
type NATGatewayPacketFilterEntry struct {
	Action             string `json:"action"` // PacketFilterActionPermit or PacketFilterActionDeny.
	Description        string `json:"description,omitempty"`
	SourceAddress      string `json:"sourceAddress"`
	DestinationAddress string `json:"destinationAddress"`
	SourcePorts        string `json:"sourcePorts,omitempty"`
	DestinationPorts   string `json:"destinationPorts,omitempty"`
	IPProtocol         int    `json:"ipProtocol,omitempty"`
}

// NATGatewayPacketFilter is a server-side packet filter including its
// assigned ID.
type NATGatewayPacketFilter struct {
	ID int `json:"id"`
	NATGatewayPacketFilterRequest
}

// NATGatewayPacketFilterSummary is the compact entry returned by the
// packet_filter_summaries endpoint.
type NATGatewayPacketFilterSummary struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}

// natGatewayPacketFilterResponse is the API envelope for create/get/update.
type natGatewayPacketFilterResponse struct {
	Message string                  `json:"message"`
	Terms   string                  `json:"terms"`
	Data    *NATGatewayPacketFilter `json:"data"`
}

// natGatewayPacketFilterSummariesResponse is the API envelope for the
// summaries list endpoint.
type natGatewayPacketFilterSummariesResponse struct {
	Message string                           `json:"message"`
	Terms   string                           `json:"terms"`
	Data    []*NATGatewayPacketFilterSummary `json:"data"`
}

// --- Prefix lists ---------------------------------------------------------

// Prefix list action values accepted by the API.
const (
	PrefixListActionPermit = "permit"
	PrefixListActionDeny   = "deny"
)

// Address family values accepted by the API.
const (
	AddressFamilyIPv4 = "IPv4"
	AddressFamilyIPv6 = "IPv6"
)

// NATGatewayPrefixList is the create/update/get payload for a prefix list on
// a NAT Gateway. The API returns the server-assigned ID on read.
type NATGatewayPrefixList struct {
	ID            int                         `json:"id,omitempty"`
	Description   string                      `json:"description"`
	AddressFamily string                      `json:"addressFamily"` // AddressFamilyIPv4 or AddressFamilyIPv6.
	Entries       []NATGatewayPrefixListEntry `json:"entries"`
}

// NATGatewayPrefixListEntry is a single entry in a prefix list. Ge/Le are
// exposed as ints for ergonomics; the SDK converts to/from the API's string
// representation transparently.
type NATGatewayPrefixListEntry struct {
	Action string `json:"action"` // PrefixListActionPermit or PrefixListActionDeny.
	Prefix string `json:"prefix"`
	Ge     int    `json:"ge,omitempty"`
	Le     int    `json:"le,omitempty"`
}

// NATGatewayPrefixListSummary is the compact entry returned by the
// prefix_list_summaries endpoint.
type NATGatewayPrefixListSummary struct {
	ID            int    `json:"id"`
	Description   string `json:"description"`
	AddressFamily string `json:"addressFamily"`
}

// apiNATGatewayPrefixList is the wire-level representation — the API sends
// Ge/Le as strings. See (NATGatewayPrefixList).toAPI / fromAPI for
// conversion.
type apiNATGatewayPrefixList struct {
	ID            int                            `json:"id,omitempty"`
	Description   string                         `json:"description"`
	AddressFamily string                         `json:"addressFamily"`
	Entries       []apiNATGatewayPrefixListEntry `json:"entries"`
}

type apiNATGatewayPrefixListEntry struct {
	Action string `json:"action"`
	Prefix string `json:"prefix"`
	Ge     string `json:"ge,omitempty"`
	Le     string `json:"le,omitempty"`
}

// toAPI converts the user-facing NATGatewayPrefixList to its wire-level
// representation (Ge/Le as strings, zero values omitted).
func (p *NATGatewayPrefixList) toAPI() *apiNATGatewayPrefixList {
	out := &apiNATGatewayPrefixList{
		ID:            p.ID,
		Description:   p.Description,
		AddressFamily: p.AddressFamily,
		Entries:       make([]apiNATGatewayPrefixListEntry, len(p.Entries)),
	}
	for i, e := range p.Entries {
		apiEntry := apiNATGatewayPrefixListEntry{
			Action: e.Action,
			Prefix: e.Prefix,
		}
		if e.Ge > 0 {
			apiEntry.Ge = strconv.Itoa(e.Ge)
		}
		if e.Le > 0 {
			apiEntry.Le = strconv.Itoa(e.Le)
		}
		out.Entries[i] = apiEntry
	}
	return out
}

// fromAPI converts a wire-level apiNATGatewayPrefixList into the user-facing
// NATGatewayPrefixList. Non-numeric Ge/Le strings produce an error.
func (a *apiNATGatewayPrefixList) toPrefixList() (*NATGatewayPrefixList, error) {
	out := &NATGatewayPrefixList{
		ID:            a.ID,
		Description:   a.Description,
		AddressFamily: a.AddressFamily,
		Entries:       make([]NATGatewayPrefixListEntry, len(a.Entries)),
	}
	for i, e := range a.Entries {
		entry := NATGatewayPrefixListEntry{Action: e.Action, Prefix: e.Prefix}
		if e.Ge != "" {
			ge, err := strconv.Atoi(e.Ge)
			if err != nil {
				return nil, fmt.Errorf("prefix list entry %d: invalid ge %q: %w", i, e.Ge, err)
			}
			entry.Ge = ge
		}
		if e.Le != "" {
			le, err := strconv.Atoi(e.Le)
			if err != nil {
				return nil, fmt.Errorf("prefix list entry %d: invalid le %q: %w", i, e.Le, err)
			}
			entry.Le = le
		}
		out.Entries[i] = entry
	}
	return out, nil
}

// natGatewayPrefixListResponse is the API envelope for create/get/update.
type natGatewayPrefixListResponse struct {
	Message string                   `json:"message"`
	Terms   string                   `json:"terms"`
	Data    *apiNATGatewayPrefixList `json:"data"`
}

// natGatewayPrefixListSummariesResponse is the API envelope for the
// summaries list endpoint.
type natGatewayPrefixListSummariesResponse struct {
	Message string                         `json:"message"`
	Terms   string                         `json:"terms"`
	Data    []*NATGatewayPrefixListSummary `json:"data"`
}

// --- Diagnostics (async looking-glass) -----------------------------------

// BGP route direction values for the BGP neighbor endpoint.
const (
	BGPRouteDirectionReceived   = "RECEIVED"
	BGPRouteDirectionAdvertised = "ADVERTISED"
)

// NATGatewayBGPNeighborRoutesRequest contains the parameters for the BGP
// neighbor diagnostics endpoint.
type NATGatewayBGPNeighborRoutesRequest struct {
	ProductUID    string
	PeerIPAddress string
	Direction     string // BGPRouteDirectionReceived or BGPRouteDirectionAdvertised.
}

// NATGatewayRouteVXCRef identifies the VXC that carries a next-hop IP in a
// looking-glass response.
type NATGatewayRouteVXCRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NATGatewayRouteNextHop describes the next hop for a diagnostics route.
type NATGatewayRouteNextHop struct {
	IP  string                `json:"ip"`
	VXC NATGatewayRouteVXCRef `json:"vxc"`
}

// NATGatewayIPRoute is a single IP route returned by the diagnostics IP
// routes endpoint.
type NATGatewayIPRoute struct {
	Prefix   string                 `json:"prefix"`
	Protocol string                 `json:"protocol"`
	Distance int                    `json:"distance,omitempty"`
	Metric   int                    `json:"metric,omitempty"`
	NextHop  NATGatewayRouteNextHop `json:"nextHop"`
}

// NATGatewayBGPRoute is a single BGP route returned by the diagnostics BGP
// and BGP neighbor endpoints.
type NATGatewayBGPRoute struct {
	Prefix       string                 `json:"prefix"`
	ASPath       string                 `json:"asPath,omitempty"`
	Origin       string                 `json:"origin,omitempty"`
	Source       string                 `json:"source,omitempty"`
	LocalPref    int                    `json:"localPref,omitempty"`
	MED          int                    `json:"med,omitempty"`
	Best         bool                   `json:"best,omitempty"`
	External     bool                   `json:"external,omitempty"`
	Since        string                 `json:"since,omitempty"`
	Communities  []string               `json:"communities,omitempty"`
	AdvertisedTo []string               `json:"advertisedTo,omitempty"`
	NextHop      NATGatewayRouteNextHop `json:"nextHop"`
}

// NATGatewayRoute is a discriminated wrapper for looking-glass routes. The
// async operation endpoint returns a heterogeneous list of IP and BGP
// routes; exactly one of IP / BGP will be set per entry.
type NATGatewayRoute struct {
	IP  *NATGatewayIPRoute
	BGP *NATGatewayBGPRoute
}

// UnmarshalJSON distinguishes IP vs BGP routes based on which BGP-specific
// fields are present in the payload.
func (r *NATGatewayRoute) UnmarshalJSON(b []byte) error {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(b, &probe); err != nil {
		return err
	}
	// BGP-specific fields present in LookingGlassBgpRoute but not
	// LookingGlassIpRoute.
	_, hasAsPath := probe["asPath"]
	_, hasLocalPref := probe["localPref"]
	_, hasBest := probe["best"]
	_, hasOrigin := probe["origin"]
	if hasAsPath || hasLocalPref || hasBest || hasOrigin {
		var bgp NATGatewayBGPRoute
		if err := json.Unmarshal(b, &bgp); err != nil {
			return err
		}
		r.BGP = &bgp
		return nil
	}
	var ip NATGatewayIPRoute
	if err := json.Unmarshal(b, &ip); err != nil {
		return err
	}
	r.IP = &ip
	return nil
}

// natGatewayDiagnosticsAsyncResponse is the API envelope returned by the
// three diagnostics list endpoints — Data contains the operationId UUID.
type natGatewayDiagnosticsAsyncResponse struct {
	Message string `json:"message"`
	Terms   string `json:"terms"`
	Data    string `json:"data"`
}

// natGatewayDiagnosticsRoutesResponse is the envelope returned by the
// operation endpoint.
type natGatewayDiagnosticsRoutesResponse struct {
	Message string             `json:"message"`
	Terms   string             `json:"terms"`
	Data    []*NATGatewayRoute `json:"data"`
}
