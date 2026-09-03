# Unreleased

## New Features
- Add `WithCallContext` client option that sets the `X-Call-Context` header so API calls act on behalf of a managed account (identified by company UID).
- Add `AsOverride` (`*bool`) to `BgpConnectionConfig` so consumers can enable AS Override for eBGP peering. Unset leaves the API default in place.

## Changes
- Bump Go toolchain to 1.26.5 to pick up a `crypto/tls` fix for an Encrypted Client Hello privacy leak ([GO-2026-5856](https://pkg.go.dev/vuln/GO-2026-5856)).
- Bump Go toolchain to 1.26.7. This clears the four standard library advisories `govulncheck` reports against 1.26.5: [GO-2026-6218](https://pkg.go.dev/vuln/GO-2026-6218) (`net/url`), [GO-2026-6090](https://pkg.go.dev/vuln/GO-2026-6090) (`crypto/tls`), [GO-2026-5972](https://pkg.go.dev/vuln/GO-2026-5972) (`encoding/asn1`), and [GO-2026-5026](https://pkg.go.dev/vuln/GO-2026-5026) (`net/http`).
- Breaking: the `MCRLookingGlassService` route methods now call the real endpoints under `/v2/product/mcr2/{productUid}/diagnostics/routes/`. The old `lookingGlass` paths never existed, so every call returned 404.
- Breaking: `ListIPRoutes`, `ListIPRoutesWithFilter`, `ListBGPRoutes`, `ListBGPRoutesWithFilter`, and `ListBGPNeighborRoutes` now submit an asynchronous operation and poll for the result. The `List*Async`, `GetAsync*`, and `WaitForAsync*` route methods are removed.
- Breaking: `ListBGPSessions` is removed. The API has no BGP sessions endpoint.
- Breaking: `ListIPRoutesRequest.Protocol` is removed. The API has no protocol filter.
- Breaking: `ListBGPNeighborRoutesRequest.SessionID` is replaced by `PeerIPAddress`, and `IPFilter` is removed. `Direction` is now required and validated, and its type changes from `LookingGlassRouteDirection` to `string`. Use `BGPRouteDirectionReceived` (`"RECEIVED"`) or `BGPRouteDirectionAdvertised` (`"ADVERTISED"`); the old lowercase `"received"` and `"advertised"` literals still compile and now fail at runtime with `ErrMCRDiagnosticsDirectionInvalid`.
- Breaking: `LookingGlassIPRoute` and `LookingGlassBGPRoute` now match the API schemas, so every field access needs review. `NextHop` is a struct carrying the next hop IP and its VXC, and the optional `*int` counters are plain `int`. `LookingGlassIPRoute` keeps only `Prefix`, `Protocol` (now a `string`), `Metric`, `NextHop`, and the new `Distance`; every other field is gone. On `LookingGlassBGPRoute`, `ASPath` is now a `string`, and `Age`, `VXCID`, `VXCName`, `NeighborIP`, and `NeighborASN` are gone.
- Breaking: `ListBGPNeighborRoutes` now returns `[]*LookingGlassBGPRoute`. `LookingGlassBGPNeighborRoute` is removed, because both BGP endpoints return the same shape.
- Breaking: the types behind the removed methods are gone too: `RouteProtocol` and its constants, `LookingGlassRouteDirection` and its constants, `BGPSessionStatus`, `LookingGlassBGPSession`, `ListBGPSessionsRequest`, `LookingGlassAsyncStatus`, `LookingGlassAsyncJob`, `AsyncIPRoutesData`, `AsyncBGPNeighborRoutesData`, and every exported `LookingGlass*Response` envelope.
- The nil-request guards on `ListIPRoutesWithFilter`, `ListBGPRoutesWithFilter`, and `ListBGPNeighborRoutes` now return `ErrListIPRoutesRequestNil`, `ErrListBGPRoutesRequestNil`, and `ErrListBGPNeighborRoutesRequestNil`, so callers can use `errors.Is` instead of matching the message.

# 1.0.0 Release

## New Features
- All API methods now take `context`
- More configurable 
    - Custom HTTP client support
    - Structured logging is configurable and is handled using the `slog` package
- Documentation is improved
- Errors are easier to work with and are defined at the package level
- All APIs are now available in the `megaport` package rather than multiple packages in the `service` directory
- General code cleanup and linting rule enforcement 
- Missing types have been implemented

# 0.2.0 Release

## New Features
- X-Auth tokens are being deprecated from October 2023, authentication has been updated to
  support API keys in addition to username/password.

  NOTE: The username & password is being deprecated and will be removed from v0.3.0

# 0.1.16 Release

## New Features
 - Support for MCR contract term.
 - Oracle partner support and OCI VXC connectivity. Credit @aszynkow

# 0.1.15 Release

## New Features
 - Support for extracting AWS Hosted Connection ID's.

# 0.1.14 Release

## New Features
 - Support for filtering Partner Megaports by Diversity Zone.

# 0.1.13 Release

## New Features
 - Support for MCR Prefix Filter Lists.
 - Support for BGP Peer Filters and BGP Prefix Filters on VXC's.

# 0.1.12 Release

## New Features
 - Support for manually supplying peering values for Azure VXC's.

# 0.1.11 Release

## New Features
 - Add GetPorts functionality. Credit @daniel-noland

# 0.1.10 Release

## Changes
 - Move MarshallMcrAEndConfig to megaport/terraform-provider-megaport. Credit @daniel-noland

# 0.1.9 Release

## New Features
 - Add optional connection name attribute for AWS connections. Credit @ngarratt
 - Add import support for pre-existing AWS connections. Credit @ngarratt

# 0.1.8 Release

## New Features
 - PartnerConfigInterface (MCR A end configuration in VXC) handles static ip routes

# 0.1.7 Release

## New Features
 - Add NAT support for MCR VXC Connections.

# 0.1.6 Release

## New Features
 - Migrate BGP Connection support from AWS-VXC to all VXC Connections.

## Changes
 - VXC type resources now expect aEnd and bEnd configuration objects to represent these configurations.

# 0.1.5 Release

## New Features
 - Add BGP Connection support for AWS VXC Connections

## Changes
 - Rewrite and rename BuyAWSHostedVIF to BuyAWSVXC to handle new parameters and use cases

# 0.1.4 Release

## Changes
 - Fix marshalling issue with VirtualRouter in VXCResource.

# 0.1.3-beta Release

## New Features
 - Optionally Specify the Google Interconnect Location when creating a GCP Connection. Credit @kdw174

# 0.1.2-beta Release

Welcome to the first release of the `megaportgo` library!

## New Features
 - Support for the following Megaport Products has been added:
   - Ports (Single and LAG).
   - VXC.
   - MCR2.
   - AWS Hosted VIF and Hosted Connection.
   - Google Cloud Interconnect.
   - Azure ExpressRoute

 - The following lookup functionality is available:
   - Locations.
   - Partners Ports.

## Changes
 - (added in 0.1.2) Changed the `WaitForPortProvisioning` function so that it considers
   "LIVE" or "CONFIGURED" as an active status.

## Notes
This product is a `beta` release, please test all your changes in the
Megaport staging environment before running on Production. Details can
be found in the documentation. If you run into any issues, please log
an issue on GitHub issues.
