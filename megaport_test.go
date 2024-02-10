package megaport

import (
	"flag"
	"log/slog"
)

var runIntegrationTests = flag.Bool("integration", false, "perform integration tests")

var accessKey string
var secretKey string

var megaportClient *Client

var programLevel = new(slog.LevelVar)

const (
	MEGAPORTURL = "https://api-staging.megaport.com/"
)
