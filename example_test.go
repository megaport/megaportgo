package megaport_test

import (
	"context"
	"fmt"

	megaport "github.com/megaport/megaportgo"
)

// Create a new client using your credentials and interact with the Megaport API
func Example() {
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
	fmt.Println(authInfo.Expiration) // You can use the expriation here to reauthorize the client when your access token expires

	// After you have authorized you can interact with the API
	locations, err := client.LocationService.ListLocations(context.TODO())
	if err != nil {
		// ...
	}

	for _, location := range locations {
		fmt.Println(location.Name)
	}
}
