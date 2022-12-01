As you can see, when using client credentials, no `refresh token` will be issued. The access token returned is in JWT format,
but not customized yet.

Notice also that to retrieve the access token the client try to autodetect where the client ID and client secret have to be put,
whether in an `Authorization` header or url encoded in the body of the request.

Other considerations:
- server is issuing access tokens with 5 seconds expiration
- client (main.go in the simple example), is hitting the `hit-me` endpoint every three seconds
  You can see that from time to time the client requests a new token using the `/token` endpoint

### Problems

- `refresh tokens` are meaningless when using client credentials. This is because, since it is the app that is authenticating, it can use its own client credentials to authenticate and get new access tokens. Look at this flow as password grant but for the app itself, more or less.
- this makes a lot of sense: `refresh tokens` were only born to let third party apps access data of a user without letting him reauthenticate and stop. Here, we have no user interaction, never.
- using multiple client credentials for the same app but different deployments or whatever, for me sounds like we are not using client credential flow like intended.
    - client credentials cannot be per deployment, they are intended to be per app.
    - and since we have only one app, namely `falcoctl`, there should be only one client credentials around per service/registry used.
    - we can use client credentials for a workaround, but we have to be aware about that.
- rules at the end of the day have to be put somewhere. Either GCR/GAR, DockerHub, ECR, GHCR, none of these allow using client credentials flow, based on my research. So we can do all the PoCs we want but they will be meaningless. 
