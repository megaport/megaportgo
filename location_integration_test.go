package megaport

// func TestBadID(t *testing.T) {
// 	assert := assert.New(t)
// 	ctx := context.Background()
// 	// Make sure that an id with no record returns an error as expected.
// 	_, idErr := megaportClient.LocationService.GetLocationByID(ctx, -999999)
// 	assert.Equal(mega_err.ERR_LOCATION_NOT_FOUND, idErr.Error())
// }

// func TestBadName(t *testing.T) {
// 	assert := assert.New(t)

// 	ctx := context.Background()

// 	// Make sure that a name with no record returns an error as expected.
// 	_, nameErr := megaportClient.LocationService.GetLocationByName(ctx, "DefinitelyNotARealName")
// 	assert.Equal(mega_err.ERR_LOCATION_NOT_FOUND, nameErr.Error())
// }

// func TestGetLocationByID(t *testing.T) {
// 	assert := assert.New(t)
// 	ctx := context.Background()

// 	// Make sure our location by id works as expected
// 	byId, idErr := megaportClient.LocationService.GetLocationByID(ctx, 137)
// 	assert.Nil(idErr)
// 	assert.Equal("3DC/Telecity Sofia", byId.Name)

// 	byId, idErr = megaportClient.LocationService.GetLocationByID(ctx, 383)
// 	assert.Nil(idErr)
// 	assert.Equal("NextDC B2", byId.Name)
// }

// func TestGetLocationByName(t *testing.T) {
// 	assert := assert.New(t)
// 	ctx := context.Background()

// 	// Make sure our location by name works as expected
// 	byName, nameErr := megaportClient.LocationService.GetLocationByName(ctx, "3DC/Telecity Sofia")
// 	assert.Nil(nameErr)
// 	assert.Equal(137, byName.ID)

// 	byName, nameErr = megaportClient.LocationService.GetLocationByName(ctx, "NextDC B2")
// 	assert.Nil(nameErr)
// 	assert.Equal(383, byName.ID)

// 	byName, nameErr = megaportClient.LocationService.GetLocationByName(ctx, "Equinix SY3")
// 	assert.Nil(nameErr)
// 	assert.Equal(6, byName.ID)
// }

// func TestGetLocationByNameFuzzy(t *testing.T) {
// 	assert := assert.New(t)
// 	ctx := context.Background()

// 	byFuzzy, fuzzyErr := megaportClient.LocationService.GetLocationByNameFuzzy(ctx, "NextDC")
// 	assert.True(len(byFuzzy) > 0)
// 	assert.NoError(fuzzyErr)

// 	failFuzzy, failFuzzyErr := megaportClient.LocationService.GetLocationByNameFuzzy(ctx, "definitely not a location name at all")
// 	assert.True(len(failFuzzy) == 0)
// 	assert.Error(errors.New(mega_err.ERR_NO_MATCHING_LOCATIONS), failFuzzyErr)
// }

// // first one should always be Australia
// func TestListCountries(t *testing.T) {
// 	assert := assert.New(t)
// 	ctx := context.Background()

// 	countries, countriesErr := megaportClient.LocationService.ListCountries(ctx)
// 	assert.NoError(countriesErr)
// 	assert.Equal("Australia", countries[0].Name)
// 	assert.Equal("AUS", countries[0].Code)
// 	assert.Equal("AU", countries[0].Prefix)
// 	assert.Greater(countries[0].SiteCount, 0)
// }

// func TestMarketCodes(t *testing.T) {
// 	assert := assert.New(t)
// 	ctx := context.Background()

// 	marketCodes, _ := megaportClient.LocationService.ListMarketCodes(ctx)
// 	found := false

// 	for i := 0; i < len(marketCodes); i++ {
// 		if marketCodes[i] == "AU" {
// 			found = true
// 		}
// 	}

// 	assert.True(found)
// }

// func TestFilterLocationsByMarketCode(t *testing.T) {
// 	assert := assert.New(t)
// 	ctx := context.Background()

// 	locations, err := megaportClient.LocationService.ListLocations(ctx)
// 	assert.NoError(err)
// 	currentCount := len(locations)
// 	filterErr := megaportClient.LocationService.FilterLocationsByMarketCode(ctx, "AU", &locations)
// 	assert.NoError(filterErr)

// 	assert.Less(len(locations), currentCount)
// 	assert.Equal("AU", locations[0].Market)
// }
