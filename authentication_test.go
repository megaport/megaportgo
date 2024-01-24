package megaport

// var accessKey string
// var secretKey string

// var megaportClient *Client

// var programLevel = new(slog.LevelVar)

// const (
// 	MEGAPORTURL = "https://api-staging.megaport.com/"
// )

// func TestMain(m *testing.M) {

// 	accessKey = os.Getenv("MEGAPORT_ACCESS_KEY")
// 	secretKey = os.Getenv("MEGAPORT_SECRET_KEY")

// 	httpClient := NewHttpClient()

// 	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
// 	programLevel.Set(slog.LevelDebug)

// 	var err error

// 	megaportClient, err = New(httpClient, SetBaseURL(MEGAPORTURL), SetLogHandler(handler))
// 	if err != nil {
// 		log.Fatalf("could not initialize megaport test client: %s", err.Error())
// 	}

// 	os.Exit(m.Run())
// }

// func TestLoginOauth(t *testing.T) {
// 	megaportClient.Logger.Debug("testing login oauth")
// 	if accessKey == "" {
// 		megaportClient.Logger.Error("MEGAPORT_ACCESS_KEY environment variable not set.")
// 		os.Exit(1)
// 	}

// 	if secretKey == "" {
// 		megaportClient.Logger.Error("MEGAPORT_SECRET_KEY environment variable not set.")
// 		os.Exit(1)
// 	}

// 	ctx := context.Background()
// 	token, loginErr := megaportClient.AuthenticationService.LoginOauth(ctx, accessKey, secretKey)
// 	if loginErr != nil {
// 		megaportClient.Logger.Error("login error", "error", loginErr.Error())
// 	}
// 	assert.NoError(t, loginErr)

// 	// Session Token is not empty
// 	assert.NotEmpty(t, token)
// 	// SessionToken is a valid guid
// 	assert.NotNil(t, shared.IsGuid(token))

// 	megaportClient.Logger.Info("", "token", token)
// 	megaportClient.SessionToken = token
// }
