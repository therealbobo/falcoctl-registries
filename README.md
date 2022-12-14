# oauth

This little project is the home of a test OAuth2.0 server for `falcoctl` and `oras-go`.
PoC v2 is the latest version that uses JWT and also implements a simple rate limiting logic.

## PoC v1 (old)

<img src="oauth-flow.png"/>

First of all, build the container in the nginx folder and start it

```shell
$ cd nginx && docker build -t loresuso/oauth-proxy .
$ docker run --rm -it --net=host loresuso/oauth-proxy
```

Then, start the OAuth2.0 server
```shell
$ go run server.go
```

The server can perform all possible OAuh2.0 grant types. Regarding the client credentials flow, 
the server can authenticate a client app having `000000` and `999999` as client ID and client secret, respectively.
Take this in mind if you want later use `falcoctl`:

```shell
$ falcoctl registry oauth  --client-id=000000 --client-secret=999999  --token-url="http://localhost:9096/token" --scopes="my-scope"
```

The above command save client credentials in the filesystem, so that can be used later on.

Then, start an OCI registry

```shell
$ docker run -it --rm -p 5000:5000 --name registry registry:2
```

Lastly, run falcoctl (remember to use oauth and plain http for testing)

```shell
$ falcoctl registry push ...
```

## PoC v2 

PoC v1's main drawback is that every request made to the proxy corresponds to another request done against the OAuth server for token introspection. To avoid this, we can make use of signed JWTs. Signed JWTs allows the proxy to verify that authenticity and the integrity of a JWT token. For the sake of simplicity, this PoC uses HMAC as signing algorithm, and verification happens by using a shared common secret between proxy and Oauth server. Any other (and more robust) algorithm can be used for production use cases. 

First of all, let `falcoctl` store client credentials, so that it can be able to make authenticated requests later:

```shell
$ ./falcoctl registry oauth --client-id=000000 --client-secret=999999 --token-url "http://localhost:9096/token"
```
This will validate the client credentials and, if so, they will be stored in a file for later use. 

Then, start a Redis server:
```shell
$ docker run --rm --name my-redis-container -p 6379:6379 -d redis
```
This is used to implement a very simple rate limiting algorithm by making use of `INCR` and `EXPIRE`. 
Keys are composed by `clientID | currentMinute`. We keep increasing a counter everytime a client hit our proxy with a request in a given minute. If the counter goes beyond a threshold, we do not pass the request to the registry. We also set the `EXPIRE` everytime this key is hit, and this is set to 59 seconds. This way, when we will roll from minute 59 to 00, we are sure that the key for that minute was expired and we can start increasing the counter for the first minute of the new hour.

Then, start also a container registry:
Then, start a Redis server:
```shell
$ docker run --rm -it --name=registry -p 5000:5000 registry
```
Keep it in another terminal so that you can see what kind of operations are performed on it, for debugging purposes. 

Let's launch also the OAuth server:
```shell
$ go run server.go
```

The last piece needed is a reverse proxy server written in Go, that can perform token validation and rate limiting:
```shell
$ go run proxy.go
```

Now you can see that you can be able to push and pull artifacts using `falcoctl`:
```shell
$ ./falcoctl registry push localhost:6000/test:7.0.0 --type plugin --platform linux/x86 /tmp/triage.scap --oauth --plain-http
$ ./falcoctl registry pull localhost:6000/test:7.0.0 --oauth --plain-http --platform linux/x86
```

Rate limit is set to 15 requests per minute.