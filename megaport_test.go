package megaport

import (
	"context"
	"flag"
	"log/slog"
	"math/rand"
	"regexp"
	"time"
)

// runIntegrationTests is a flag to determine if integration tests should be run
var runIntegrationTests = flag.Bool("integration", false, "perform integration tests")

// programLevel is the log level for the test suite
var programLevel = new(slog.LevelVar)

// testResourceTags is a list of resource tags for integration tests
var testResourceTags = map[string]string{
	"key1": "value1", "key2": "value2", "key3": "value3",
}

var testProductResourceTags = []ResourceTag{
	{Key: "key1", Value: "value1"},
	{Key: "key2", Value: "value2"},
	{Key: "key3", Value: "value3"},
}

var testUpdatedResourceTags = map[string]string{
	"key1updated": "value1updated", "key2updated": "value2updated", "key3updated": "value3updated", "key4updated": "value4updated",
}

var resourceTagJSONBlob = `{
    "message": "test-message",
    "terms": "test-terms",
    "data": {
        "resourceTags": [
            {"key": "key1", "value": "value1"},
            {"key": "key2", "value": "value2"},
            {"key": "key3", "value": "value3"}
        ]
    }
}`

// Default Base URL for Integration Tests
const (
	MEGAPORTURL = "https://api-staging.megaport.com/"
)

// GetTime converts a timestamp to a time.Time object.
func GetTime(timestamp int64) time.Time {
	return time.Unix(timestamp/1000, 0)
}

// GenerateRandomVLAN generates a random VLAN ID.
func GenerateRandomVLAN() int {
	// exclude reserved values 0 and 4095 as per 802.1q
	return GenerateRandomNumber(1, 4094)
}

// GenerateRandomNumber generates a random number between an upper and lower bound.
func GenerateRandomNumber(lowerBound int, upperBound int) int {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	return random.Intn(upperBound) + lowerBound
}

// IsGuid checks if a string is a valid GUID.
func IsGuid(guids ...string) bool {
	guidRegex := regexp.MustCompile(`(?mi)^[0-9a-f]{8}-[0-9a-f]{4}-[0-5][0-9a-f]{3}-[089ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	for _, s := range guids {
		if guidRegex.FindIndex([]byte(s)) == nil {
			return false
		}
	}

	return true
}

func GetRandomLocation(ctx context.Context, svc LocationService, marketCode string) (*Location, error) {
	testLocations, err := svc.ListLocations(ctx)
	if err != nil {
		return nil, err
	}
	filtered, err := svc.FilterLocationsByMarketCode(ctx, marketCode, testLocations)
	if err != nil {
		return nil, err
	}
	filteredByMCR := svc.FilterLocationsByMcrAvailability(ctx, true, filtered)
	testLocation := filteredByMCR[GenerateRandomNumber(0, len(filteredByMCR)-1)]
	return testLocation, nil
}
