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

