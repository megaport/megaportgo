# Overview
# MegaportGo

This is the Megaport Go Library. It allows users to orchestrate the creation of Megaport Services.

Before using this library, please ensure you read Megaport's [Terms and Conditions](https://www.megaport.com/legal/global-services-agreement/).

## API Overview
To view the Library Documentation, please see the [Wiki](../../wiki).

The [Megaport API Documentation](https://dev.megaport.com/) is also available online.

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
* MEGAPORT_USERNAME: The username used to login to the Megaport Portal.
* MEGAPORT_PASSWORD: The password used to login to the Megaport Portal.
* MEGAPORT_MFA_OTP_KEY: The key taken from the "Add Authentication" screen in the Megaport Portal (_this is __not__ your OTP, it is the key you used to setup your Authenticator_).
* MEGAPORT_ACCESS_KEY: The access key used to generate a token to authenticate API requests.
* MEGAPORT_SECRET_KEY: The secret key used to generate a token to authenticate API requests.

## Contributing
Please read the [Contributing](../../wiki/Contributing) Section prior to starting work on contributions.
