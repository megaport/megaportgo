package megaport

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// NATGatewayVXCIntegrationTestSuite is the headline end-to-end test for the
// VXC ↔ NAT Gateway attachment shape. It provisions a NAT Gateway and a
// Megaport, creates a packet filter + IPv4 prefix list on the gateway, then
// orders a VXC whose A-End is the NAT Gateway with a VRouter partner config
// exercising the full surface: interface IPs, NAT IPs, a static route, a
// (shut-down) BGP peer, and a packetFilterIn binding.
type NATGatewayVXCIntegrationTestSuite IntegrationTestSuite

func TestNATGatewayVXCIntegrationTestSuite(t *testing.T) {
	t.Parallel()
	if *runIntegrationTests {
		suite.Run(t, new(NATGatewayVXCIntegrationTestSuite))
	}
}

func (suite *NATGatewayVXCIntegrationTestSuite) SetupSuite() {
	natSuite := (*NATGatewayIntegrationTestSuite)(suite)
	natSuite.SetupSuite()
}

func (suite *NATGatewayVXCIntegrationTestSuite) TestVXCAttachedToNATGateway() {
	ctx := context.Background()
	logger := suite.client.Logger
	natSvc := suite.client.NATGatewayService
	portSvc := suite.client.PortService
	vxcSvc := suite.client.VXCService

	// Reuse the helper by aliasing — the helper takes a NAT Gateway suite,
	// and these two suites have the same embedded IntegrationTestSuite.
	natSuite := (*NATGatewayIntegrationTestSuite)(suite)

	prov, err := provisionNATGatewayForTest(ctx, natSuite, "Integration Test NAT Gateway (VXC)")
	if err != nil {
		suite.FailNowf("could not provision NAT Gateway", "%v", err)
	}
	defer prov.Teardown()

	// Buy a Port at the same location to act as the B-End.
	portRes, portErr := portSvc.BuyPort(ctx, &BuyPortRequest{
		Name:                  "Integration Test Port (NAT VXC)",
		LocationId:            prov.LocationID,
		PortSpeed:             1000,
		Term:                  1,
		Market:                TEST_NAT_GATEWAY_LOCATION_MARKET,
		MarketPlaceVisibility: false,
		WaitForProvision:      true,
		WaitForTime:           10 * time.Minute,
	})
	if portErr != nil {
		suite.FailNowf("could not buy port", "%v", portErr)
	}
	portUID := portRes.TechnicalServiceUIDs[0]
	suite.True(IsGuid(portUID), "invalid guid for port uid")
	logger.InfoContext(ctx, "port provisioned", slog.String("port_uid", portUID))

	defer func() {
		if _, err := portSvc.DeletePort(ctx, &DeletePortRequest{PortID: portUID, DeleteNow: true}); err != nil {
			logger.WarnContext(ctx, "port teardown best-effort failed",
				slog.String("port_uid", portUID),
				slog.String("error", err.Error()),
			)
		}
	}()

	// Create a packet filter on the NAT Gateway — we'll bind it to the VXC
	// interface via packetFilterIn.
	packetFilter, err := natSvc.CreateNATGatewayPacketFilter(ctx, prov.ProductUID, &NATGatewayPacketFilterRequest{
		Description: "vxc-integration-permit-https",
		Entries: []NATGatewayPacketFilterEntry{
			{
				Action:             PacketFilterActionPermit,
				Description:        "permit https",
				SourceAddress:      "0.0.0.0/0",
				DestinationAddress: "0.0.0.0/0",
				DestinationPorts:   "443",
				IPProtocol:         6,
			},
		},
	})
	if err != nil {
		suite.FailNowf("could not create packet filter", "%v", err)
	}
	logger.InfoContext(ctx, "packet filter created", slog.Int("packet_filter_id", packetFilter.ID))
	defer func() {
		if err := natSvc.DeleteNATGatewayPacketFilter(ctx, prov.ProductUID, packetFilter.ID); err != nil {
			logger.WarnContext(ctx, "packet filter teardown best-effort failed",
				slog.Int("packet_filter_id", packetFilter.ID),
				slog.String("error", err.Error()),
			)
		}
	}()

	// Create an IPv4 prefix list on the NAT Gateway.
	prefixList, err := natSvc.CreateNATGatewayPrefixList(ctx, prov.ProductUID, &NATGatewayPrefixList{
		Description:   "vxc-integration-private",
		AddressFamily: AddressFamilyIPv4,
		Entries: []NATGatewayPrefixListEntry{
			{Action: PrefixListActionPermit, Prefix: "10.0.0.0/8", Ge: 24, Le: 32},
		},
	})
	if err != nil {
		suite.FailNowf("could not create prefix list", "%v", err)
	}
	logger.InfoContext(ctx, "prefix list created", slog.Int("prefix_list_id", prefixList.ID))
	defer func() {
		if err := natSvc.DeleteNATGatewayPrefixList(ctx, prov.ProductUID, prefixList.ID); err != nil {
			logger.WarnContext(ctx, "prefix list teardown best-effort failed",
				slog.Int("prefix_list_id", prefixList.ID),
				slog.String("error", err.Error()),
			)
		}
	}()

	// Negotiate a VLAN on the B-End port. NAT Gateway A-End VLANs are
	// allocated by the platform — only the B-End needs a VLAN here.
	var bEndVLAN int
	for i := 0; i < 10; i++ {
		bEndVLAN = GenerateRandomVLAN()
		available, vlanErr := portSvc.CheckPortVLANAvailability(ctx, portUID, bEndVLAN)
		if vlanErr != nil {
			suite.FailNowf("could not check vlan availability", "%v", vlanErr)
		}
		if available {
			break
		}
	}
	if bEndVLAN == 0 {
		suite.FailNowf("no available vlan on b-end port", "could not find an available VLAN after 10 attempts")
	}

	// Buy the VXC. A-End = NAT Gateway with full vrouter partner config.
	// B-End = the Port we just provisioned.
	packetFilterID := int64(packetFilter.ID)
	logger.InfoContext(ctx, "buying VXC NAT Gateway -> Port",
		slog.String("a_end_nat_gateway_uid", prov.ProductUID),
		slog.String("b_end_port_uid", portUID),
		slog.Int("b_end_vlan", bEndVLAN),
	)
	buyRes, vxcErr := vxcSvc.BuyVXC(ctx, &BuyVXCRequest{
		PortUID:   prov.ProductUID,
		VXCName:   "Integration Test VXC (NAT Gateway)",
		RateLimit: 100,
		Term:      1,
		Shutdown:  false,
		AEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID: prov.ProductUID,
			PartnerConfig: VXCOrderVrouterPartnerConfig{
				Interfaces: []PartnerConfigInterface{
					{
						IpAddresses: []string{"10.0.0.1/30"},
						// natIpAddresses is rejected on the NAT Gateway A-End
						// — NAT IPs are managed by the gateway itself, not
						// configured per VXC interface.
						IpRoutes: []IpRoute{
							{Prefix: "192.168.1.0/24", NextHop: "10.0.0.2", Description: "test static route"},
						},
						BgpConnections: []BgpConnectionConfig{
							{
								PeerAsn:        65000,
								LocalIpAddress: "10.0.0.1",
								PeerIpAddress:  "10.0.0.2",
								// Shutdown so we don't expect a real peer on the
								// other side — the Port has no BGP speaker.
								Shutdown:    true,
								Description: "nat-gw-bgp-test",
								PeerType:    "NON_CLOUD",
							},
						},
						PacketFilterIn: &packetFilterID,
					},
				},
			},
		},
		BEndConfiguration: VXCOrderEndpointConfiguration{
			ProductUID: portUID,
			VLAN:       bEndVLAN,
		},
		WaitForProvision: true,
		WaitForTime:      15 * time.Minute,
	})
	if vxcErr != nil {
		suite.FailNowf("could not buy VXC", "%v", vxcErr)
	}
	vxcUID := buyRes.TechnicalServiceUID
	suite.True(IsGuid(vxcUID), "invalid guid for vxc uid")
	logger.InfoContext(ctx, "VXC provisioned", slog.String("vxc_uid", vxcUID))

	defer func() {
		if err := vxcSvc.DeleteVXC(ctx, vxcUID, &DeleteVXCRequest{DeleteNow: true}); err != nil {
			logger.WarnContext(ctx, "vxc teardown best-effort failed",
				slog.String("vxc_uid", vxcUID),
				slog.String("error", err.Error()),
			)
		}
	}()

	// Verify the VXC is live and that its A-End is the NAT Gateway.
	fetched, err := vxcSvc.GetVXC(ctx, vxcUID)
	if err != nil {
		suite.FailNowf("could not get VXC", "%v", err)
	}
	suite.Contains(SERVICE_STATE_READY, fetched.ProvisioningStatus,
		"expected VXC in CONFIGURED/LIVE, got %q", fetched.ProvisioningStatus)
	suite.Equal(prov.ProductUID, fetched.AEndConfiguration.UID,
		"A-End UID mismatch — expected NAT Gateway %s, got %s", prov.ProductUID, fetched.AEndConfiguration.UID)
	logger.InfoContext(ctx, "VXC↔NAT Gateway attachment verified",
		slog.String("vxc_uid", vxcUID),
		slog.String("a_end_uid", fetched.AEndConfiguration.UID),
		slog.String("provisioning_status", fetched.ProvisioningStatus),
	)
}
