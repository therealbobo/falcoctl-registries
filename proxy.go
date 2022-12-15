package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/golang-jwt/jwt"
)

var (
	HmacSampleSecret     = []byte("secret")
	rdb                  *redis.Client
	tooManyRequestsError = errors.New("too many requests")
)

const (
	registry        = "http://localhost:5000"
	maxRequests int = 15
	expiryTime      = 10 * time.Second
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
		err := validateToken(token)
		if errors.Is(err, tooManyRequestsError) {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		} else {
			proxy.ServeHTTP(w, r)
		}
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

	// TODO: check expiry time

	// Rate limiting
	_, minute, _ := time.Now().Clock()
	key := clientID.(string) + strconv.FormatInt(int64(minute), 10)
	count, err := rdb.Get(context.Background(), key).Result()
	i, _ := strconv.Atoi(count)
	if err == redis.Nil || i < maxRequests {
		pipe := rdb.Pipeline()

		pipe.Incr(context.Background(), key)
		pipe.Expire(context.Background(), key, 59*time.Second)

		_, err := pipe.Exec(context.Background())
		if err != nil {
			panic(err)
		}
	} else if err != nil {
		return err
	} else {
		return tooManyRequestsError
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

	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalln("cannot connect to Redis for ratelimiting requests")
		return
	}

	// handle all requests to your server using the proxy
	http.HandleFunc("/", ProxyRequestHandler(proxy))
	log.Fatal(http.ListenAndServe(":6000", nil))
}
