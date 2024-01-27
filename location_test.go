package megaport

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/megaport/megaportgo/types"
	"github.com/stretchr/testify/require"
)

// ListLocations(ctx context.Context) ([]types.Location, error)
// GetLocationByID(ctx context.Context, locationID int) (*types.Location, error)
// GetLocationByName(ctx context.Context, locationName string) (*types.Location, error)
// GetLocationByNameFuzzy(ctx context.Context, search string) ([]types.Location, error)
// ListCountries(ctx context.Context) ([]types.Country, error)
// ListMarketCodes(ctx context.Context) ([]string, error)
// IsValidMarketCode(ctx context.Context, marketCode string) (*bool, error)
// FilterLocationsByMarketCode(ctx context.Context, marketCode string, locations *[]types.Location) error

func TestListLocations(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()
	locSvc := client.LocationService
	want := []*types.Location{
		{
			Name:          "Test Data Center",
			Country:       "USA",
			LiveDate:      1595340000000,
			SiteCode:      "denverTest",
			NetworkRegion: "MP1",
			Address: map[string]string{
				"street":   "Test Street Denver",
				"suburb":   "Test Suburb Denver",
				"city":     "Denver",
				"state":    "CO",
				"country":  "USA",
				"postcode": "80011",
			},
			Campus:    "campus_deprecated",
			Latitude:  39.762714,
			Longitude: -104.761925,
			Products: map[string]interface{}{
				"mcr":      false,
				"megaport": []interface{}{float64(1), float64(10)},
			},
			Market:           "US",
			Metro:            "Denver",
			VRouterAvailable: false,
			ID:               111,
			Status:           "Active",
		},
		{
			ID:            112,
			Name:          "Test Data Center 2",
			Campus:        "campus_deprecated",
			Metro:         "Ashburn",
			Country:       "USA",
			SiteCode:      "ashburnTest",
			NetworkRegion: "MP1",
			Address: map[string]string{
				"street":   "Test Street Ashburn",
				"suburb":   "Test Suburb Ashburn",
				"city":     "Ashburn",
				"state":    "VA",
				"country":  "USA",
				"postcode": "20146",
			},
			Market:           "US",
			VRouterAvailable: false,
			LiveDate:         1483711200000,
			Status:           "Active",
			Longitude:        -77.487442,
			Latitude:         39.043757,
			Products: map[string]interface{}{
				"mcr":      false,
				"megaport": []interface{}{float64(10)},
			},
		},
	}
	path := "/v2/locations"
	jblob := `{
    "message": "List all public locations",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [{
			"id": 111,
			"name": "Test Data Center",
			"campus": "campus_deprecated",
			"metro": "Denver",
			"country": "USA",
			"siteCode": "denverTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street Denver",
				"suburb": "Test Suburb Denver",
				"city": "Denver",
				"state": "CO",
				"country": "USA",
				"postcode": "80011"
			},
			"dc": {
				"id": 111,
				"name": "Test Data Center"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1595340000000,
			"status": "Active",
			"longitude": -104.761925,
			"latitude": 39.762714,
			"products": {
				"mcr": false,
				"megaport": [
					1,
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		},
		{
			"id": 112,
			"name": "Test Data Center 2",
			"campus": "campus_deprecated",
			"metro": "Ashburn",
			"country": "USA",
			"siteCode": "ashburnTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street Ashburn",
				"suburb": "Test Suburb Ashburn",
				"city": "Ashburn",
				"state": "VA",
				"country": "USA",
				"postcode": "20146"
			},
			"dc": {
				"id": 112,
				"name": "Test Data Center 2"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1483711200000,
			"status": "Active",
			"longitude": -77.487442,
			"latitude": 39.043757,
			"products": {
				"mcr": false,
				"megaport": [
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		}]
	}`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.ListLocations(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetLocationByID(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()
	locSvc := client.LocationService
	want := &types.Location{
		Name:          "Test Data Center",
		Country:       "USA",
		LiveDate:      1595340000000,
		SiteCode:      "denverTest",
		NetworkRegion: "MP1",
		Address: map[string]string{
			"street":   "Test Street Denver",
			"suburb":   "Test Suburb Denver",
			"city":     "Denver",
			"state":    "CO",
			"country":  "USA",
			"postcode": "80011",
		},
		Campus:    "campus_deprecated",
		Latitude:  39.762714,
		Longitude: -104.761925,
		Products: map[string]interface{}{
			"mcr":      false,
			"megaport": []interface{}{float64(1), float64(10)},
		},
		Market:           "US",
		Metro:            "Denver",
		VRouterAvailable: false,
		ID:               111,
		Status:           "Active",
	}
	path := "/v2/locations"
	jblob := `{
    "message": "List all public locations",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [{
			"id": 111,
			"name": "Test Data Center",
			"campus": "campus_deprecated",
			"metro": "Denver",
			"country": "USA",
			"siteCode": "denverTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street Denver",
				"suburb": "Test Suburb Denver",
				"city": "Denver",
				"state": "CO",
				"country": "USA",
				"postcode": "80011"
			},
			"dc": {
				"id": 111,
				"name": "Test Data Center"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1595340000000,
			"status": "Active",
			"longitude": -104.761925,
			"latitude": 39.762714,
			"products": {
				"mcr": false,
				"megaport": [
					1,
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		},
		{
			"id": 112,
			"name": "Test Data Center 2",
			"campus": "campus_deprecated",
			"metro": "Ashburn",
			"country": "USA",
			"siteCode": "ashburnTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street Ashburn",
				"suburb": "Test Suburb Ashburn",
				"city": "Ashburn",
				"state": "VA",
				"country": "USA",
				"postcode": "20146"
			},
			"dc": {
				"id": 112,
				"name": "Test Data Center 2"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1483711200000,
			"status": "Active",
			"longitude": -77.487442,
			"latitude": 39.043757,
			"products": {
				"mcr": false,
				"megaport": [
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		}]
	}`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.GetLocationByID(ctx, 111)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetLocationByName(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()
	locSvc := client.LocationService
	want := &types.Location{

		ID:            112,
		Name:          "Test Data Center 2",
		Campus:        "campus_deprecated",
		Metro:         "Ashburn",
		Country:       "USA",
		SiteCode:      "ashburnTest",
		NetworkRegion: "MP1",
		Address: map[string]string{
			"street":   "Test Street Ashburn",
			"suburb":   "Test Suburb Ashburn",
			"city":     "Ashburn",
			"state":    "VA",
			"country":  "USA",
			"postcode": "20146",
		},
		Market:           "US",
		VRouterAvailable: false,
		LiveDate:         1483711200000,
		Status:           "Active",
		Longitude:        -77.487442,
		Latitude:         39.043757,
		Products: map[string]interface{}{
			"mcr":      false,
			"megaport": []interface{}{float64(10)},
		},
	}
	path := "/v2/locations"
	jblob := `{
    "message": "List all public locations",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [{
			"id": 111,
			"name": "Test Data Center",
			"campus": "campus_deprecated",
			"metro": "Denver",
			"country": "USA",
			"siteCode": "denverTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street Denver",
				"suburb": "Test Suburb Denver",
				"city": "Denver",
				"state": "CO",
				"country": "USA",
				"postcode": "80011"
			},
			"dc": {
				"id": 111,
				"name": "Test Data Center"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1595340000000,
			"status": "Active",
			"longitude": -104.761925,
			"latitude": 39.762714,
			"products": {
				"mcr": false,
				"megaport": [
					1,
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		},
		{
			"id": 112,
			"name": "Test Data Center 2",
			"campus": "campus_deprecated",
			"metro": "Ashburn",
			"country": "USA",
			"siteCode": "ashburnTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street Ashburn",
				"suburb": "Test Suburb Ashburn",
				"city": "Ashburn",
				"state": "VA",
				"country": "USA",
				"postcode": "20146"
			},
			"dc": {
				"id": 112,
				"name": "Test Data Center 2"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1483711200000,
			"status": "Active",
			"longitude": -77.487442,
			"latitude": 39.043757,
			"products": {
				"mcr": false,
				"megaport": [
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		}]
	}`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.GetLocationByName(ctx, "Test Data Center 2")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetLocationByNameFuzzy(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()
	locSvc := client.LocationService
	want := []*types.Location{
		{
			Name:          "Test Data Center",
			Country:       "USA",
			LiveDate:      1595340000000,
			SiteCode:      "denverTest",
			NetworkRegion: "MP1",
			Address: map[string]string{
				"street":   "Test Street Denver",
				"suburb":   "Test Suburb Denver",
				"city":     "Denver",
				"state":    "CO",
				"country":  "USA",
				"postcode": "80011",
			},
			Campus:    "campus_deprecated",
			Latitude:  39.762714,
			Longitude: -104.761925,
			Products: map[string]interface{}{
				"mcr":      false,
				"megaport": []interface{}{float64(1), float64(10)},
			},
			Market:           "US",
			Metro:            "Denver",
			VRouterAvailable: false,
			ID:               111,
			Status:           "Active",
		},
		{
			ID:            112,
			Name:          "Test Data Center 2",
			Campus:        "campus_deprecated",
			Metro:         "Ashburn",
			Country:       "USA",
			SiteCode:      "ashburnTest",
			NetworkRegion: "MP1",
			Address: map[string]string{
				"street":   "Test Street Ashburn",
				"suburb":   "Test Suburb Ashburn",
				"city":     "Ashburn",
				"state":    "VA",
				"country":  "USA",
				"postcode": "20146",
			},
			Market:           "US",
			VRouterAvailable: false,
			LiveDate:         1483711200000,
			Status:           "Active",
			Longitude:        -77.487442,
			Latitude:         39.043757,
			Products: map[string]interface{}{
				"mcr":      false,
				"megaport": []interface{}{float64(10)},
			},
		},
	}
	path := "/v2/locations"
	jblob := `
{
    "message": "List all public locations",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [{
			"id": 111,
			"name": "Test Data Center",
			"campus": "campus_deprecated",
			"metro": "Denver",
			"country": "USA",
			"siteCode": "denverTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street Denver",
				"suburb": "Test Suburb Denver",
				"city": "Denver",
				"state": "CO",
				"country": "USA",
				"postcode": "80011"
			},
			"dc": {
				"id": 111,
				"name": "Test Data Center"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1595340000000,
			"status": "Active",
			"longitude": -104.761925,
			"latitude": 39.762714,
			"products": {
				"mcr": false,
				"megaport": [
					1,
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		},
		{
			"id": 112,
			"name": "Test Data Center 2",
			"campus": "campus_deprecated",
			"metro": "Ashburn",
			"country": "USA",
			"siteCode": "ashburnTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street Ashburn",
				"suburb": "Test Suburb Ashburn",
				"city": "Ashburn",
				"state": "VA",
				"country": "USA",
				"postcode": "20146"
			},
			"dc": {
				"id": 112,
				"name": "Test Data Center 2"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1483711200000,
			"status": "Active",
			"longitude": -77.487442,
			"latitude": 39.043757,
			"products": {
				"mcr": false,
				"megaport": [
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		},
		{
			"id": 113,
			"name": "New York Data Center",
			"campus": "campus_deprecated",
			"metro": "New York",
			"country": "USA",
			"siteCode": "nycTest",
			"networkRegion": "MP1",
			"address": {
				"street": "Test Street New York",
				"suburb": "Test Suburb New York",
				"city": "New York",
				"state": "NY",
				"country": "USA",
				"postcode": "10016"
			},
			"dc": {
				"id": 113,
				"name": "New York Data Center"
			},
			"market": "US",
			"vRouterAvailable": false,
			"liveDate": 1483711200000,
			"status": "Active",
			"longitude": -73.971321,
			"latitude": 40.776676,
			"products": {
				"mcr": false,
				"megaport": [
					10
				]
			},
			"ordering_message": null,
			"diversityZones": {}
		}
	]
	}`
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.GetLocationByNameFuzzy(ctx, "Test")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestListCountries(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()
	locSvc := client.LocationService
	want := []*types.Country{
		{
			Code:      "AUS",
			Name:      "Australia",
			Prefix:    "AU",
			SiteCount: 54,
		},
		{
			Code:      "GBR",
			Name:      "United Kingdom",
			Prefix:    "GB",
			SiteCount: 21,
		},
		{
			Code:      "USA",
			Name:      "USA",
			Prefix:    "US",
			SiteCount: 191,
		},
	}
	jblob := `
	{
	"message": "List all public network regions",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": [
		{
			"networkRegion": "MP1",
			"countries": [
				{
					"siteCount": 54,
					"code": "AUS",
					"prefix": "AU",
					"name": "Australia"
				},
				{
					"siteCount": 21,
					"code": "GBR",
					"prefix": "GB",
					"name": "United Kingdom"
				},
				{
					"siteCount": 191,
					"code": "USA",
					"prefix": "US",
					"name": "USA"
				}
			]
		}
	]
	}`
	path := "/v2/networkRegions"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.ListCountries(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestListMarketCodes(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()
	locSvc := client.LocationService
	want := []string{"AU", "GB", "US"}
	jblob := `
	{
	"message": "List all public network regions",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": [
		{
			"networkRegion": "MP1",
			"countries": [
				{
					"siteCount": 54,
					"code": "AUS",
					"prefix": "AU",
					"name": "Australia"
				},
				{
					"siteCount": 21,
					"code": "GBR",
					"prefix": "GB",
					"name": "United Kingdom"
				},
				{
					"siteCount": 191,
					"code": "USA",
					"prefix": "US",
					"name": "USA"
				}
			]
		}
	]
	}`
	path := "/v2/networkRegions"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.ListMarketCodes(ctx)
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestIsValidMarketCode(t *testing.T) {
	setup()
	defer teardown()
	ctx := context.Background()
	locSvc := client.LocationService
	want1 := true
	want2 := false
	jblob := `
	{
	"message": "List all public network regions",
	"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
	"data": [
		{
			"networkRegion": "MP1",
			"countries": [
				{
					"siteCount": 54,
					"code": "AUS",
					"prefix": "AU",
					"name": "Australia"
				},
				{
					"siteCount": 21,
					"code": "GBR",
					"prefix": "GB",
					"name": "United Kingdom"
				},
				{
					"siteCount": 191,
					"code": "USA",
					"prefix": "US",
					"name": "USA"
				}
			]
		}
	]
	}`
	path := "/v2/networkRegions"
	mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got1, err := locSvc.IsValidMarketCode(ctx, "US")
	require.NoError(t, err)
	require.Equal(t, &want1, got1)
	got2, err := locSvc.IsValidMarketCode(ctx, "BADCODE")
	require.NoError(t, err)
	require.Equal(t, &want2, got2)
}
