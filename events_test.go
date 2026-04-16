package megaport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

// EventsTestSuite tests the Events service.
type EventsTestSuite struct {
	ClientTestSuite
}

func TestEventsTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(EventsTestSuite))
}

func (suite *EventsTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *EventsTestSuite) TearDownTest() {
	suite.server.Close()
}

func (suite *EventsTestSuite) TestGetMaintenanceEvents() {
	sampleJSON := `{
        "message": "Maintenance events retrieved successfully",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": [
            {
                "eventId": "CSS-1234",
                "state": "Scheduled",
                "startTime": "2024-05-24T09:12:00.000Z",
                "endTime": "2024-05-24T09:42:00.000Z",
                "impact": "There will be minor impact on services.",
                "purpose": "Services will become more effective",
                "eventType": "Emergency",
                "services": [
                    "f06c80bc",
                    "0746e9a3"
                ]
            },
            {
                "eventId": "CSS-1235",
                "state": "Cancelled",
                "startTime": "2024-05-24T09:12:00.000Z",
                "endTime": "2024-05-24T09:42:00.000Z",
                "impact": "There will be minor impact on services.",
                "purpose": "Services will become more effective",
                "cancelReason": "Not Needed",
                "eventType": "Emergency",
                "services": [
                    "f06c80bc",
                    "0746e9a3"
                ]
            }
        ]
    }`

	suite.mux.HandleFunc("/ens/v1/status/maintenance", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal("Scheduled", r.URL.Query().Get("state"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(sampleJSON))
		if err != nil {
			suite.FailNowf("Failed to write response", "Error: %v", err)
		}
	})

	events, err := suite.client.EventsService.GetMaintenanceEvents(context.Background(), "Scheduled")
	suite.NoError(err)
	suite.Len(events, 2)
	suite.Equal("CSS-1234", events[0].EventID)
	suite.Equal("Not Needed", events[1].CancelReason)
}

func (suite *EventsTestSuite) TestGetMaintenanceEvents_CaseInsensitive() {
	suite.mux.HandleFunc("/ens/v1/status/maintenance", func(w http.ResponseWriter, r *http.Request) {
		// canonical casing must be sent regardless of input casing
		suite.Equal("Scheduled", r.URL.Query().Get("state"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message":"","terms":"","data":[]}`))
		if err != nil {
			suite.FailNowf("Failed to write response", "Error: %v", err)
		}
	})

	events, err := suite.client.EventsService.GetMaintenanceEvents(context.Background(), "scheduled")
	suite.NoError(err)
	suite.Empty(events)
}

func (suite *EventsTestSuite) TestGetMaintenanceEvents_InvalidState() {
	_, err := suite.client.EventsService.GetMaintenanceEvents(context.Background(), "invalid-state")
	suite.ErrorIs(err, ErrInvalidMaintenanceState)
}

func (suite *EventsTestSuite) TestGetMaintenanceEvents_HTTPError() {
	suite.mux.HandleFunc("/ens/v1/status/maintenance", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := suite.client.EventsService.GetMaintenanceEvents(context.Background(), "Scheduled")
	suite.Error(err)
}

func (suite *EventsTestSuite) TestGetMaintenanceEvents_MalformedJSON() {
	suite.mux.HandleFunc("/ens/v1/status/maintenance", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{not valid json`))
		suite.Require().NoError(err)
	})

	_, err := suite.client.EventsService.GetMaintenanceEvents(context.Background(), "Scheduled")
	suite.Error(err)
}

func (suite *EventsTestSuite) TestGetOutageEvents() {
	sampleJSON := `{
        "message": "Outage events retrieved successfully",
        "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
        "data": [
            {
                "outageId": "c2037361-eb5b-48a3-9c73-fb4efbf2c886",
                "state": "Ongoing",
                "eventId": "CSS-1234",
                "purpose": "Due to high CPU Usage, service outage occurred",
                "startTime": "2024-05-22T09:08:00.000Z",
                "createdBy": "john.cena@fibre.com",
                "createdDate": "2024-05-22T13:39:32.468Z",
                "updatedDate": "2024-05-22T13:39:32.468Z",
                "services": [],
                "notices": []
            },
            {
                "outageId": "ce0dd76b-655c-425f-923f-af5ae896756f",
                "state": "Ongoing",
                "eventId": "CSS-12345",
                "purpose": "This happened because something broke",
                "startTime": "2024-05-23T08:32:00.000Z",
                "createdBy": "john.cena@fibre.com",
                "createdDate": "2024-05-23T13:02:30.968Z",
                "updatedDate": "2024-05-23T13:02:30.968Z",
                "services": [],
                "notices": []
            }
        ]
    }`

	suite.mux.HandleFunc("/ens/v1/status/outage", func(w http.ResponseWriter, r *http.Request) {
		suite.Equal("Ongoing", r.URL.Query().Get("state"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(sampleJSON))
		if err != nil {
			suite.FailNowf("Failed to write response", "Error: %v", err)
		}
	})

	events, err := suite.client.EventsService.GetOutageEvents(context.Background(), "Ongoing")
	suite.NoError(err)
	suite.Len(events, 2)
	suite.Equal("c2037361-eb5b-48a3-9c73-fb4efbf2c886", events[0].OutageID)
	suite.Equal("CSS-12345", events[1].EventID)
}

func (suite *EventsTestSuite) TestGetOutageEvents_CaseInsensitive() {
	suite.mux.HandleFunc("/ens/v1/status/outage", func(w http.ResponseWriter, r *http.Request) {
		// canonical casing must be sent regardless of input casing
		suite.Equal("Ongoing", r.URL.Query().Get("state"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"message":"","terms":"","data":[]}`))
		if err != nil {
			suite.FailNowf("Failed to write response", "Error: %v", err)
		}
	})

	events, err := suite.client.EventsService.GetOutageEvents(context.Background(), "ongoing")
	suite.NoError(err)
	suite.Empty(events)
}

func (suite *EventsTestSuite) TestGetOutageEvents_InvalidState() {
	_, err := suite.client.EventsService.GetOutageEvents(context.Background(), "invalid-state")
	suite.ErrorIs(err, ErrInvalidOutageState)
}

func (suite *EventsTestSuite) TestGetOutageEvents_MalformedJSON() {
	suite.mux.HandleFunc("/ens/v1/status/outage", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{not valid json`))
		suite.Require().NoError(err)
	})

	_, err := suite.client.EventsService.GetOutageEvents(context.Background(), "Ongoing")
	suite.Error(err)
}

func (suite *EventsTestSuite) TestGetOutageEvents_HTTPError() {
	suite.mux.HandleFunc("/ens/v1/status/outage", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := suite.client.EventsService.GetOutageEvents(context.Background(), "Ongoing")
	suite.Error(err)
}
