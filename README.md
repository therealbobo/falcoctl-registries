# oauth

This little project is the home of a test OAuth2.0 server for `falcoctl` and `oras-go`.


## PoC

First of all, start the OAuth2.0 server
```shell
$ go run server.go
```

The server can perform all possible OAuh2.0 grant types. Regarding the client credentials flow, 
the server can authenticate a client app having `000000` and `999999` as client ID and client secret, respectively.
Take this in mind if you want later use `falcoctl`:

```shell
$ falcoctl registry oauth  --client-id=000000 --client-secret=999999  --token-url="http://localhost:9096/token" --scopes="my-scope"
```

Then, start a fake http server, mimicking the tags endpoint for an OCI registry:

```shell
$ go run fake-registry.go
```

Lastly, start the client

```shell
$ go run main.go
```

Steps performed by the client and explanation of the output:

1. client will try to authenticate to OCI registry via OAuth2.0 registry
   - this is done by issuing a POST request at the token endpoint.
   - if clientID and client secret are registered in the OAuth server, an access token with expiration will be sent back
2. client tries to list down tags of a given repository
   - before sending the real request (`GET /v2/myrepo/tags/list`), it checks if the access token is valid (not expired)
   - if valid, request is sent directly 
   - if not, a new request to token endpoint to get a new access token using the long lived credentials (client ID and client secret), as it was done at the beginning


