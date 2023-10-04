This is the Megaport Go Library. It allows users to orchestrate the creation of Megaport Services. Before using this library, please ensure you read Megaport's [Terms and Conditions](https://www.megaport.com/legal/global-services-agreement/). It is recommended that you begin testing against Megaport's Staging API, as you are liable for any charges incurred against the Production API.

## Getting Started
To begin using this library, you must first generate a config Object that contains the Megaport URL for the API, a logger, and set the Session token by authenticating with the authentication service.

``` go
import (
    "github.com/megaport/megaportgo/service/authentication"
    "github.com/megaport/megaportgo/service/port"
    "github.com/megaport/megaportgo/config"
)

func main() {
    logger := config.NewDefaultLogger()

    clientID := os.Getenv("MEGAPORT_ACCESS_KEY")
    clientSecret := os.Getenv("MEGAPORT_SECRET_KEY")

    cfg := config.Config{
        Log:      logger,
        Endpoint: "https://api-staging.megaport.com/",
    }

    auth := authentication.New(&cfg)
    token, _ := auth.LoginOauth(clientID, clientSecret)

    port := port.New(&cfg)
    port.GetPortDetails("1234")
}
```

### API URL
To set the API, you need to set the environment variable `MEGAPORT_URL`. The following environments are available for public testing:
* Staging: `https://api-staging.megaport.com/`
* Production: `https://api.megaport.com/`

## Testing

Tests can be executed for this library by running `make integration` to run all integration tests or by calling one of the following to run the tests per service:

* auth-integ
* location-integ
* mcr-integ
* partner-integ
* port-integ
* vxc-integ

In order to run theses tests valid user Credentials will need to be provided as per the Credentials section below.

### Credentials
Go to Tools > API Key Generator in Megaport Portal to manage Active API Keys.

For the purposes of testing Megaport Credentials can be passed to the integration tests by setting the following environment variables:
* MEGAPORT_ACCESS_KEY: The access key used to generate a token to authenticate API requests.
* MEGAPORT_SECRET_KEY: The secret key used to generate a token to authenticate API requests.

### Test User

A test user can be created to be used within the Megaport ecosystem by running `make create-user`. Is is recommended you do not use this functionality for anything except testing, and only in the staging environment.

## Additional API Information
The first port of call for all information regarding the API should be the go
docs. The below articles are subjects that require additional information that
is not documented in the API. If you would like something documented, please lodge
a GitHub issue.

* [Contributing](Contributing)
