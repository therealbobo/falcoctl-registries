package main

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"log"
)

func main() {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     "d008217df5b47b564888",
		ClientSecret: "fbc82d6ddfb60be86af679a666c53099b910d153",
		Scopes:       []string{"write:packages", "read:packages"},
		Endpoint:     github.Endpoint,
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", tok)

	client := conf.Client(ctx, tok)
	client.Get("...")
}
