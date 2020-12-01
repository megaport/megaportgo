This is the Megaport Go Library. It allows users to orchestrate the creation of Megaport Services. Before using this library, please ensure you read Megaport's [Terms and Conditions](https://www.megaport.com/legal/global-services-agreement/). It is recommended that you begin testing against Megaport's Staging API, as you are liable for any charges incurred against the Production API.

## Getting Started
To begin using this library, you must set the URL the API will use and the credentials to access the Megaport API.
### API URL
To set the API, you need to set the environment variable `MEGAPORT_URL`. The following environments are available for public testing:
* Staging: `https://api-staging.megaport.com/`
* Production: `https://api.megaport.com/`
```bash
export MEGAPORT_URL=https://api-staging.megaport.com/
```
### Credentials
Megaport Credentials utilised by this tool should be stored in `~/.megaport/credentials`. There are three values you can set in the credentials file, all other values will be ignored. The values are:
* __username__: The username you use to login to the Megaport Portal.
* __password__: The password you use to login to the Megaport Portal.
* __one_time_password_key__: The key taken from the "Add Authentication" screen in the Megaport Portal (_this is __not__ your OTP, it is the key you used to setup your Authenticator_).

To test your credentials, you can run `go test -v ./authentication`. If the tests pass, your credentials have been setup correctly. 

Do __NOT__ run the credential test suite or you will lose your credential settings. It is for CI/CD usage only.

## Additional API Information
The first port of call for all information regarding the API should be the go
docs. The below articles are subjects that require additional information that
is not documented in the API. If you would like something documented, please lodge
a GitHub issue.
* [Authentication](Authentication)
* [Credentials](Credentials)
* [Contributing](Contributing)
