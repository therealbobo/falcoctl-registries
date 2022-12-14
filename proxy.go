package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt"
)

var HmacSampleSecret = []byte("secret")

const (
	registry = "http://localhost:5000"
)

// NewProxy takes target host and creates a reverse proxy
func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(url), nil
}

// ProxyRequestHandler handles the http request using proxy
func ProxyRequestHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")
		if err := validateToken(token); err != nil {
			log.Fatal("validation error: %w", err)
			return
		}

		proxy.ServeHTTP(w, r)
	}
}

func validateToken(tokenString string) error {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return HmacSampleSecret, nil
	})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("cannot cast to map claims")
	}

	clientID, ok := claims["aud"]
	if !ok {
		return fmt.Errorf("cannot extract audience field")
	}

	expire, ok := claims["exp"]
	if !ok {
		return fmt.Errorf("cannot extract expire field")
	}

	fmt.Println(clientID, expire)

	return nil
}

func main() {
	// initialize a reverse proxy and pass the actual backend server url here
	proxy, err := NewProxy(registry)
	if err != nil {
		panic(err)
	}

	// handle all requests to your server using the proxy
	http.HandleFunc("/", ProxyRequestHandler(proxy))
	log.Fatal(http.ListenAndServe(":6000", nil))
}
