package megaport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// cspCredentials mirrors the terraform-provider-megaport
// testdata/csp_credentials.json schema. Tests load it from either the
// CSP_CREDENTIALS_JSON env var (preferred in CI) or the on-disk file
// testdata/csp_credentials.json (developer convenience). Neither source is
// committed — the example file at testdata/csp_credentials.json.example
// documents the shape.
type cspCredentials struct {
	AzureServiceKeys  []string `json:"azure_service_keys"`
	GooglePairingKeys []string `json:"google_pairing_keys"`
}

// cspPickResult holds a validated CSP key along with the partner port UID,
// location, and market that the key resolves to. Tests should place their
// A-End port or MCR at LocationID and use Market for BuyPortRequest /
// BuyMCRRequest so the A-End is always in the same market as the partner port.
type cspPickResult struct {
	Key            string
	PartnerPortUID string
	LocationID     int
	Market         string
}

// pickAzureServiceKey returns the first Azure service key from the pool that
// has available VXC capacity. Calls t.Skip if no usable key is found.
func pickAzureServiceKey(t *testing.T, client *Client) cspPickResult {
	t.Helper()
	return pickCSPKey(t, client, PARTNER_AZURE, PARTNER_AZURE)
}

// pickGCPPairingKey returns the first GCP pairing key from the pool that has
// available VXC capacity. Calls t.Skip if no usable key is found.
func pickGCPPairingKey(t *testing.T, client *Client) cspPickResult {
	t.Helper()
	return pickCSPKey(t, client, PARTNER_GOOGLE, PARTNER_GOOGLE)
}

// cspClaimedKeys tracks CSP keys and partner port UIDs already handed out so
// parallel tests don't reuse the same key or hit the same Azure port
// (different keys can map to the same ExpressRoute circuit, causing VLAN
// conflicts).
var (
	cspClaimedMu    sync.Mutex
	cspClaimedKeys  = map[string]bool{}
	cspClaimedPorts = map[string]bool{}
)

// pickCSPKey is the shared implementation for CSP key pickers. It validates
// each key via LookupPartnerPorts, then resolves the partner port's location
// via ListPartnerMegaports so the caller can place its A-End at a compatible
// site. Each key is claimed exclusively so parallel tests get unique keys.
func pickCSPKey(t *testing.T, client *Client, partner, connectType string) cspPickResult {
	t.Helper()
	creds, err := loadCSPCredentials()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("skipping: %v", err)
			return cspPickResult{}
		}
		t.Fatalf("failed to load CSP credentials: %v", err)
		return cspPickResult{} // unreachable
	}

	var keys []string
	switch partner {
	case PARTNER_AZURE:
		keys = creds.AzureServiceKeys
	case PARTNER_GOOGLE:
		keys = creds.GooglePairingKeys
	}
	if len(keys) == 0 {
		t.Skipf("skipping: no %s keys in testdata/csp_credentials.json (or CSP_CREDENTIALS_JSON)", partner)
		return cspPickResult{}
	}

	ctx := context.Background()
	partnerPorts, err := client.PartnerService.ListPartnerMegaports(ctx)
	if err != nil {
		t.Skipf("skipping: could not list partner ports: %v", err)
		return cspPickResult{}
	}
	portLocation := make(map[string]int, len(partnerPorts))
	for _, pp := range partnerPorts {
		if strings.EqualFold(pp.ConnectType, connectType) {
			portLocation[pp.ProductUID] = pp.LocationId
		}
	}

	//nolint:gosec // weak random is fine for test key shuffling
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	shuffled := make([]string, len(keys))
	copy(shuffled, keys)
	r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

	mask := func(s string) string {
		if s == "" {
			return ""
		}
		return "[REDACTED]"
	}

	for _, key := range shuffled {
		masked := mask(key)
		cspClaimedMu.Lock()
		if cspClaimedKeys[key] {
			cspClaimedMu.Unlock()
			t.Logf("pick%sKey: key %s already claimed, skipping", partner, masked)
			continue
		}
		cspClaimedMu.Unlock()

		resp, lookupErr := client.VXCService.LookupPartnerPorts(ctx, &LookupPartnerPortsRequest{
			Partner:   partner,
			Key:       key,
			PortSpeed: 1000,
		})
		if lookupErr != nil {
			t.Logf("pick%sKey: key %s unavailable: %v", partner, masked, lookupErr)
			continue
		}
		locID := portLocation[resp.ProductUID]
		if locID == 0 {
			t.Logf("pick%sKey: key %s resolved but location unknown, skipping", partner, masked)
			continue
		}

		loc, locErr := client.LocationService.GetLocationByIDV3(ctx, locID)
		if locErr != nil {
			t.Logf("pick%sKey: key %s resolved but could not look up location %d: %v", partner, masked, locID, locErr)
			continue
		}
		market := loc.Market

		cspClaimedMu.Lock()
		if cspClaimedKeys[key] || cspClaimedPorts[resp.ProductUID] {
			cspClaimedMu.Unlock()
			t.Logf("pick%sKey: key %s or port already claimed, skipping", partner, masked)
			continue
		}
		cspClaimedKeys[key] = true
		cspClaimedPorts[resp.ProductUID] = true
		claimedKey := key
		claimedPort := resp.ProductUID
		cspClaimedMu.Unlock()

		t.Cleanup(func() {
			cspClaimedMu.Lock()
			delete(cspClaimedKeys, claimedKey)
			delete(cspClaimedPorts, claimedPort)
			cspClaimedMu.Unlock()
		})

		t.Logf("pick%sKey: using key %s (location %d, market %s)", partner, masked, locID, market)
		return cspPickResult{Key: key, PartnerPortUID: resp.ProductUID, LocationID: locID, Market: market}
	}

	t.Skipf("skipping: no %s key with available capacity found", partner)
	return cspPickResult{}
}

// awsPickResult holds a selected AWS partner port together with the market
// code of its location. Tests should use Market for BuyPortRequest / BuyMCRRequest
// so the A-End is always placed in the same market as the partner port.
type awsPickResult struct {
	Port   *PartnerMegaport
	Market string
}

// pickAWSPartnerPort finds a live AWS partner port in staging matching the
// given AWS connect type (CONNECT_TYPE_AWS_VIF or
// CONNECT_TYPE_AWS_HOSTED_CONNECTION). AWS does not need a CSP key from the
// customer — the B-End is just one of Megaport's AWS partner ports. The
// returned awsPickResult includes the market code derived from the partner
// port's location so callers don't need to hardcode a market. Calls t.Skip
// if none found.
func pickAWSPartnerPort(t *testing.T, client *Client, connectType string) awsPickResult {
	t.Helper()
	ctx := context.Background()
	partners, err := client.PartnerService.ListPartnerMegaports(ctx)
	if err != nil {
		t.Skipf("skipping: could not list partner ports: %v", err)
		return awsPickResult{}
	}
	filtered, err := client.PartnerService.FilterPartnerMegaportByConnectType(ctx, partners, connectType, true)
	if err != nil {
		t.Skipf("skipping: could not filter partner ports: %v", err)
		return awsPickResult{}
	}
	if len(filtered) == 0 {
		t.Skipf("skipping: no %s partner ports available in staging", connectType)
		return awsPickResult{}
	}
	//nolint:gosec // weak random is fine for test selection
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := filtered[r.Intn(len(filtered))]

	loc, locErr := client.LocationService.GetLocationByIDV3(ctx, port.LocationId)
	if locErr != nil {
		t.Skipf("skipping: could not look up location %d for partner port %s: %v", port.LocationId, port.ProductUID, locErr)
		return awsPickResult{}
	}

	return awsPickResult{Port: port, Market: loc.Market}
}

// loadCSPCredentials reads the CSP credentials pool from the
// CSP_CREDENTIALS_JSON env var (preferred for CI) or the local
// testdata/csp_credentials.json file (developer convenience). A missing file
// is not an error — tests will skip on empty pools, which is the intended
// behaviour when running without CSP credentials.
func loadCSPCredentials() (cspCredentials, error) {
	if raw := os.Getenv("CSP_CREDENTIALS_JSON"); raw != "" {
		var creds cspCredentials
		if err := json.Unmarshal([]byte(raw), &creds); err != nil {
			return cspCredentials{}, fmt.Errorf("CSP_CREDENTIALS_JSON: %w", err)
		}
		return creds, nil
	}
	data, err := os.ReadFile("testdata/csp_credentials.json")
	if errors.Is(err, os.ErrNotExist) {
		return cspCredentials{}, fmt.Errorf("testdata/csp_credentials.json: %w", os.ErrNotExist)
	}
	if err != nil {
		return cspCredentials{}, fmt.Errorf("testdata/csp_credentials.json: %w", err)
	}
	var creds cspCredentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return cspCredentials{}, fmt.Errorf("testdata/csp_credentials.json: %w", err)
	}
	return creds, nil
}
