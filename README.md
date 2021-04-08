# Authentication proxy

Authentication proxy used to add SSO authentication support to applications which does not
natively support this.

## Config Google Project

First things first, we need to create our Google Project and create OAuth2 credentials.

* Go to Google Cloud Platform
* Create a new project or select one if you already have it.
* Go to Credentials and then create a new one choosing “OAuth client ID”
* Add "authorized redirect URL", for this example localhost:8000/auth/google/callback
* Copy the client_id and client secret

## Environmental variables

The proxy is fully controlled through environmental variables. The table below lists the 
available environmental variables and their default value.

| Variable | Default | Description |
| -------- | ------- | ----------- |
| PORT | 8080 | The port number which the service listens on |
| TARGET | - | The URL where the auth-proxy should forward requests after authenticating |
| TOKEN | - |Bearer token to append to all requests towards the TARGET |
| LOGLEVEL | info | Default log level set to any of `error, warn, info, debug, trace`. If this parameter is not set, it defaults to `info` |
| PROFILE | - | Set this variable to enable profiling of the golang application |
| GOOGLE_OAUTH_CLIENT_ID | - | Google Oauth2 Client ID |
| GOOGLE_OAUTH_CLIENT_SECRET | - | Google Oauth2 Client Secret |
| GOOGLE_OAUTH_CALLBACK_URL | - | Google Oauth2 callback url, ex. https://<domain>/auth/google/callback |

## Todo

* Implement basic instrumentation, publishing basic HTTP stats to prometheus
