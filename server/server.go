package main

import (
	"context"
	"fmt"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

func main() {
	manager := manage.NewDefaultManager()
	// token memory store
	manager.MustTokenStorage(store.NewMemoryTokenStore())

	// client memory store
	clientStore := store.NewClientStore()
	clientStore.Set("000000", &models.Client{
		ID:     "000000",
		Secret: "999999",
		Domain: "http://localhost:3000/callback",
		UserID: "loresuso",
	})
	manager.MapClientStorage(clientStore)

	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte("00000000"), jwt.SigningMethodHS512))

	// config used for client credentials
	cfg := &manage.Config{
		AccessTokenExp:    5 * time.Second,
		RefreshTokenExp:   0,
		IsGenerateRefresh: false,
	}
	manager.SetClientTokenCfg(cfg)

	// useful to test other grant types
	refreshTokenConfig := &manage.RefreshingConfig{
		AccessTokenExp:     time.Second * 3,
		RefreshTokenExp:    time.Hour * 24,
		IsGenerateRefresh:  true,
		IsResetRefreshTime: false,
		IsRemoveAccess:     false,
		IsRemoveRefreshing: false,
	}
	manager.SetRefreshTokenCfg(refreshTokenConfig)

	srv := server.NewDefaultServer(manager)

	srv.SetAllowGetAccessRequest(false) // it was true before

	srv.SetClientInfoHandler(server.ClientFormHandler)

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	srv.SetUserAuthorizationHandler(func(w http.ResponseWriter, r *http.Request) (string, error) {
		return "id", nil
	})

	srv.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (userID string, err error) {
		if clientID == "000000" && username == "loresuso" && password == "loresuso" {
			return "loresuso", nil
		}
		return "", errors.ErrAccessDenied
	})

	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(formatRequest(r))
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(formatRequest(r))
		srv.HandleTokenRequest(w, r)
	})

	http.HandleFunc("/hitme", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(formatRequest(r))
		w.Write([]byte("ok hit"))
	})

	log.Fatal(http.ListenAndServe(":9096", nil))
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		r.ParseForm()
		request = append(request, "\n")
		request = append(request, r.Form.Encode())
	}

	request = append(request, "---------------------------------")
	// Return the request as a string
	return strings.Join(request, "\n")
}
