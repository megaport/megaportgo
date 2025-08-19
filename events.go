package megaport

import (
	"encoding/json"
	"net/http"
	"strings"
)

type EventsService interface {
	// GetMaintenanceEvents returns details about maintenance events, filtered by the specified state value.
	GetMaintenanceEvents(state string) ([]MaintenanceEvent, error)
	// GetOutageEvents returns details about outage events, filtered by the specified state value.
	GetOutageEvents(state string) ([]OutageEvent, error)
}

type EventsServiceOp struct {
	client *Client
}

func NewEventsService(client *Client) EventsService {
	return &EventsServiceOp{
		client: client,
	}
}

type MaintenanceState string
type OutageState string

var (
	VALID_MAINTENANCE_STATES = []MaintenanceState{
		MAINTENANCE_STATE_COMPLETED,
		MAINTENANCE_STATE_SCHEDULED,
		MAINTENANCE_STATE_CANCELLED,
		MAINTENANCE_STATE_RUNNING,
	}
	VALID_OUTAGE_STATES = []OutageState{
		OUTAGE_STATE_ONGOING,
		OUTAGE_STATE_RESOLVED,
	}
)

const (
	MAINTENANCE_STATE_COMPLETED = MaintenanceState("Completed")
	MAINTENANCE_STATE_SCHEDULED = MaintenanceState("Scheduled")
	MAINTENANCE_STATE_CANCELLED = MaintenanceState("Cancelled")
	MAINTENANCE_STATE_RUNNING   = MaintenanceState("Running")
	OUTAGE_STATE_ONGOING        = OutageState("Ongoing")
	OUTAGE_STATE_RESOLVED       = OutageState("Resolved")
)

// MaintenanceEvent represents a maintenance event returned by the Events API.
//
// Returns details about maintenance events, filtered by the specified state value.
//
// The following information is returned for maintenance events in the response, with some fields being optional and only included under certain conditions.
type MaintenanceEvent struct {
	// EventID is the ticket number against which a particular event is created.
	EventID string `json:"eventId"`

	// State is the current state of the event.
	State string `json:"state"`

	// StartTime is the event start time in ISO 8601 UTC format (yyyy-MM-dd'T'HH:mm:ss.SSSX).
	StartTime string `json:"startTime"`

	// EndTime is the event end time in ISO 8601 UTC format (yyyy-MM-dd'T'HH:mm:ss.SSSX).
	EndTime string `json:"endTime"`

	// Impact is the impact of the event on the services, if any.
	Impact string `json:"impact"`

	// Purpose is the reason why this event is created.
	Purpose string `json:"purpose"`

	// CancelReason is returned if the event is canceled, stating the cancellation reason.
	CancelReason string `json:"cancelReason"`

	// EventType is "Emergency" if the event is created on short notice; otherwise, it is a "Planned" event.
	EventType string `json:"eventType"`

	// ServiceIDs is the list of services affected by the event, containing the short UUIDs of the services.
	ServiceIDs []string `json:"services"`
}

// OutageEvent represents an outage event returned by the Events API.
//
// The following information is returned for outage events in the response, with some fields being optional and only included under certain conditions.
type OutageEvent struct {
	// OutageID is a unique identifier for each outage event.
	OutageID string `json:"outageId"`

	// EventID is the ticket number against which a particular event is created.
	EventID string `json:"eventId"`

	// State is the current state of the event.
	State string `json:"state"`

	// StartTime is the event start time in ISO 8601 UTC format (yyyy-MM-dd'T'HH:mm:ss.SSSX).
	StartTime string `json:"startTime"`

	// EndTime is the event end time in ISO 8601 UTC format (yyyy-MM-dd'T'HH:mm:ss.SSSX).
	EndTime string `json:"endTime"`

	// Purpose is the reason why this event is created.
	Purpose string `json:"purpose"`

	// Services is the list of services affected by the event, containing the short UUIDs of the services.
	Services []string `json:"services"`

	// RootCause is the reason explaining why an outage happened. This field is present only when an outage is resolved.
	RootCause string `json:"rootCause"`

	// Resolution explains the solution taken to resolve the outage. Present when an outage is resolved.
	Resolution string `json:"resolution"`

	// MitigationActions explains the steps taken to avoid such outages in the future. Present for resolved outages.
	MitigationActions string `json:"mitigationActions"`

	// CreatedBy is the user who created the outage.
	CreatedBy string `json:"createdBy"`

	// CreatedDate is the date and time when an outage event is created, in ISO 8601 UTC format (yyyy-MM-dd'T'HH:mm:ss.SSSX).
	CreatedDate string `json:"createdDate"`

	// UpdatedDate is the date and time when an outage event is updated, in ISO 8601 UTC format (yyyy-MM-dd'T'HH:mm:ss.SSSX).
	UpdatedDate string `json:"updatedDate"`

	// Notices is the list of notices sent as an update for an ongoing outage.
	Notices []string `json:"notices"`
}

// GetMaintenanceEvents retrieves maintenance events from the Megaport API, filtered by the specified state.
// It validates the state against the valid maintenance states and returns an error if invalid.
func (s *EventsServiceOp) GetMaintenanceEvents(state string) ([]MaintenanceEvent, error) {
	// Validate state
	valid := false
	for _, st := range VALID_MAINTENANCE_STATES {
		if strings.EqualFold(string(st), state) {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidMaintenanceState
	}

	// Build URL
	url := s.client.BaseURL.JoinPath("/ens/v1/status/maintenance").String() + "?state=" + state

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Perform request
	resp, err := s.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode response as an array of MaintenanceEvent
	var events []MaintenanceEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	return events, nil
}

// GetOutageEvents retrieves outage events from the Megaport API, filtered by the specified state.
// It validates the state against the valid outage states and returns an error if invalid.
func (s *EventsServiceOp) GetOutageEvents(state string) ([]OutageEvent, error) {
	// Validate state
	valid := false
	for _, st := range VALID_OUTAGE_STATES {
		if strings.EqualFold(string(st), state) {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidOutageState
	}

	// Build URL
	url := s.client.BaseURL.JoinPath("/ens/v1/status/outage").String() + "?state=" + state

	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Perform request
	resp, err := s.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode response as an array of OutageEvent
	var events []OutageEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	return events, nil
}
