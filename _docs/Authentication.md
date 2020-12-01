The `authentication` package is used to log into and out of Megaport's API. It can authenticate with a user and password (one time password optional) and will return a session token. These details are kept in the files as outlined in [Credentials](Credentials).

The four values maanged by this package are:
1. `username`: your Megaport Username.
2. `password`: your Megaport Password.
3. `one_time_password_key`: This is the key taken from the Add Authenticator screen from the Megaport Portal. It is the key used to generate a OTP, NOT a OTP itself.
2. `session_token`: The session token is used to authenticate all requests (except `Login`) using the Megaport API. It is passed into each request as the Header `X-Auth-Token`. This value is loaded into the environment variable `MEGAPORT_SESSION_TOKEN`.

This package also contains utility functions used for testing the library, including creating new users, however this functionality should not be used except in testing circumstances.

## CreateUser(firstName string, lastName string, credentials *types.Credential)
`firstName`: User's first name.
`lastName`: User's last name.
`credentials`: a credential's object that provides the credentials for the created user.

#### Overview
This function creates a user for use within the Megaport Ecosystem. Is is recommended you do not use this function for anything except testing, and only in the staging environment.

#### Returns
`error`

#### Example
```go
credentials := types.Credential{}
CreateUser("Test", "User", &credentials)
```

## Login(forceNew bool)
* `forceNew`: If this is set to true, a session token will be retrieved from the Megaport API. If it is not true, `Login` will attempt to restore the session token saved in the session token file.

#### Overview
Login is a wrapper around `GetSessionToken` which loads a username and password from file first before initiating the retrieval of a Session Token.

#### Returns
* `*types.Credential`: The credential with details loaded from file, as well as a retrieved Session Token for the Megaport API.
* `error`

#### Example

```go
credentials := Login(true)
fmt.Printf(credentials.SessionToken) // "4fae6b2b-7dba-4637-93df-8c4ab8476d0b"
```

## Logout()

#### Overview
Deletes the stored Session Token and clears the `MEGAPORT_SESSION_TOKEN` environment variable.

#### Example

```go
fmt.Printf(os.Getenv("MEGAPORT_SESSION_TOKEN")) // "4fae6b2b-7dba-4637-93df-8c4ab8476d0b"
Logout()
fmt.Printf(os.Getenv("MEGAPORT_SESSION_TOKEN")) // ""
```

## GetSessionToken(credentials *types.Credential) 
* `credentials`: The credentials you want to retrieve a session token for.

#### Overview
GetSessionToken connects to the Megaport API, passes the saved username, password, and a generated OTP (if otp key is set). The Megaport API returns a session token which is then saved in the session token file. The Session Token is also stored in the environment variable `MEGAPORT_SESSION_TOKEN`.

#### Returns
* `error`

#### Example
```go
credentials := Load()
GetSessionToken(credentials)
fmt.Printf(credentials.SessionToken) // "4fae6b2b-7dba-4637-93df-8c4ab8476d0b"
```

## GenerateOneTimePassword(credentials *types.Credential)
* `credentials`: The credentials to use to generate a one_time_password.

#### Overview
Generates a OTP using a Google Authenticator-compatible OTP Key. The field `one_time_password_key` must be set in your Megaport credentials file.

#### Returns
* `string`: A one time password to be used in a Megaport API login call.
* `error`:
    * `ERR_NO_OTP_KEY_DEFINED`


#### Example
```go
credentials := Load()
fmt.Printf(GenerateOneTimePassword(credentials)) // "123456"
```
