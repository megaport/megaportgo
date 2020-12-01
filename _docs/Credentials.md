The `credentials` package is used to manage credentials for the Megaport API. It strictly deals with handling the files in ~/.megaport. for information on how to authenticate with the Megaport API, see [Authentication](Authentication).

As part of the Megaport Go Library, there are two separate files that save authentication information:
1. `~/.megaport/credentials`: Contains your username, password, and one time password key (shouldn't change).
2. `~/.megaport/session_token`: Contains your session token which is used to authenticate all of requests. Should change after every call to `authentication.Login()`.

This package is responsible for managing the lifecycle of these files and `types.Credential` objects.

## New(username string, password string)
`username`: Megaport username.  
`password`: Megaport password.

#### Overview

Returns an Credential object (as defined in `types`). When created using this method, it will set `Username` and `Password` of the object to the parameter values, and set `Initialised` to `true`.

#### Returns
`*types.Credential`: Bare-minimum `*types.Credential`.

#### Example
```golang
myCredential := New("test_user", "test_password")
fmt.Printf(myCredential.Username) // "test_user"
fmt.Printf(myCredential.Password) // "test_password"
fmt.Printf("%v", myCredential.Initialised) // "true"
fmt.Printf(myCredential.SessionToken) // ""
fmt.Printf(myCredential.OneTimePasswordKey) // ""
```

## Save(credential *types.Credential, tokenOnly bool, overwrite bool)
`credential`: A credential to save to file.  
`tokenOnly`: Whether to only save the session token to file (ignores Credential information).  
`overwrite`: If the credentials file exists, and this is set to false, it will fail to save.  

#### Overview

This function is responsible for serialising a `types.Credential` object and then saving the relevant fields to these two files.

#### Returns
`error`: 
* `FileExistsNoOverwriteFlag`: The credentials file exists and `overwrite` is set to false.
* `ErrorWritingCredentialsFile`: There was an error writing to the credentials file (includes inner error text).
* `ErrorWritingSessionFile`: There was an error writing to the session token file.

#### Example
```golang
myCredential := New("test_user", "test_password")
myCredential.SessionToken = "4fae6b2b-7dba-4637-93df-8c4ab8476d0b"
Save(myCredential, false, true)
```

```shell script
$ cat ~/.megaport/credentials
username=test_user
password=test_password

$ cat ~/.megaport/session_token
4fae6b2b-7dba-4637-93df-8c4ab8476d0b
```

## Load()
#### Overview

This function loads credential files into a `types.Credential` object from each relevant file.

#### Returns
`*types.Credential`: A credentials file loaded with all available fields taken from both `credentials` and `session_token` files.  
`error`:
* `ConfigDirectoryNotExist`: `~/.megaport` doesn't exist.
* `CredentialsFileNotExist`: The credentials file doesn't exist

#### Example
```shell script
$ cat ~/.megaport/credentials
username=test_user
password=test_password

$ cat ~/.megaport/session_token
4fae6b2b-7dba-4637-93df-8c4ab8476d0b
```

```golang
myCredential := Load()
fmt.Printf(myCredential.Username) // "test_user"
fmt.Printf(myCredential.Password) // "test_password"
fmt.Printf(myCredential.SessionToken) // "4fae6b2b-7dba-4637-93df-8c4ab8476d0b"
fmt.Printf(myCredential.OneTimePasswordKey) // ""
```

## Delete()
`includeDirectory`: Whether or not `~/.megaport` should be deleted.  
`includeCredentials`: Whether the credentials file should be deleted.

#### Overview

This function deletes files under `~/.megaport` (and the directory itself, if `includeDirectory` is `true`). The only file that is always deleted is the session token file.

#### Returns
`error`
* `CredentialFileNotExist`: Returned if there is no credential file but you've set `includeCredentials` to `true`.
* `SessionTokenFileNotExist`: If the session token file doesn't exist.
* Generic errors are also returned.

#### Examples

```golang
Delete(false, true)
```

```shell script
ls ~/.megaport/credentials
ls: /Users/$USER/.megaport/credentials: No such file or directory
```

## ConfigDirectoryPath()

#### Overview

A helper function that gets around `~/` not usable as a directory location for `$HOME`. Returns the Megaport parent configuration directory.

#### Returns
* `string`: `$HOME/.megaport`.

## FilePath

#### Overview

A helper function that gets around `~/` not usable as a directory location for `$HOME`. Returns the location of the credential file.

#### Returns
* `string`: `$HOME/.megaport/credentials`.

## SessionTokenFilePath()

#### Overview

A helper function that gets around `~/` not usable as a directory location for `$HOME`. Returns the location of the session token file.

#### Returns
* `string`: `$HOME/.megaport/session_token`.