# Megaport Go

[![Go Reference](https://pkg.go.dev/badge/github.com/megaport/megaportgo.svg)](https://pkg.go.dev/github.com/megaport/megaportgo)

## Overview

This is the Megaport Go Library. It allows users to orchestrate the creation of Megaport Services.

Before using this library, please ensure you read Megaport's [Terms and Conditions](https://www.megaport.com/legal/global-services-agreement/).

The [Megaport API Documentation](https://dev.megaport.com/) is also available online.

## Getting started

```go
package main

import (
	"context"
	"fmt"

	megaport "github.com/megaport/megaportgo"
)

// Create a new client using your credentials and interact with the Megaport API
func main() {
	// Creates a new client using the default HTTP cleint with the specified credentials against the staging environment
	client, err := megaport.New(nil,
		megaport.WithCredentials("ACCESS_KEY", "SECRET_KEY"),
		megaport.WithEnvironment(megaport.EnvironmentStaging),
	)
	if err != nil {
		// ...
	}

	// Authorize the client using the client's credentials
	authInfo, err := client.Authorize(context.TODO())
	if err != nil {
		// ...
	}
	fmt.Println(authInfo.AccessToken)
	fmt.Println(authInfo.Expiration) // You can use the expiration here to reauthorize the client when your access token expires

	// After you have authorized you can interact with the API
	locations, err := client.LocationService.ListLocationsV3(context.TODO())
	if err != nil {
		// ...
	}

	for _, location := range locations {
		fmt.Println(location.Name)
	}
}
```

## Testing

For mock tests `go test ./...`

To run integration tests against the Megaport API you will need to [generate an API key](https://docs.megaport.com/api/api-key/)

```
export MEGAPORT_ACCESS_KEY=YOUR_KEY
export MEGAPORT_SECRET_KEY=YOUR_KEY

go test -timeout 20m -integration ./...
```

## Location API Migration (v2 → v3)

### ⚠️ BREAKING CHANGE: Location API v2 Deprecated

The Megaport Location API v2 has been deprecated and **will no longer work**. All existing code using v2 location methods must be migrated to v3. This is a significant change that affects most applications using the Megaport Go library.

### Quick Migration Guide

**❌ Old v2 methods (deprecated):**

```go
// These methods no longer work and will return errors
locations, err := client.LocationService.ListLocations(ctx)
location, err := client.LocationService.GetLocationByID(ctx, 123)
location, err := client.LocationService.GetLocationByName(ctx, "Equinix SY3")
matches, err := client.LocationService.GetLocationByNameFuzzy(ctx, "Sydney")
filtered, err := client.LocationService.FilterLocationsByMarketCode(ctx, "AU", locations)
mcrLocations := client.LocationService.FilterLocationsByMcrAvailability(ctx, true, locations)
```

**✅ New v3 methods (required):**

```go
// Use these v3 methods instead
locations, err := client.LocationService.ListLocationsV3(ctx)
location, err := client.LocationService.GetLocationByIDV3(ctx, 123)
location, err := client.LocationService.GetLocationByNameV3(ctx, "Equinix SY3")
matches, err := client.LocationService.GetLocationByNameFuzzyV3(ctx, "Sydney")
filtered, err := client.LocationService.FilterLocationsByMarketCodeV3(ctx, "AU", locations)
mcrLocations := client.LocationService.FilterLocationsByMcrAvailabilityV3(ctx, true, locations)
```

### Key Changes in v3

#### 1. Data Structure Changes

The `LocationV3` struct has significant differences from the legacy `Location` struct:

**Enhanced Address Structure:**

```go
// v2: map[string]string
address := location.Address["city"] // v2 way

// v3: structured data
city := location.Address.City       // v3 way
state := location.Address.State
country := location.Address.Country
street := location.Address.Street
```

**New Data Center Information:**

```go
// v3 only - not available in v2
dataCenterName := location.GetDataCenterName()
dataCenterID := location.GetDataCenterID()
```

**Product Availability Restructured:**

```go
// v2: Simple boolean and arrays
hasMCR := location.Products.MCR
speeds := location.Products.MCR2

// v3: Diversity zones with detailed product info
hasMCR := location.HasMCRSupport()
speeds := location.GetMCRSpeeds()
hasMVE := location.HasMVESupport()
maxCores := location.GetMVEMaxCpuCores()
```

#### 2. Removed Fields

The following fields from v2 are **no longer available** in v3:

- `NetworkRegion` - Not provided by v3 API
- `SiteCode` - Not provided by v3 API
- `Campus` - Not provided by v3 API
- `LiveDate` - Not provided by v3 API
- `VRouterAvailable` - Not provided by v3 API
- `Products.MCRVersion` - Restructured in v3
- `Products.MVE` - Completely restructured in v3

#### 3. New v3 Features

The v3 API provides enhanced functionality not available in v2:

**Diversity Zones:**

```go
// Check if location has red/blue diversity zones
if location.DiversityZones != nil {
    if location.DiversityZones.Red != nil {
        redMCRSpeeds := location.DiversityZones.Red.McrSpeedMbps
    }
    if location.DiversityZones.Blue != nil {
        blueMegaportSpeeds := location.DiversityZones.Blue.MegaportSpeedMbps
    }
}
```

**Cross-Connect Support:**

```go
// New in v3
hasCrossConnect := location.HasCrossConnectSupport()
crossConnectType := location.GetCrossConnectType()
```

**Enhanced MVE Information:**

```go
// More detailed MVE support information
if location.HasMVESupport() {
    maxCores := location.GetMVEMaxCpuCores()
    // Check specific diversity zones for MVE availability
    redMVE := location.DiversityZones.Red.MveAvailable
    blueMVE := location.DiversityZones.Blue.MveAvailable
}
```

### Complete Migration Example

**Before (v2 - broken):**

```go
func getLocationInfo(client *megaport.Client) {
    locations, err := client.LocationService.ListLocations(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    for _, loc := range locations {
        fmt.Printf("Location: %s\n", loc.Name)
        fmt.Printf("Country: %s\n", loc.Country)
        fmt.Printf("City: %s\n", loc.Address["city"])
        fmt.Printf("MCR Available: %t\n", loc.Products.MCR)
        fmt.Printf("Site Code: %s\n", loc.SiteCode) // Not available in v3
    }
}
```

**After (v3 - working):**

```go
func getLocationInfo(client *megaport.Client) {
    locations, err := client.LocationService.ListLocationsV3(context.TODO())
    if err != nil {
        log.Fatal(err)
    }

    for _, loc := range locations {
        fmt.Printf("Location: %s\n", loc.Name)
        fmt.Printf("Country: %s\n", loc.GetCountry())
        fmt.Printf("City: %s\n", loc.Address.City)
        fmt.Printf("Data Center: %s\n", loc.GetDataCenterName())
        fmt.Printf("MCR Available: %t\n", loc.HasMCRSupport())
        fmt.Printf("MCR Speeds: %v\n", loc.GetMCRSpeeds())
        fmt.Printf("MVE Available: %t\n", loc.HasMVESupport())
        fmt.Printf("Cross Connect: %t\n", loc.HasCrossConnectSupport())
    }
}
```

### Backward Compatibility Helper

For legacy code that cannot be immediately migrated, you can convert v3 locations to v2 format (with limitations):

```go
// Convert v3 location to legacy format (some data will be lost)
v3Location, err := client.LocationService.GetLocationByIDV3(ctx, 123)
if err != nil {
    return err
}
legacyLocation := v3Location.ToLegacyLocation()
// Note: Some fields will be empty/nil as they're not available in v3
```

⚠️ **Warning:** The `ToLegacyLocation()` method should only be used as a temporary migration aid. Plan to update your code to use v3 data structures directly.

### Migration Checklist

- [ ] Replace all `ListLocations()` calls with `ListLocationsV3()`
- [ ] Replace all `GetLocationByID()` calls with `GetLocationByIDV3()`
- [ ] Replace all `GetLocationByName()` calls with `GetLocationByNameV3()`
- [ ] Replace all `GetLocationByNameFuzzy()` calls with `GetLocationByNameFuzzyV3()`
- [ ] Replace all `FilterLocationsByMarketCode()` calls with `FilterLocationsByMarketCodeV3()`
- [ ] Replace all `FilterLocationsByMcrAvailability()` calls with `FilterLocationsByMcrAvailabilityV3()`
- [ ] Update address access from `location.Address["key"]` to `location.Address.Key`
- [ ] Replace `location.Products.MCR` with `location.HasMCRSupport()`
- [ ] Replace direct product speed access with helper methods like `location.GetMCRSpeeds()`
- [ ] Remove code that depends on removed fields (`SiteCode`, `NetworkRegion`, etc.)
- [ ] Test thoroughly as data structures and available information have changed significantly

## Contributing

Contributions via pull request are welcome. Familiarize yourself with these guidelines to increase the likelihood of your pull request being accepted.

All contributions are subject to the [Megaport Contributor Licence Agreement](CLA.md).
The CLA clarifies the terms of the [Mozilla Public Licence 2.0](LICENSE) used to Open Source this respository and ensures that contributors are explictly informed of the conditions. Megaport requires all contributors to accept these terms to ensure that the Megaport Terraform Provider remains available and licensed for the community.

The main themes of the [Megaport Contributor Licence Agreement](CLA.md) cover the following conditions:

- Clarifying the Terms of the [Mozilla Public Licence 2.0](LICENSE), used to Open Source this project.
- As a contributor, you have permission to agree to the License terms.
- As a contributor, you are not obligated to provide support or warranty for your contributions.
- Copyright is assigned to Megaport to use as Megaport determines, including within commercial products.
- Grant of Patent Licence to Megaport for any contributions containing patented or future patented works.

The [Megaport Contributor Licence Agreement](CLA.md) is
the authoritative document over these conditions and any other communications unless explicitly stated otherwise.

When you open a Pull Request, all authors of the contributions are required to comment on the Pull Request confirming
acceptance of the CLA terms. Pull Requests can not be merged until this is complete.

The [Megaport Contributor Licence Agreement](CLA.md) applies to contributions.
All users are free to use the `megaportgo` project under the [MPL-2.0 Open Source Licence](LICENSE).

Megaport users are also bound by the [Acceptable Use Policy](https://www.megaport.com/legal/acceptable-use-policy).

### Getting Started

Prior to working on new code, review the [Open Issues](../issues). Check whether your issue has already been raised, and consider working on an issue with votes or clear demand.

If you don't see an open issue for your need, open one and let others know what you are working on. Avoid lengthy or complex changes that rewrite the repository or introduce breaking changes. Straightforward pull requests based on discussion or ideas and Megaport feedback are the most likely to be accepted.

Megaport is under no obligation to accept any pull requests or to accept them in full. You are free to fork and modify the code for your own use as long is it is published under the MPL-2.0 License.

## Notes

### What's new in V1

The new V1 release of the `megaportgo` project has several changes users should be aware of:

- **🚨 BREAKING: Location API v3 Migration** - The v2 locations API is deprecated and no longer works. All location-related code must be migrated to v3 methods. See the [Location API Migration](#location-api-migration-v2--v3) section above for detailed migration instructions.
- All API methods now take `context`
- More configurable
  - Custom HTTP client support
  - Structured logging is configurable and is handled using the `slog` package
- Documentation is improved
- Errors are easier to work with and are defined at the package level
- All APIs are now available in the `megaport` package rather than multiple packages in the `service` directory
- General code cleanup and linting rule enforcement
- Missing types have been implemented
