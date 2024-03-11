package megaport_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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

// Example with a custom logger
func Example_logger() {
	// A handler that logs JSON logs to stdout but only errors
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError,
	})

	// Create a new client with your custom log handler
	client, err := megaport.New(nil,
		megaport.WithCredentials("ACCESS_KEY", "SECRET_KEY"),
		megaport.WithEnvironment(megaport.EnvironmentStaging),
		megaport.WithLogHandler(handler),
	)
	if err != nil {
		// ...
	}

	client.Logger.ErrorContext(context.Background(), "testing") // will print
	client.Logger.InfoContext(context.Background(), "testing")  // won't print
}
