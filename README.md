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
	locations, err := client.LocationService.ListLocations(context.TODO())
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

###  Getting Started

Prior to working on new code, review the [Open Issues](../issues). Check whether your issue has already been raised, and consider working on an issue with votes or clear demand.

If you don't see an open issue for your need, open one and let others know what you are working on. Avoid lengthy or complex changes that rewrite the repository or introduce breaking changes. Straightforward pull requests based on discussion or ideas and Megaport feedback are the most likely to be accepted. 

Megaport is under no obligation to accept any pull requests or to accept them in full. You are free to fork and modify the code for your own use as long is it is published under the MPL-2.0 License.

## Notes

### What's new in V1

The new V1 release of the `megaportgo` project has several changes users should be aware of:

- All API methods now take `context`
- More configurable 
    - Custom HTTP client support
    - Structured logging is configurable and is handled using the `slog` package
- Documentation is improved
- Errors are easier to work with and are defined at the package level
- All APIs are now available in the `megaport` package rather than multiple packages in the `service` directory
- General code cleanup and linting rule enforcement 
- Missing types have been implemented
