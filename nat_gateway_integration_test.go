package megaport

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"slices"
	"testing"
	"time"
)

const (
	TEST_NAT_GATEWAY_LOCATION_MARKET = "AU"
)

// NATGatewayIntegrationTestSuite is the integration test suite for the NAT Gateway service.
type NATGatewayIntegrationTestSuite IntegrationTestSuite

func TestNATGatewayIntegrationTestSuite(t *testing.T) {
	runIntegrationMethods[NATGatewayIntegrationTestSuite](t)
}

func (suite *NATGatewayIntegrationTestSuite) SetupSuite() {
	accessKey := os.Getenv("MEGAPORT_ACCESS_KEY")
	secretKey := os.Getenv("MEGAPORT_SECRET_KEY")

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	programLevel.Set(slog.LevelDebug)

	megaportClient, err := New(nil, WithBaseURL(MEGAPORTURL), WithLogHandler(handler), WithCredentials(accessKey, secretKey))
	if err != nil {
		suite.FailNowf("", "could not initialize megaport test client: %s", err.Error())
	}

	_, err = megaportClient.Authorize(ctx)
	if err != nil {
		suite.FailNowf("", "could not authorize megaport test client: %s", err.Error())
	}

	suite.client = megaportClient
}

// TestNATGatewayLifecycle tests the full CRUD lifecycle of a NAT Gateway.
func (suite *NATGatewayIntegrationTestSuite) TestNATGatewayLifecycle() {
	ctx := context.Background()
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService

	// Step 1: List available sessions to pick a valid speed/session count.
	logger.DebugContext(ctx, "Listing NAT Gateway sessions.")
	sessions, err := natSvc.ListNATGatewaySessions(ctx)
	if err != nil {
		suite.FailNowf("could not list sessions", "could not list NAT Gateway sessions: %v", err)
	}
	suite.NotEmpty(sessions, "expected at least one session configuration")

	testSpeed := sessions[0].SpeedMbps
	testSessionCount := sessions[0].SessionCount[0]
	logger.DebugContext(ctx, "Selected session config",
		slog.Int("speed", testSpeed),
		slog.Int("session_count", testSessionCount),
	)

	// Step 2: Pick a location that advertises NAT Gateway support at the chosen speed.
	testLocation, locErr := findActiveNATGatewayLocation(ctx, suite.T(), suite.client, TEST_NAT_GATEWAY_LOCATION_MARKET, testSpeed)
	if locErr != nil {
		suite.FailNowf("could not get nat gateway location", "could not get nat gateway location: %v", locErr)
	}
	suite.NotNil(testLocation)
	logger.DebugContext(ctx, "Test location determined", slog.String("location", testLocation.Name), slog.Int("location_id", testLocation.ID))

	// Step 3: Create a NAT Gateway (stays in NEW status, not provisioned).
	logger.DebugContext(ctx, "Creating NAT Gateway.")
	createReq := &CreateNATGatewayRequest{
		AutoRenewTerm: true,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: false,
			DiversityZone:      "red",
			SessionCount:       testSessionCount,
		},
		LocationID:  testLocation.ID,
		ProductName: "Integration Test NAT Gateway",
		Speed:       testSpeed,
		Term:        1,
	}
	gw, err := natSvc.CreateNATGateway(ctx, createReq)
	if err != nil {
		suite.FailNowf("could not create NAT Gateway", "could not create NAT Gateway: %v", err)
	}
	suite.NotEmpty(gw.ProductUID)
	suite.Equal("Integration Test NAT Gateway", gw.ProductName)
	suite.Equal(testSpeed, gw.Speed)
	suite.Equal(1, gw.Term)
	logger.DebugContext(ctx, "NAT Gateway created", slog.String("product_uid", gw.ProductUID), slog.String("provisioning_status", gw.ProvisioningStatus))

	productUID := gw.ProductUID

	// Step 4: Get the NAT Gateway by UID.
	logger.DebugContext(ctx, "Retrieving NAT Gateway by UID.")
	fetched, err := natSvc.GetNATGateway(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not get NAT Gateway", "could not get NAT Gateway: %v", err)
	}
	suite.Equal(productUID, fetched.ProductUID)
	suite.Equal("Integration Test NAT Gateway", fetched.ProductName)
	suite.Equal(testLocation.ID, fetched.LocationID)

	// Step 5: List NAT Gateways and verify ours appears.
	logger.DebugContext(ctx, "Listing NAT Gateways.")
	gateways, err := natSvc.ListNATGateways(ctx)
	if err != nil {
		suite.FailNowf("could not list NAT Gateways", "could not list NAT Gateways: %v", err)
	}
	found := false
	for _, g := range gateways {
		if g.ProductUID == productUID {
			found = true
			break
		}
	}
	suite.True(found, "created NAT Gateway not found in list")

	// Step 6: Update the NAT Gateway.
	logger.DebugContext(ctx, "Updating NAT Gateway.")
	updated, err := natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID:    productUID,
		AutoRenewTerm: false,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: true,
			DiversityZone:      "red",
			SessionCount:       testSessionCount,
		},
		LocationID:  testLocation.ID,
		ProductName: "Integration Test NAT Gateway [Updated]",
		Speed:       testSpeed,
		Term:        1,
	})
	if err != nil {
		suite.FailNowf("could not update NAT Gateway", "could not update NAT Gateway: %v", err)
	}
	suite.Equal("Integration Test NAT Gateway [Updated]", updated.ProductName)
	suite.False(updated.AutoRenewTerm)
	logger.DebugContext(ctx, "NAT Gateway updated", slog.String("product_name", updated.ProductName))

	// Step 7: Delete the DESIGN-state NAT Gateway. DeleteNATGateway inspects
	// ProvisioningStatus internally and routes to the design-only DELETE
	// endpoint for gateways that have never been purchased.
	logger.DebugContext(ctx, "Deleting NAT Gateway.")
	err = natSvc.DeleteNATGateway(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not delete NAT Gateway", "could not delete NAT Gateway: %v", err)
	}
	logger.DebugContext(ctx, "NAT Gateway deleted", slog.String("product_uid", productUID))

	// Step 8: Verify the NAT Gateway is gone. DESIGN-state deletes hard-remove
	// the record, so it should no longer appear in the list.
	postDeleteList, err := natSvc.ListNATGateways(ctx)
	if err != nil {
		suite.FailNowf("could not list NAT Gateways post-delete", "could not list NAT Gateways: %v", err)
	}
	found = false
	for _, g := range postDeleteList {
		if g.ProductUID == productUID {
			found = true
			break
		}
	}
	suite.False(found, "deleted DESIGN NAT Gateway %s still appears in list", productUID)
	logger.InfoContext(ctx, "design NAT Gateway teardown verified (hard-removed from list)", slog.String("product_uid", productUID))
}

// TestNATGatewayFullLifecycle exercises the end-to-end flow: create the
// design record, validate and buy the gateway via the network-design
// endpoints, wait for it to reach CONFIGURED/LIVE, update a mutable field,
// and tear down via ProductService (the DESIGN-only DELETE endpoint no
// longer applies once the order has been bought).
func (suite *NATGatewayIntegrationTestSuite) TestNATGatewayFullLifecycle() {
	ctx := context.Background()
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService

	// Step 1: List sessions to pick the smallest speed/session-count combo
	// (small speeds are supported at more locations).
	sessions, err := natSvc.ListNATGatewaySessions(ctx)
	if err != nil {
		suite.FailNowf("could not list sessions", "could not list NAT Gateway sessions: %v", err)
	}
	suite.NotEmpty(sessions, "expected at least one session configuration")
	minSession := sessions[0]
	for _, s := range sessions[1:] {
		if s.SpeedMbps < minSession.SpeedMbps {
			minSession = s
		}
	}
	testSpeed := minSession.SpeedMbps
	testSessionCount := minSession.SessionCount[0]
	logger.InfoContext(ctx, "Selected session config",
		slog.Int("speed", testSpeed),
		slog.Int("session_count", testSessionCount),
	)

	// Step 2: Pick a location that advertises NAT Gateway support at the
	// chosen speed. v3/locations surfaces availability via
	// diversityZones.{red,blue}.natGatewaySpeedMbps, so we can filter
	// up front instead of probing with validate.
	locations, err := suite.client.LocationService.ListLocationsV3(ctx)
	if err != nil {
		suite.FailNowf("could not list locations", "could not list locations: %v", err)
	}
	marketLocations, err := suite.client.LocationService.FilterLocationsByMarketCodeV3(ctx, TEST_NAT_GATEWAY_LOCATION_MARKET, locations)
	if err != nil {
		suite.FailNowf("could not filter by market", "could not filter by market: %v", err)
	}
	eligible := suite.client.LocationService.FilterLocationsByNATGatewaySpeedV3(ctx, testSpeed, marketLocations)
	if len(eligible) == 0 {
		suite.FailNowf("no eligible location", "no location in market %q advertises NAT Gateway speed %d", TEST_NAT_GATEWAY_LOCATION_MARKET, testSpeed)
	}
	testLocation := eligible[0]
	logger.InfoContext(ctx, "Selected eligible location",
		slog.String("location", testLocation.Name),
		slog.Int("location_id", testLocation.ID),
		slog.Int("eligible_count", len(eligible)),
	)

	// Step 3: Create the NAT Gateway design.
	gw, err := natSvc.CreateNATGateway(ctx, &CreateNATGatewayRequest{
		AutoRenewTerm: false,
		Config: NATGatewayNetworkConfig{
			ASN:                64512,
			BGPShutdownDefault: false,
			DiversityZone:      "red",
			SessionCount:       testSessionCount,
		},
		LocationID:  testLocation.ID,
		ProductName: "Integration Test NAT Gateway (Full Lifecycle)",
		Speed:       testSpeed,
		Term:        1,
	})
	if err != nil {
		suite.FailNowf("could not create NAT Gateway", "could not create NAT Gateway: %v", err)
	}
	productUID := gw.ProductUID
	suite.NotEmpty(productUID)
	logger.InfoContext(ctx, "NAT Gateway design created",
		slog.String("product_uid", productUID),
		slog.String("provisioning_status", gw.ProvisioningStatus),
	)

	// Teardown: DeleteNATGateway handles any lifecycle stage the gateway
	// might be in when the defer fires. This is a safety net — the happy
	// path below performs an explicit delete + verification — so skip the
	// call if the gateway is already in a terminal state.
	defer func() {
		logger.InfoContext(ctx, "Tearing down NAT Gateway", slog.String("product_uid", productUID))

		// deleteDesign hits the design-only endpoint directly, bypassing the
		// pre-flight GET inside DeleteNATGateway so the teardown never
		// incurs a second GET on the happy DESIGN path and stays usable when
		// state-inspection is down.
		deleteDesign := func() error {
			path := fmt.Sprintf("/v3/products/nat_gateways/%s", url.PathEscape(productUID))
			req, err := suite.client.NewRequest(ctx, http.MethodDelete, path, nil)
			if err != nil {
				return err
			}
			resp, err := suite.client.Do(ctx, req, nil)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			return nil
		}
		cancelNow := func() error {
			_, err := suite.client.ProductService.DeleteProduct(ctx, &DeleteProductRequest{
				ProductID: productUID,
				DeleteNow: true,
			})
			return err
		}

		current, getErr := natSvc.GetNATGateway(ctx, productUID)
		if getErr != nil {
			// Can't tell which state we're in — attempt both paths so we do
			// not leak either a DESIGN record or a provisioned gateway. Each
			// endpoint returns 400 for the wrong state; we log but don't
			// fail the teardown on those.
			logger.WarnContext(ctx, "teardown: could not inspect state, attempting both cleanup paths",
				slog.String("product_uid", productUID),
				slog.String("error", getErr.Error()),
			)
			if err := deleteDesign(); err != nil {
				logger.WarnContext(ctx, "teardown (DESIGN DELETE) best-effort failed",
					slog.String("product_uid", productUID),
					slog.String("error", err.Error()),
				)
			}
			if err := cancelNow(); err != nil {
				logger.WarnContext(ctx, "teardown (CANCEL_NOW) best-effort failed",
					slog.String("product_uid", productUID),
					slog.String("error", err.Error()),
				)
			}
			return
		}
		if current.ProvisioningStatus == STATUS_DECOMMISSIONED ||
			current.ProvisioningStatus == STATUS_CANCELLED {
			logger.InfoContext(ctx, "teardown skipped: already in terminal state",
				slog.String("product_uid", productUID),
				slog.String("provisioning_status", current.ProvisioningStatus),
			)
			return
		}
		// Dispatch directly from the state we already fetched — no second GET.
		var dErr error
		if current.ProvisioningStatus == STATUS_DESIGN {
			dErr = deleteDesign()
		} else {
			dErr = cancelNow()
		}
		if dErr != nil {
			logger.WarnContext(ctx, "teardown failed",
				slog.String("product_uid", productUID),
				slog.String("provisioning_status", current.ProvisioningStatus),
				slog.String("error", dErr.Error()),
			)
		}
	}()

	// Step 4: Validate the order (pricing preview).
	validation, err := natSvc.ValidateNATGatewayOrder(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not validate NAT Gateway", "could not validate NAT Gateway: %v", err)
	}
	suite.Equal(productUID, validation.ProductUID)
	logger.InfoContext(ctx, "NAT Gateway order validated",
		slog.String("product_uid", productUID),
		slog.Float64("monthly_rate", validation.Price.MonthlyRate),
		slog.String("currency", validation.Price.Currency),
	)

	// Step 5: Buy the gateway via /v3/networkdesign/buy — kicks off provisioning.
	bought, err := natSvc.BuyNATGateway(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not buy NAT Gateway", "could not buy NAT Gateway: %v", err)
	}
	suite.Equal(productUID, bought.ProductUID)
	logger.InfoContext(ctx, "NAT Gateway order bought",
		slog.String("product_uid", productUID),
		slog.String("provisioning_status", bought.ProvisioningStatus),
	)

	// Step 6: Poll until the gateway reaches CONFIGURED/LIVE, or fail fast
	// on a terminal error state.
	const (
		pollInterval = 10 * time.Second
		pollTimeout  = 15 * time.Minute
	)
	pollCtx, cancel := context.WithTimeout(ctx, pollTimeout)
	defer cancel()

	var provisioned *NATGateway
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

PollLoop:
	for {
		fetched, getErr := natSvc.GetNATGateway(pollCtx, productUID)
		if getErr != nil {
			suite.FailNowf("could not poll NAT Gateway", "error while polling NAT Gateway %s: %v", productUID, getErr)
		}
		logger.DebugContext(pollCtx, "poll",
			slog.String("product_uid", productUID),
			slog.String("provisioning_status", fetched.ProvisioningStatus),
		)
		switch {
		case slices.Contains(SERVICE_STATE_READY, fetched.ProvisioningStatus):
			provisioned = fetched
			break PollLoop
		case fetched.ProvisioningStatus == STATUS_DECOMMISSIONED ||
			fetched.ProvisioningStatus == STATUS_CANCELLED:
			suite.FailNowf("NAT Gateway reached terminal state", "gateway %s reached %s", productUID, fetched.ProvisioningStatus)
		}

		select {
		case <-pollCtx.Done():
			suite.FailNowf("timed out waiting for provisioning", "gateway %s did not reach CONFIGURED/LIVE within %s (last status %q)", productUID, pollTimeout, fetched.ProvisioningStatus)
		case <-ticker.C:
		}
	}

	suite.NotNil(provisioned)
	suite.Contains(SERVICE_STATE_READY, provisioned.ProvisioningStatus)
	logger.InfoContext(ctx, "NAT Gateway provisioned",
		slog.String("product_uid", productUID),
		slog.String("provisioning_status", provisioned.ProvisioningStatus),
	)

	// Step 7: Update a field that remains mutable after deployment
	// (productName). Speed/location/promoCode are immutable post-deploy per
	// the API docs.
	const updatedName = "Integration Test NAT Gateway (Updated)"
	updated, err := natSvc.UpdateNATGateway(ctx, &UpdateNATGatewayRequest{
		ProductUID:    productUID,
		AutoRenewTerm: false,
		Config: NATGatewayNetworkConfig{
			ASN:                provisioned.Config.ASN,
			BGPShutdownDefault: provisioned.Config.BGPShutdownDefault,
			DiversityZone:      provisioned.Config.DiversityZone,
			SessionCount:       provisioned.Config.SessionCount,
		},
		LocationID:  provisioned.LocationID,
		ProductName: updatedName,
		Speed:       provisioned.Speed,
		Term:        provisioned.Term,
	})
	if err != nil {
		suite.FailNowf("could not update NAT Gateway", "could not update provisioned NAT Gateway: %v", err)
	}
	suite.Equal(updatedName, updated.ProductName)
	logger.InfoContext(ctx, "NAT Gateway updated", slog.String("product_name", updated.ProductName))

	// Step 8: Delete the provisioned NAT Gateway and verify it's gone.
	// DeleteNATGateway uses the generic product-cancellation endpoint so it
	// works even though the gateway is no longer in DESIGN state. This is
	// the primary deletion path; the defer above remains as a safety net if
	// any earlier step failed before reaching here.
	//
	// Unlike the DESIGN-state delete in TestNATGatewayLifecycle (which
	// hard-removes the record so verification is by list-exclusion), the
	// provisioned teardown via CANCEL_NOW leaves a tombstone record behind,
	// so verification is by GET on the product UID and a check that the
	// record has moved to DECOMMISSIONED/CANCELLED.
	if err := natSvc.DeleteNATGateway(ctx, productUID); err != nil {
		suite.FailNowf("could not delete provisioned NAT Gateway", "could not delete NAT Gateway: %v", err)
	}
	postDelete, err := natSvc.GetNATGateway(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not GET deleted NAT Gateway", "could not GET deleted NAT Gateway: %v", err)
	}
	suite.Contains(
		[]string{STATUS_DECOMMISSIONED, STATUS_CANCELLED},
		postDelete.ProvisioningStatus,
		"expected deleted NAT Gateway to be DECOMMISSIONED or CANCELLED, got %q",
		postDelete.ProvisioningStatus,
	)
	logger.InfoContext(ctx, "provisioned NAT Gateway teardown verified (tombstone record in terminal state)",
		slog.String("product_uid", productUID),
		slog.String("provisioning_status", postDelete.ProvisioningStatus),
	)
}

// TestNATGatewayPacketFilterLifecycle exercises the packet filter CRUD
// surface against a live, provisioned NAT Gateway: create with two entries,
// fetch + verify, list (assert summary present), update (third entry,
// changed description), delete, list (assert gone).
func (suite *NATGatewayIntegrationTestSuite) TestNATGatewayPacketFilterLifecycle() {
	ctx := context.Background()
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService

	prov, err := provisionNATGatewayForTest(ctx, suite, "Integration Test NAT Gateway (Packet Filter)")
	if err != nil {
		suite.FailNowf("could not provision NAT Gateway", "%v", err)
	}
	defer prov.Teardown()
	productUID := prov.ProductUID

	// Create — 2 entries: permit TCP/443 inbound, deny everything else.
	createReq := &NATGatewayPacketFilterRequest{
		Description: "integration-test-filter",
		Entries: []NATGatewayPacketFilterEntry{
			{
				Action:             PacketFilterActionPermit,
				Description:        "permit https",
				SourceAddress:      "0.0.0.0/0",
				DestinationAddress: "0.0.0.0/0",
				DestinationPorts:   "443",
				IPProtocol:         6, // TCP
			},
			{
				Action:             PacketFilterActionDeny,
				Description:        "deny everything else",
				SourceAddress:      "0.0.0.0/0",
				DestinationAddress: "0.0.0.0/0",
			},
		},
	}
	created, err := natSvc.CreateNATGatewayPacketFilter(ctx, productUID, createReq)
	if err != nil {
		suite.FailNowf("could not create packet filter", "%v", err)
	}
	suite.NotZero(created.ID)
	suite.Equal("integration-test-filter", created.Description)
	suite.Len(created.Entries, 2)
	logger.InfoContext(ctx, "packet filter created", slog.Int("packet_filter_id", created.ID))

	// Get.
	fetched, err := natSvc.GetNATGatewayPacketFilter(ctx, productUID, created.ID)
	if err != nil {
		suite.FailNowf("could not get packet filter", "%v", err)
	}
	suite.Equal(created.ID, fetched.ID)
	suite.Equal(createReq.Description, fetched.Description)
	suite.Len(fetched.Entries, 2)

	// List — must include the new filter.
	summaries, err := natSvc.ListNATGatewayPacketFilters(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not list packet filters", "%v", err)
	}
	found := false
	for _, s := range summaries {
		if s.ID == created.ID {
			found = true
			suite.Equal("integration-test-filter", s.Description)
			break
		}
	}
	suite.True(found, "created packet filter not present in summary list")

	// Update — append a third entry, change description.
	updateReq := &NATGatewayPacketFilterRequest{
		Description: "integration-test-filter [updated]",
		Entries: append(append([]NATGatewayPacketFilterEntry{},
			createReq.Entries...),
			NATGatewayPacketFilterEntry{
				Action:             PacketFilterActionPermit,
				Description:        "permit dns",
				SourceAddress:      "0.0.0.0/0",
				DestinationAddress: "0.0.0.0/0",
				DestinationPorts:   "53",
				IPProtocol:         17, // UDP
			},
		),
	}
	updated, err := natSvc.UpdateNATGatewayPacketFilter(ctx, productUID, created.ID, updateReq)
	if err != nil {
		suite.FailNowf("could not update packet filter", "%v", err)
	}
	suite.Equal("integration-test-filter [updated]", updated.Description)
	suite.Len(updated.Entries, 3)

	// Delete.
	if err := natSvc.DeleteNATGatewayPacketFilter(ctx, productUID, created.ID); err != nil {
		suite.FailNowf("could not delete packet filter", "%v", err)
	}

	// List — must NOT include the deleted filter.
	postDelete, err := natSvc.ListNATGatewayPacketFilters(ctx, productUID)
	if err != nil {
		suite.FailNowf("could not list packet filters post-delete", "%v", err)
	}
	for _, s := range postDelete {
		suite.NotEqual(created.ID, s.ID, "deleted packet filter %d still in summary list", created.ID)
	}
	logger.InfoContext(ctx, "packet filter lifecycle complete", slog.Int("packet_filter_id", created.ID))
}

// TestNATGatewayPrefixListLifecycle exercises the prefix list CRUD surface
// against a provisioned NAT Gateway, once for IPv4 (with ge/le) and once
// for IPv6 (without). Confirms the string<->int conversion round-trips.
func (suite *NATGatewayIntegrationTestSuite) TestNATGatewayPrefixListLifecycle() {
	ctx := context.Background()
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService

	prov, err := provisionNATGatewayForTest(ctx, suite, "Integration Test NAT Gateway (Prefix List)")
	if err != nil {
		suite.FailNowf("could not provision NAT Gateway", "%v", err)
	}
	defer prov.Teardown()
	productUID := prov.ProductUID

	cases := []struct {
		name               string
		create             *NATGatewayPrefixList
		expectGe, expectLe int
		// extraPrefix is appended on update — must differ from the create
		// entries' prefix to avoid the API's duplicate-prefix rejection.
		extraPrefix string
	}{
		{
			name: "ipv4-with-ge-le",
			create: &NATGatewayPrefixList{
				Description:   "integration-test-v4",
				AddressFamily: AddressFamilyIPv4,
				Entries: []NATGatewayPrefixListEntry{
					{Action: PrefixListActionPermit, Prefix: "10.0.0.0/8", Ge: 24, Le: 32},
				},
			},
			expectGe:    24,
			expectLe:    32,
			extraPrefix: "172.16.0.0/12",
		},
		{
			name: "ipv6-no-ge-le",
			create: &NATGatewayPrefixList{
				Description:   "integration-test-v6",
				AddressFamily: AddressFamilyIPv6,
				Entries: []NATGatewayPrefixListEntry{
					{Action: PrefixListActionDeny, Prefix: "2001:db8::/32"},
				},
			},
			extraPrefix: "2001:db8:1::/48",
		},
	}

	for _, tc := range cases {
		tc := tc
		suite.Run(tc.name, func() {
			created, err := natSvc.CreateNATGatewayPrefixList(ctx, productUID, tc.create)
			if err != nil {
				suite.FailNowf("could not create prefix list", "%v", err)
			}
			suite.NotZero(created.ID)
			suite.Equal(tc.create.AddressFamily, created.AddressFamily)
			suite.Len(created.Entries, 1)
			suite.Equal(tc.expectGe, created.Entries[0].Ge)
			suite.Equal(tc.expectLe, created.Entries[0].Le)
			logger.InfoContext(ctx, "prefix list created",
				slog.Int("prefix_list_id", created.ID),
				slog.String("address_family", created.AddressFamily),
			)

			// Get.
			fetched, err := natSvc.GetNATGatewayPrefixList(ctx, productUID, created.ID)
			if err != nil {
				suite.FailNowf("could not get prefix list", "%v", err)
			}
			suite.Equal(created.ID, fetched.ID)
			suite.Equal(tc.expectGe, fetched.Entries[0].Ge)
			suite.Equal(tc.expectLe, fetched.Entries[0].Le)

			// List.
			summaries, err := natSvc.ListNATGatewayPrefixLists(ctx, productUID)
			if err != nil {
				suite.FailNowf("could not list prefix lists", "%v", err)
			}
			found := false
			for _, s := range summaries {
				if s.ID == created.ID {
					found = true
					suite.Equal(tc.create.AddressFamily, s.AddressFamily)
					break
				}
			}
			suite.True(found, "created prefix list not present in summary list")

			// Update — keep the original entry and append a second one with a
			// distinct prefix (the API rejects duplicates).
			updateReq := &NATGatewayPrefixList{
				Description:   tc.create.Description + " [updated]",
				AddressFamily: tc.create.AddressFamily,
				Entries: []NATGatewayPrefixListEntry{
					tc.create.Entries[0],
					{Action: PrefixListActionPermit, Prefix: tc.extraPrefix},
				},
			}
			updated, err := natSvc.UpdateNATGatewayPrefixList(ctx, productUID, created.ID, updateReq)
			if err != nil {
				suite.FailNowf("could not update prefix list", "%v", err)
			}
			suite.Len(updated.Entries, 2)

			// Delete.
			if err := natSvc.DeleteNATGatewayPrefixList(ctx, productUID, created.ID); err != nil {
				suite.FailNowf("could not delete prefix list", "%v", err)
			}
		})
	}
}
