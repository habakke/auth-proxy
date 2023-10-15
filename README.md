# Authentication proxy

Authentication proxy used to add authentication to applications which does not
natively support it.

Supported authentication providers:
* Google Oauth2

### Prerequisites

This project has the following prerequisites which must be installed locally to build and run tests
* golangci-lint

### Build
To build from source run

```shell
make build
```

### Test
Running the unit and module tests can be done by running

```shell
make test
```

### Release
Before releasing make sure the correct version is listed in the `VERSION` file, and a valid GITHUB_TOKEN is exported
before running the release command.

```shell
make release
```

### Install

The latest docker image can be pulled from ghcr.io

```shell
docker pull ghcr.io/habakke/auth-proxy:latest
```

## Configuration

The proxy is fully controlled through environmental variables. The table below lists the 
available environmental variables and their default value.

| Variable | Default | Description |
| -------- | ------- | ----------- |
| PORT | 8080 | The port number which the service listens on |
| TARGET | - | The URL where the auth-proxy should forward requests after authenticating |
| TOKEN | - |Bearer token to append to all requests towards the TARGET |
| COOKIE_SEED | - | Seed used to introduce entropy in the cookie signatures |
| COOKIE_KEY | - | Key used to encrypt cookie payload |
| LOGLEVEL | info | Default log level set to any of `error, warn, info, debug, trace`. If this parameter is not set, it defaults to `info` |
| PROFILE | - | Set this variable to enable profiling of the golang application |
| GOOGLE_OAUTH_CLIENT_ID | - | Google Oauth2 Client ID |
| GOOGLE_OAUTH_CLIENT_SECRET | - | Google Oauth2 Client Secret |
| GOOGLE_OAUTH_CALLBACK_URL | - | Google Oauth2 callback url, ex. https://example.com/auth/google/callback |
| HOMEPAGE_URL | - | Homepage URL which is inserted into templates |
| CONTACT_EMAIL | - | Contact email which is inserted into templates |

### Provider configuration

#### Config Google Project

First things first, we need to create a Google Project and create OAuth2 credentials.

* Go to Google Cloud Platform
* Create a new project or select one if you already have it.
* Go to Credentials and then create a new one choosing “OAuth client ID”
* Add "authorized redirect URL", for this example localhost:8000/auth/google/callback
* Copy the client_id and client secret

## TODO

Add support for additional authentication providers
* Bluebit Ninja
