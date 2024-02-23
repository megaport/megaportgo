package megaport

import (
	"flag"
	"log/slog"
)

// runIntegrationTests is a flag to determine if integration tests should be run
var runIntegrationTests = flag.Bool("integration", false, "perform integration tests")

var accessKey string
var secretKey string

// megaportClient is the client for the test suite
var megaportClient *Client

// programLevel is the log level for the test suite
var programLevel = new(slog.LevelVar)

// Default Base URL for Integration Tests
const (
	MEGAPORTURL = "https://api-staging.megaport.com/"
)
