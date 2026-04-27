package megaport

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"regexp"
	"strings"
	"testing"
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

func GetRandomLocation(ctx context.Context, svc LocationService, marketCode string) (*LocationV3, error) {
	testLocations, err := svc.ListLocationsV3(ctx)
	if err != nil {
		return nil, err
	}
	filtered, err := svc.FilterLocationsByMarketCodeV3(ctx, marketCode, testLocations)
	if err != nil {
		return nil, err
	}
	filteredByMCR := svc.FilterLocationsByMcrAvailabilityV3(ctx, true, filtered)
	if len(filteredByMCR) == 0 {
		return nil, fmt.Errorf("no MCR-capable locations in market %s", marketCode)
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	testLocation := filteredByMCR[rng.Intn(len(filteredByMCR))]
	return testLocation, nil
}

// portLocationOpts refines findActivePortLocation selection.
type portLocationOpts struct {
	// Metro, if non-empty, restricts candidates to a specific metro
	// (e.g. "Sydney") — required by IX-style tests where the fabric name
	// is metro-scoped ("Sydney IX").
	Metro string
}

// findActivePortLocation returns a random ACTIVE location in the given market
// that advertises Megaport port capacity at the given speed in at least one
// diversity zone AND accepts a probe port order with those parameters. It also
// claims the location via claimPortLocation so parallel suites don't pick the
// same site; the claim is released via t.Cleanup when the caller's test
// method completes.
//
//nolint:unparam // marketCode is parameterised for future callers targeting non-AU markets.
func findActivePortLocation(ctx context.Context, t *testing.T, c *Client, marketCode string, speedMbps int, opts ...portLocationOpts) (*LocationV3, error) {
	t.Helper()
	var opt portLocationOpts
	if len(opts) > 0 {
		opt = opts[0]
	}
	locations, err := c.LocationService.ListLocationsV3(ctx)
	if err != nil {
		return nil, err
	}
	filtered, err := c.LocationService.FilterLocationsByMarketCodeV3(ctx, marketCode, locations)
	if err != nil {
		return nil, err
	}
	shuffled := make([]*LocationV3, 0, len(filtered))
	for _, loc := range filtered {
		if !strings.EqualFold(loc.Status, "active") {
			continue
		}
		if !locationHasPortSpeed(loc, speedMbps) {
			continue
		}
		if opt.Metro != "" && !strings.EqualFold(loc.Metro, opt.Metro) {
			continue
		}
		shuffled = append(shuffled, loc)
	}
	//nolint:gosec // test-only shuffle; cryptographic randomness not required
	rPort := rand.New(rand.NewSource(time.Now().UnixNano()))
	rPort.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	if len(shuffled) == 0 {
		return nil, fmt.Errorf("no active %s locations advertise a %d Mbps port (metro=%q)", marketCode, speedMbps, opt.Metro)
	}

	for _, loc := range shuffled {
		err := c.PortService.ValidatePortOrder(ctx, &BuyPortRequest{
			LocationId: loc.ID,
			Name:       "probe",
			Term:       1,
			PortSpeed:  speedMbps,
			Market:     marketCode,
		})
		if err != nil {
			c.Logger.DebugContext(ctx, "port location probe failed, trying next",
				slog.Int("location_id", loc.ID),
				slog.String("location_name", loc.Name),
				slog.String("error", err.Error()))
			continue
		}
		if !claimPortLocation(t, loc.ID) {
			continue
		}
		return loc, nil
	}
	return nil, fmt.Errorf("no active %s location accepted a %d Mbps port validate probe (metro=%q)", marketCode, speedMbps, opt.Metro)
}

// findActiveMVELocation returns an ACTIVE location in the given market that
// accepts a probe MVE order with the given vendor config. Mirrors the
// terraform provider's findMVETestLocation — probes ValidateMVEOrder since
// MVE capacity is per-image and can't be derived from LocationV3 alone.
//
//nolint:unparam // marketCode is parameterised for future non-AU callers.
func findActiveMVELocation(ctx context.Context, t *testing.T, c *Client, marketCode string, vendorConfig VendorConfig, vnics []MVENetworkInterface, diversityZone string) (*LocationV3, error) {
	t.Helper()
	locations, err := c.LocationService.ListLocationsV3(ctx)
	if err != nil {
		return nil, err
	}
	filtered, err := c.LocationService.FilterLocationsByMarketCodeV3(ctx, marketCode, locations)
	if err != nil {
		return nil, err
	}
	shuffled := make([]*LocationV3, 0, len(filtered))
	for _, loc := range filtered {
		if !strings.EqualFold(loc.Status, "active") || !loc.HasMVESupport() {
			continue
		}
		shuffled = append(shuffled, loc)
	}
	//nolint:gosec // test-only shuffle; cryptographic randomness not required
	rMVE := rand.New(rand.NewSource(time.Now().UnixNano()))
	rMVE.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	for _, loc := range shuffled {
		err := c.MVEService.ValidateMVEOrder(ctx, &BuyMVERequest{
			LocationID:    loc.ID,
			Name:          "probe",
			Term:          1,
			VendorConfig:  vendorConfig,
			Vnics:         vnics,
			DiversityZone: diversityZone,
		})
		if err != nil {
			c.Logger.DebugContext(ctx, "mve location probe failed, trying next",
				slog.Int("location_id", loc.ID),
				slog.String("location_name", loc.Name),
				slog.String("error", err.Error()))
			continue
		}
		if !claimMVELocation(t, loc.ID) {
			continue
		}
		return loc, nil
	}
	return nil, fmt.Errorf("no active %s location accepted the MVE validate probe", marketCode)
}

// findActiveNATGatewayLocation returns an ACTIVE location in the given market
// that advertises NAT Gateway capacity at the given speed. Uses the
// DiversityZones.natGatewaySpeedMbps field surfaced by v3/locations.
//
// Unlike findActivePortLocation and findActiveMCRLocation, this helper does NOT
// issue a probe validate order. The NAT Gateway validate endpoint
// (POST /v3/networkdesign/validate) requires a productUID of an already-created
// gateway in DESIGN state; there is no pre-creation location-probe API
// equivalent to ValidatePortOrder / ValidateMCROrder. The v3/locations
// advertised-speed field is therefore the authoritative capacity signal here.
//
//nolint:unparam // marketCode is parameterised for future non-AU callers.
func findActiveNATGatewayLocation(ctx context.Context, t *testing.T, c *Client, marketCode string, speedMbps int) (*LocationV3, error) {
	t.Helper()
	locations, err := c.LocationService.ListLocationsV3(ctx)
	if err != nil {
		return nil, err
	}
	filtered, err := c.LocationService.FilterLocationsByMarketCodeV3(ctx, marketCode, locations)
	if err != nil {
		return nil, err
	}
	eligible := make([]*LocationV3, 0, len(filtered))
	for _, loc := range filtered {
		if !strings.EqualFold(loc.Status, "active") {
			continue
		}
		if loc.SupportsNATGatewaySpeed(speedMbps) {
			eligible = append(eligible, loc)
		}
	}
	if len(eligible) == 0 {
		return nil, fmt.Errorf("no active %s location advertises %d Mbps NAT Gateway capacity", marketCode, speedMbps)
	}
	//nolint:gosec // test-only shuffle; cryptographic randomness not required
	rNAT := rand.New(rand.NewSource(time.Now().UnixNano()))
	rNAT.Shuffle(len(eligible), func(i, j int) { eligible[i], eligible[j] = eligible[j], eligible[i] })
	for _, loc := range eligible {
		if claimNATGatewayLocation(t, loc.ID) {
			return loc, nil
		}
	}
	return nil, fmt.Errorf("no unclaimed %s location with %d Mbps NAT Gateway capacity", marketCode, speedMbps)
}

// findActiveMCRLocation returns an ACTIVE location in the given market that
// advertises MCR capacity at the given speed in the given diversity zone AND
// accepts a probe MCR order with those parameters. Pass an empty string for
// diversityZone to allow any zone. Probing via ValidateMCROrder catches sites
// where the speed is listed but the pool is exhausted ("No available capacity
// for this request"), which the advertised-speeds check alone can't detect.
// Mirrors the approach used by the terraform provider's findMVETestLocation.
//
//nolint:unparam // marketCode is parameterised for future callers targeting non-AU markets.
func findActiveMCRLocation(ctx context.Context, t *testing.T, c *Client, marketCode, diversityZone string, speedMbps int, addOns ...MCRAddOn) (*LocationV3, error) {
	t.Helper()
	locations, err := c.LocationService.ListLocationsV3(ctx)
	if err != nil {
		return nil, err
	}
	filtered, err := c.LocationService.FilterLocationsByMarketCodeV3(ctx, marketCode, locations)
	if err != nil {
		return nil, err
	}
	shuffled := make([]*LocationV3, 0, len(filtered))
	for _, loc := range filtered {
		if !strings.EqualFold(loc.Status, "active") {
			continue
		}
		if !zoneHasMCRSpeed(loc, diversityZone, speedMbps) {
			continue
		}
		shuffled = append(shuffled, loc)
	}
	//nolint:gosec // test-only shuffle; cryptographic randomness not required
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	const maxProbes = 5
	probeCount := 0
	for _, loc := range shuffled {
		if probeCount >= maxProbes {
			return nil, fmt.Errorf("exceeded %d probe attempts for MCR location in market %q", maxProbes, marketCode)
		}
		// Claim before probing so a parallel test that already holds this
		// location doesn't consume a probe attempt. Claim collisions are
		// free — only actual probe calls count against maxProbes.
		if !claimMCRLocation(t, loc.ID) {
			continue
		}
		if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < 15*time.Second {
			releaseMCRLocation(loc.ID)
			return nil, fmt.Errorf("context deadline too close to attempt MCR location probe")
		}
		probeCount++
		probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		probeAddOns := make([]MCRAddOn, 0, len(addOns))
		for _, a := range addOns {
			switch addon := a.(type) {
			case *MCRAddOnIPsecConfig:
				if addon == nil {
					cancel()
					return nil, fmt.Errorf("nil *MCRAddOnIPsecConfig in addOns")
				}
				cp := *addon
				probeAddOns = append(probeAddOns, &cp)
			default:
				cancel()
				return nil, fmt.Errorf("unsupported MCRAddOn type %T for probe cloning", a)
			}
		}
		err := c.MCRService.ValidateMCROrder(probeCtx, &BuyMCRRequest{
			LocationID:    loc.ID,
			Name:          "probe",
			Term:          1,
			PortSpeed:     speedMbps,
			MCRAsn:        0,
			DiversityZone: diversityZone,
			AddOns:        probeAddOns,
		})
		cancel()
		if err != nil {
			releaseMCRLocation(loc.ID)
			c.Logger.DebugContext(ctx, "mcr location probe failed, trying next",
				slog.Int("location_id", loc.ID),
				slog.String("location_name", loc.Name),
				slog.String("diversity_zone", diversityZone),
				slog.String("error", err.Error()))
			continue
		}
		return loc, nil
	}
	return nil, fmt.Errorf("no active %s location accepted a %d Mbps MCR validate probe (zone=%q)", marketCode, speedMbps, diversityZone)
}

// zoneHasMCRSpeed reports whether the named diversity zone at loc advertises
// the given MCR speed. An empty zone means "any zone".
func zoneHasMCRSpeed(loc *LocationV3, zone string, speedMbps int) bool {
	if loc == nil || loc.DiversityZones == nil {
		return false
	}
	check := func(z *LocationV3DiversityZone) bool {
		if z == nil {
			return false
		}
		for _, s := range z.McrSpeedMbps {
			if s == speedMbps {
				return true
			}
		}
		return false
	}
	switch strings.ToLower(zone) {
	case "red":
		return check(loc.DiversityZones.Red)
	case "blue":
		return check(loc.DiversityZones.Blue)
	case "":
		return check(loc.DiversityZones.Red) || check(loc.DiversityZones.Blue)
	default:
		return false
	}
}

func locationHasPortSpeed(loc *LocationV3, speedMbps int) bool {
	if loc == nil || loc.DiversityZones == nil {
		return false
	}
	check := func(zone *LocationV3DiversityZone) bool {
		if zone == nil {
			return false
		}
		for _, s := range zone.MegaportSpeedMbps {
			if s == speedMbps {
				return true
			}
		}
		return false
	}
	return check(loc.DiversityZones.Red) || check(loc.DiversityZones.Blue)
}
