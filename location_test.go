package megaport

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

// LocationV3ClientTestSuite tests the Location Service V3 endpoints.
type LocationV3ClientTestSuite struct {
	ClientTestSuite
}

func TestLocationV3ClientTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(LocationV3ClientTestSuite))
}

func (suite *LocationV3ClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.client = NewClient(nil, nil)
	url, _ := url.Parse(suite.server.URL)
	suite.client.BaseURL = url
}

func (suite *LocationV3ClientTestSuite) TearDownTest() {
	suite.server.Close()
}

// TestListLocationsV3 tests the ListLocationsV3 method
func (suite *LocationV3ClientTestSuite) TestListLocationsV3() {
	ctx := context.Background()
	locSvc := suite.client.LocationService
	want := []*LocationV3{
		{
			ID:     2,
			Name:   "Equinix SY1",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			Address: LocationV3Address{
				Street:   "639 Gardeners Road",
				Suburb:   "Mascot",
				City:     "Sydney",
				State:    "NSW",
				Postcode: "2020",
				Country:  "Australia",
			},
			Latitude:  -33.921867,
			Longitude: 151.18802,
			DataCentre: LocationV3DataCentre{
				ID:   5,
				Name: "Equinix",
			},
			DiversityZones: &LocationV3DiversityZones{
				Red: &LocationV3DiversityZone{
					McrSpeedMbps:       []int{1000, 2500, 5000, 10000},
					MegaportSpeedMbps:  []int{1000, 10000, 100000},
					MveMaxCpuCoreCount: nil,
					MveAvailable:       true,
				},
				Blue: &LocationV3DiversityZone{
					McrSpeedMbps:       []int{1000, 2500, 5000, 10000, 25000, 50000, 100000},
					MegaportSpeedMbps:  []int{1000, 10000, 100000},
					MveMaxCpuCoreCount: nil,
					MveAvailable:       true,
				},
			},
			ProductAddOns: &LocationV3ProductAddOns{
				CrossConnect: &LocationV3CrossConnect{
					Available: true,
					Type:      stringPtr("STANDARD"),
				},
			},
			OrderingMessage: nil,
		},
		{
			ID:     3,
			Name:   "Global Switch Sydney West",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			Address: LocationV3Address{
				Street:   "400 Harris Street",
				Suburb:   "Ultimo",
				City:     "Sydney",
				State:    "NSW",
				Postcode: "2007",
				Country:  "Australia",
			},
			Latitude:  -33.87555,
			Longitude: 151.19783,
			DataCentre: LocationV3DataCentre{
				ID:   6,
				Name: "Global Switch",
			},
			DiversityZones: &LocationV3DiversityZones{
				Red: &LocationV3DiversityZone{
					McrSpeedMbps:       []int{1000, 2500, 5000, 10000},
					MegaportSpeedMbps:  []int{1000, 10000, 100000},
					MveMaxCpuCoreCount: nil,
					MveAvailable:       false,
				},
				Blue: &LocationV3DiversityZone{
					McrSpeedMbps:       []int{1000, 2500, 5000, 10000},
					MegaportSpeedMbps:  []int{1000, 10000, 100000},
					MveMaxCpuCoreCount: nil,
					MveAvailable:       true,
				},
			},
			ProductAddOns: &LocationV3ProductAddOns{
				CrossConnect: &LocationV3CrossConnect{
					Available: false,
					Type:      nil,
				},
			},
			OrderingMessage: nil,
		},
	}
	path := "/v3/locations"
	jblob := `{
    "message": "List public locations",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [
        {
            "id": 2,
            "name": "Equinix SY1",
            "address": {
                "street": "639 Gardeners Road",
                "suburb": "Mascot",
                "city": "Sydney",
                "state": "NSW",
                "postcode": "2020",
                "country": "Australia"
            },
            "dataCentre": {
                "id": 5,
                "name": "Equinix"
            },
            "metro": "Sydney",
            "market": "AU",
            "status": "Active",
            "longitude": 151.18802,
            "latitude": -33.921867,
            "orderingMessage": null,
            "diversityZones": {
                "red": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                },
                "blue": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000, 25000, 50000, 100000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                }
            },
            "productAddOns": {
                "crossConnect": {
                    "available": true,
                    "type": "STANDARD"
                }
            }
        },
        {
            "id": 3,
            "name": "Global Switch Sydney West",
            "address": {
                "street": "400 Harris Street",
                "suburb": "Ultimo",
                "city": "Sydney",
                "state": "NSW",
                "postcode": "2007",
                "country": "Australia"
            },
            "dataCentre": {
                "id": 6,
                "name": "Global Switch"
            },
            "metro": "Sydney",
            "market": "AU",
            "status": "Active",
            "longitude": 151.19783,
            "latitude": -33.87555,
            "orderingMessage": null,
            "diversityZones": {
                "red": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": false
                },
                "blue": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                }
            },
            "productAddOns": {
                "crossConnect": {
                    "available": false,
                    "type": null
                }
            }
        }
    ]
}`
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.ListLocationsV3(ctx)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetLocationByIDV3 tests the GetLocationByIDV3 method
func (suite *LocationV3ClientTestSuite) TestGetLocationByIDV3() {
	ctx := context.Background()
	locSvc := suite.client.LocationService
	want := &LocationV3{
		ID:     2,
		Name:   "Equinix SY1",
		Metro:  "Sydney",
		Market: "AU",
		Status: "Active",
		Address: LocationV3Address{
			Street:   "639 Gardeners Road",
			Suburb:   "Mascot",
			City:     "Sydney",
			State:    "NSW",
			Postcode: "2020",
			Country:  "Australia",
		},
		Latitude:  -33.921867,
		Longitude: 151.18802,
		DataCentre: LocationV3DataCentre{
			ID:   5,
			Name: "Equinix",
		},
		DiversityZones: &LocationV3DiversityZones{
			Red: &LocationV3DiversityZone{
				McrSpeedMbps:       []int{1000, 2500, 5000, 10000},
				MegaportSpeedMbps:  []int{1000, 10000, 100000},
				MveMaxCpuCoreCount: nil,
				MveAvailable:       true,
			},
			Blue: &LocationV3DiversityZone{
				McrSpeedMbps:       []int{1000, 2500, 5000, 10000, 25000, 50000, 100000},
				MegaportSpeedMbps:  []int{1000, 10000, 100000},
				MveMaxCpuCoreCount: nil,
				MveAvailable:       true,
			},
		},
		ProductAddOns: &LocationV3ProductAddOns{
			CrossConnect: &LocationV3CrossConnect{
				Available: true,
				Type:      stringPtr("STANDARD"),
			},
		},
		OrderingMessage: nil,
	}
	path := "/v3/locations"
	jblob := `{
    "message": "List public locations",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [
        {
            "id": 2,
            "name": "Equinix SY1",
            "address": {
                "street": "639 Gardeners Road",
                "suburb": "Mascot",
                "city": "Sydney",
                "state": "NSW",
                "postcode": "2020",
                "country": "Australia"
            },
            "dataCentre": {
                "id": 5,
                "name": "Equinix"
            },
            "metro": "Sydney",
            "market": "AU",
            "status": "Active",
            "longitude": 151.18802,
            "latitude": -33.921867,
            "orderingMessage": null,
            "diversityZones": {
                "red": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                },
                "blue": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000, 25000, 50000, 100000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                }
            },
            "productAddOns": {
                "crossConnect": {
                    "available": true,
                    "type": "STANDARD"
                }
            }
        },
        {
            "id": 3,
            "name": "Global Switch Sydney West",
            "address": {
                "street": "400 Harris Street",
                "suburb": "Ultimo",
                "city": "Sydney",
                "state": "NSW",
                "postcode": "2007",
                "country": "Australia"
            },
            "dataCentre": {
                "id": 6,
                "name": "Global Switch"
            },
            "metro": "Sydney",
            "market": "AU",
            "status": "Active",
            "longitude": 151.19783,
            "latitude": -33.87555,
            "orderingMessage": null,
            "diversityZones": {
                "red": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": false
                },
                "blue": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                }
            },
            "productAddOns": {
                "crossConnect": {
                    "available": false,
                    "type": null
                }
            }
        }
    ]
}`
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.GetLocationByIDV3(ctx, 2)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetLocationByNameV3 tests the GetLocationByNameV3 method.
func (suite *LocationV3ClientTestSuite) TestGetLocationByNameV3() {
	ctx := context.Background()
	locSvc := suite.client.LocationService
	want := &LocationV3{
		ID:     3,
		Name:   "Global Switch Sydney West",
		Metro:  "Sydney",
		Market: "AU",
		Status: "Active",
		Address: LocationV3Address{
			Street:   "400 Harris Street",
			Suburb:   "Ultimo",
			City:     "Sydney",
			State:    "NSW",
			Postcode: "2007",
			Country:  "Australia",
		},
		Latitude:  -33.87555,
		Longitude: 151.19783,
		DataCentre: LocationV3DataCentre{
			ID:   6,
			Name: "Global Switch",
		},
		DiversityZones: &LocationV3DiversityZones{
			Red: &LocationV3DiversityZone{
				McrSpeedMbps:       []int{1000, 2500, 5000, 10000},
				MegaportSpeedMbps:  []int{1000, 10000, 100000},
				MveMaxCpuCoreCount: nil,
				MveAvailable:       false,
			},
			Blue: &LocationV3DiversityZone{
				McrSpeedMbps:       []int{1000, 2500, 5000, 10000},
				MegaportSpeedMbps:  []int{1000, 10000, 100000},
				MveMaxCpuCoreCount: nil,
				MveAvailable:       true,
			},
		},
		ProductAddOns: &LocationV3ProductAddOns{
			CrossConnect: &LocationV3CrossConnect{
				Available: false,
				Type:      nil,
			},
		},
		OrderingMessage: nil,
	}
	path := "/v3/locations"
	jblob := `{
    "message": "List public locations",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [
        {
            "id": 2,
            "name": "Equinix SY1",
            "address": {
                "street": "639 Gardeners Road",
                "suburb": "Mascot",
                "city": "Sydney",
                "state": "NSW",
                "postcode": "2020",
                "country": "Australia"
            },
            "dataCentre": {
                "id": 5,
                "name": "Equinix"
            },
            "metro": "Sydney",
            "market": "AU",
            "status": "Active",
            "longitude": 151.18802,
            "latitude": -33.921867,
            "orderingMessage": null,
            "diversityZones": {
                "red": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                },
                "blue": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000, 25000, 50000, 100000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                }
            },
            "productAddOns": {
                "crossConnect": {
                    "available": true,
                    "type": "STANDARD"
                }
            }
        },
        {
            "id": 3,
            "name": "Global Switch Sydney West",
            "address": {
                "street": "400 Harris Street",
                "suburb": "Ultimo",
                "city": "Sydney",
                "state": "NSW",
                "postcode": "2007",
                "country": "Australia"
            },
            "dataCentre": {
                "id": 6,
                "name": "Global Switch"
            },
            "metro": "Sydney",
            "market": "AU",
            "status": "Active",
            "longitude": 151.19783,
            "latitude": -33.87555,
            "orderingMessage": null,
            "diversityZones": {
                "red": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": false
                },
                "blue": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                }
            },
            "productAddOns": {
                "crossConnect": {
                    "available": false,
                    "type": null
                }
            }
        }
    ]
}`
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.GetLocationByNameV3(ctx, "Global Switch Sydney West")
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestGetLocationByNameFuzzyV3 tests the GetLocationByNameFuzzyV3 method.
func (suite *LocationV3ClientTestSuite) TestGetLocationByNameFuzzyV3() {
	ctx := context.Background()
	locSvc := suite.client.LocationService
	want := []*LocationV3{
		{
			ID:     3,
			Name:   "Global Switch Sydney West",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			Address: LocationV3Address{
				Street:   "400 Harris Street",
				Suburb:   "Ultimo",
				City:     "Sydney",
				State:    "NSW",
				Postcode: "2007",
				Country:  "Australia",
			},
			Latitude:  -33.87555,
			Longitude: 151.19783,
			DataCentre: LocationV3DataCentre{
				ID:   6,
				Name: "Global Switch",
			},
			DiversityZones: &LocationV3DiversityZones{
				Red: &LocationV3DiversityZone{
					McrSpeedMbps:       []int{1000, 2500, 5000, 10000},
					MegaportSpeedMbps:  []int{1000, 10000, 100000},
					MveMaxCpuCoreCount: nil,
					MveAvailable:       false,
				},
				Blue: &LocationV3DiversityZone{
					McrSpeedMbps:       []int{1000, 2500, 5000, 10000},
					MegaportSpeedMbps:  []int{1000, 10000, 100000},
					MveMaxCpuCoreCount: nil,
					MveAvailable:       true,
				},
			},
			ProductAddOns: &LocationV3ProductAddOns{
				CrossConnect: &LocationV3CrossConnect{
					Available: false,
					Type:      nil,
				},
			},
			OrderingMessage: nil,
		},
	}
	path := "/v3/locations"
	jblob := `{
    "message": "List public locations",
    "terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
    "data": [
        {
            "id": 2,
            "name": "Equinix SY1",
            "address": {
                "street": "639 Gardeners Road",
                "suburb": "Mascot",
                "city": "Sydney",
                "state": "NSW",
                "postcode": "2020",
                "country": "Australia"
            },
            "dataCentre": {
                "id": 5,
                "name": "Equinix"
            },
            "metro": "Sydney",
            "market": "AU",
            "status": "Active",
            "longitude": 151.18802,
            "latitude": -33.921867,
            "orderingMessage": null,
            "diversityZones": {
                "red": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                },
                "blue": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000, 25000, 50000, 100000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                }
            },
            "productAddOns": {
                "crossConnect": {
                    "available": true,
                    "type": "STANDARD"
                }
            }
        },
        {
            "id": 3,
            "name": "Global Switch Sydney West",
            "address": {
                "street": "400 Harris Street",
                "suburb": "Ultimo",
                "city": "Sydney",
                "state": "NSW",
                "postcode": "2007",
                "country": "Australia"
            },
            "dataCentre": {
                "id": 6,
                "name": "Global Switch"
            },
            "metro": "Sydney",
            "market": "AU",
            "status": "Active",
            "longitude": 151.19783,
            "latitude": -33.87555,
            "orderingMessage": null,
            "diversityZones": {
                "red": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": false
                },
                "blue": {
                    "mcrSpeedMbps": [1000, 2500, 5000, 10000],
                    "megaportSpeedMbps": [1000, 10000, 100000],
                    "mveMaxCpuCoreCount": null,
                    "mveAvailable": true
                }
            },
            "productAddOns": {
                "crossConnect": {
                    "available": false,
                    "type": null
                }
            }
        }
    ]
}`
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	got, err := locSvc.GetLocationByNameFuzzyV3(ctx, "Global Switch")
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestFilterLocationsByMarketCodeV3 tests the FilterLocationsByMarketCodeV3 method.
func (suite *LocationV3ClientTestSuite) TestFilterLocationsByMarketCodeV3() {
	ctx := context.Background()
	locSvc := suite.client.LocationService

	// Mock the /v2/networkRegions endpoint that IsValidMarketCode depends on
	path := "/v2/networkRegions"
	jblob := `{
		"message": "Network Regions",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"networkRegion": "MP1",
				"countries": [
					{
						"country": "Australia",
						"countryPrefix": "AU",
						"prefix": "AU"
					},
					{
						"country": "United States",
						"countryPrefix": "US", 
						"prefix": "US"
					}
				]
			}
		]
	}`
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})

	locations := []*LocationV3{
		{
			ID:     2,
			Name:   "Equinix SY1",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
		},
		{
			ID:     3,
			Name:   "Global Switch Sydney West",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
		},
		{
			ID:     100,
			Name:   "Test US Location",
			Metro:  "Denver",
			Market: "US",
			Status: "Active",
		},
	}

	want := []*LocationV3{
		{
			ID:     2,
			Name:   "Equinix SY1",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
		},
		{
			ID:     3,
			Name:   "Global Switch Sydney West",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
		},
	}

	got, err := locSvc.FilterLocationsByMarketCodeV3(ctx, "AU", locations)
	suite.NoError(err)
	suite.Equal(want, got)
}

// TestFilterLocationsByMcrAvailabilityV3 tests the FilterLocationsByMcrAvailabilityV3 method.
func (suite *LocationV3ClientTestSuite) TestFilterLocationsByMcrAvailabilityV3() {
	ctx := context.Background()
	locSvc := suite.client.LocationService

	locations := []*LocationV3{
		{
			ID:     2,
			Name:   "Equinix SY1",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			DiversityZones: &LocationV3DiversityZones{
				Red: &LocationV3DiversityZone{
					McrSpeedMbps: []int{1000, 2500, 5000, 10000},
				},
			},
		},
		{
			ID:     3,
			Name:   "Global Switch Sydney West",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			DiversityZones: &LocationV3DiversityZones{
				Red: &LocationV3DiversityZone{
					McrSpeedMbps: []int{},
				},
			},
		},
		{
			ID:             4,
			Name:           "Location with no MCR",
			Metro:          "Melbourne",
			Market:         "AU",
			Status:         "Active",
			DiversityZones: &LocationV3DiversityZones{},
		},
	}

	want := []*LocationV3{
		{
			ID:     2,
			Name:   "Equinix SY1",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
			DiversityZones: &LocationV3DiversityZones{
				Red: &LocationV3DiversityZone{
					McrSpeedMbps: []int{1000, 2500, 5000, 10000},
				},
			},
		},
	}

	got := locSvc.FilterLocationsByMcrAvailabilityV3(ctx, true, locations)
	suite.Equal(want, got)
}

// TestFilterLocationsByMetroV3 tests the FilterLocationsByMetroV3 method.
func (suite *LocationV3ClientTestSuite) TestFilterLocationsByMetroV3() {
	ctx := context.Background()
	locSvc := suite.client.LocationService

	locations := []*LocationV3{
		{
			ID:     2,
			Name:   "Equinix SY1",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
		},
		{
			ID:     3,
			Name:   "Global Switch Sydney West",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
		},
		{
			ID:     100,
			Name:   "Test US Location",
			Metro:  "Denver",
			Market: "US",
			Status: "Active",
		},
	}

	want := []*LocationV3{
		{
			ID:     2,
			Name:   "Equinix SY1",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
		},
		{
			ID:     3,
			Name:   "Global Switch Sydney West",
			Metro:  "Sydney",
			Market: "AU",
			Status: "Active",
		},
	}

	got := locSvc.FilterLocationsByMetroV3(ctx, "Sydney", locations)
	suite.Equal(want, got)

	// Test with no matching metro
	gotEmpty := locSvc.FilterLocationsByMetroV3(ctx, "Auckland", locations)
	suite.Empty(gotEmpty)
}

// TestLocationV3HelperMethods tests the helper methods for LocationV3 struct.
func (suite *LocationV3ClientTestSuite) TestLocationV3HelperMethods() {
	// Test location with MCR support
	locationWithMCR := &LocationV3{
		ID:   2,
		Name: "Equinix SY1",
		DiversityZones: &LocationV3DiversityZones{
			Red: &LocationV3DiversityZone{
				McrSpeedMbps:      []int{1000, 2500, 5000, 10000},
				MegaportSpeedMbps: []int{1000, 10000, 100000},
				MveAvailable:      true,
			},
			Blue: &LocationV3DiversityZone{
				McrSpeedMbps:       []int{1000, 2500, 5000, 10000, 25000, 50000, 100000},
				MegaportSpeedMbps:  []int{1000, 10000, 100000},
				MveMaxCpuCoreCount: intPtr(16),
				MveAvailable:       true,
			},
		},
		ProductAddOns: &LocationV3ProductAddOns{
			CrossConnect: &LocationV3CrossConnect{
				Available: true,
				Type:      stringPtr("STANDARD"),
			},
		},
		DataCentre: LocationV3DataCentre{
			ID:   5,
			Name: "Equinix",
		},
		Address: LocationV3Address{
			Country: "Australia",
		},
	}

	// Test HasMCRSupport
	suite.True(locationWithMCR.HasMCRSupport())

	// Test GetMCRSpeeds
	expectedMCRSpeeds := []int{1000, 2500, 5000, 10000, 25000, 50000, 100000}
	suite.Equal(expectedMCRSpeeds, locationWithMCR.GetMCRSpeeds())

	// Test GetMegaportSpeeds
	expectedMegaportSpeeds := []int{1000, 10000, 100000}
	suite.Equal(expectedMegaportSpeeds, locationWithMCR.GetMegaportSpeeds())

	// Test HasMVESupport
	suite.True(locationWithMCR.HasMVESupport())

	// Test GetMVEMaxCpuCores
	suite.Equal(intPtr(16), locationWithMCR.GetMVEMaxCpuCores())

	// Test HasCrossConnectSupport
	suite.True(locationWithMCR.HasCrossConnectSupport())

	// Test GetCrossConnectType
	suite.Equal("STANDARD", locationWithMCR.GetCrossConnectType())

	// Test GetDataCenterName
	suite.Equal("Equinix", locationWithMCR.GetDataCenterName())

	// Test GetDataCenterID
	suite.Equal(5, locationWithMCR.GetDataCenterID())

	// Test GetCountry
	suite.Equal("Australia", locationWithMCR.GetCountry())

	// Test location without MCR support
	locationWithoutMCR := &LocationV3{
		ID:   4,
		Name: "Location without MCR",
		DiversityZones: &LocationV3DiversityZones{
			Red: &LocationV3DiversityZone{
				McrSpeedMbps: []int{},
				MveAvailable: false,
			},
		},
		ProductAddOns: &LocationV3ProductAddOns{
			CrossConnect: &LocationV3CrossConnect{
				Available: false,
				Type:      nil,
			},
		},
	}

	// Test HasMCRSupport returns false
	suite.False(locationWithoutMCR.HasMCRSupport())

	// Test HasMVESupport returns false
	suite.False(locationWithoutMCR.HasMVESupport())

	// Test HasCrossConnectSupport returns false
	suite.False(locationWithoutMCR.HasCrossConnectSupport())

	// Test GetCrossConnectType returns empty string
	suite.Equal("", locationWithoutMCR.GetCrossConnectType())
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

// TestGetRandom tests the GetRandom method.
func (suite *LocationV3ClientTestSuite) TestGetRandom() {
	ctx := context.Background()
	locSvc := suite.client.LocationService
	want := []*LocationV3{
		{
			ID:     111,
			Name:   "Test Data Center",
			Metro:  "Denver",
			Market: "US",
			Status: "Active",
			Address: LocationV3Address{
				Street:   "Test Street Denver",
				Suburb:   "Test Suburb Denver",
				City:     "Denver",
				State:    "CO",
				Postcode: "80011",
				Country:  "USA",
			},
			Latitude:  39.762714,
			Longitude: -104.761925,
			DataCentre: LocationV3DataCentre{
				ID:   111,
				Name: "Test Data Center",
			},
			DiversityZones: &LocationV3DiversityZones{
				Red: &LocationV3DiversityZone{
					McrSpeedMbps:      []int{1000, 2500, 5000, 10000},
					MegaportSpeedMbps: []int{1000, 10000},
					MveAvailable:      true,
				},
			},
			ProductAddOns: &LocationV3ProductAddOns{
				CrossConnect: &LocationV3CrossConnect{
					Available: true,
					Type:      stringPtr("STANDARD"),
				},
			},
			OrderingMessage: nil,
		},
		{
			ID:     112,
			Name:   "Test Data Center 2",
			Metro:  "Ashburn",
			Market: "US",
			Status: "Active",
			Address: LocationV3Address{
				Street:   "Test Street Ashburn",
				Suburb:   "Test Suburb Ashburn",
				City:     "Ashburn",
				State:    "VA",
				Postcode: "20146",
				Country:  "USA",
			},
			Latitude:  39.043757,
			Longitude: -77.487442,
			DataCentre: LocationV3DataCentre{
				ID:   112,
				Name: "Test Data Center 2",
			},
			DiversityZones: &LocationV3DiversityZones{
				Red: &LocationV3DiversityZone{
					McrSpeedMbps:      []int{1000, 2500, 5000, 10000},
					MegaportSpeedMbps: []int{1000, 10000},
					MveAvailable:      true,
				},
			},
			ProductAddOns: &LocationV3ProductAddOns{
				CrossConnect: &LocationV3CrossConnect{
					Available: true,
					Type:      stringPtr("STANDARD"),
				},
			},
			OrderingMessage: nil,
		},
	}
	path := "/v3/locations"
	jblob := `{
		"message": "List public locations",
		"terms": "This data is subject to the Acceptable Use Policy https://www.megaport.com/legal/acceptable-use-policy",
		"data": [
			{
				"id": 111,
				"name": "Test Data Center",
				"address": {
					"street": "Test Street Denver",
					"suburb": "Test Suburb Denver",
					"city": "Denver",
					"state": "CO",
					"postcode": "80011",
					"country": "USA"
				},
				"dataCentre": {
					"id": 111,
					"name": "Test Data Center"
				},
				"metro": "Denver",
				"market": "US",
				"status": "Active",
				"longitude": -104.761925,
				"latitude": 39.762714,
				"orderingMessage": null,
				"diversityZones": {
					"red": {
						"mcrSpeedMbps": [1000, 2500, 5000, 10000],
						"megaportSpeedMbps": [1000, 10000],
						"mveMaxCpuCoreCount": null,
						"mveAvailable": true
					}
				},
				"productAddOns": {
					"crossConnect": {
						"available": true,
						"type": "STANDARD"
					}
				}
			},
			{
				"id": 112,
				"name": "Test Data Center 2",
				"address": {
					"street": "Test Street Ashburn",
					"suburb": "Test Suburb Ashburn",
					"city": "Ashburn",
					"state": "VA",
					"postcode": "20146",
					"country": "USA"
				},
				"dataCentre": {
					"id": 112,
					"name": "Test Data Center 2"
				},
				"metro": "Ashburn",
				"market": "US",
				"status": "Active",
				"longitude": -77.487442,
				"latitude": 39.043757,
				"orderingMessage": null,
				"diversityZones": {
					"red": {
						"mcrSpeedMbps": [1000, 2500, 5000, 10000],
						"megaportSpeedMbps": [1000, 10000],
						"mveMaxCpuCoreCount": null,
						"mveAvailable": true
					}
				},
				"productAddOns": {
					"crossConnect": {
						"available": true,
						"type": "STANDARD"
					}
				}
			},
			{
				"id": 113,
				"name": "New York Data Center",
				"address": {
					"street": "Test Street New York",
					"suburb": "Test Suburb New York",
					"city": "New York",
					"state": "NY",
					"postcode": "10016",
					"country": "USA"
				},
				"dataCentre": {
					"id": 113,
					"name": "New York Data Center"
				},
				"metro": "New York",
				"market": "US",
				"status": "Active",
				"longitude": -73.971321,
				"latitude": 40.776676,
				"orderingMessage": null,
				"diversityZones": {
					"red": {
						"mcrSpeedMbps": [],
						"megaportSpeedMbps": [10000],
						"mveMaxCpuCoreCount": null,
						"mveAvailable": false
					}
				},
				"productAddOns": {
					"crossConnect": {
						"available": false,
						"type": null
					}
				}
			},
			{
				"id": 114,
				"name": "London Data Center",
				"address": {
					"street": "Test Street London",
					"suburb": "",
					"city": "London",
					"state": "",
					"postcode": "SL1 4AX",
					"country": "United Kingdom"
				},
				"dataCentre": {
					"id": 114,
					"name": "London Data Center"
				},
				"metro": "London",
				"market": "UK",
				"status": "Active",
				"longitude": -0.628975,
				"latitude": 51.522484,
				"orderingMessage": null,
				"diversityZones": {
					"red": {
						"mcrSpeedMbps": [],
						"megaportSpeedMbps": [10000],
						"mveMaxCpuCoreCount": null,
						"mveAvailable": false
					}
				},
				"productAddOns": {
					"crossConnect": {
						"available": false,
						"type": null
					}
				}
			}
		]
	}`
	jblob2 := `{
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
	path2 := "/v2/networkRegions"
	suite.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob)
	})
	suite.mux.HandleFunc(path2, func(w http.ResponseWriter, r *http.Request) {
		suite.testMethod(r, http.MethodGet)
		fmt.Fprint(w, jblob2)
	})
	got, err := GetRandomLocation(ctx, locSvc, "US")
	suite.NoError(err)
	suite.Contains(want, got)
}
