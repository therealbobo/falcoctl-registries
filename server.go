package main

import (
	"context"
	"fmt"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	oauthserver "oauth/server"
	"time"

	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
)

func main() {
	ctx := context.Background()

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
		fmt.Println(oauthserver.FormatRequest(r))
		err := srv.HandleAuthorizeRequest(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	})

	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(oauthserver.FormatRequest(r))
		srv.HandleTokenRequest(w, r)
	})

	http.HandleFunc("/hitme", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(oauthserver.FormatRequest(r))
		w.Write([]byte("ok hit"))
	})

	// Token introspection endpoint
	http.HandleFunc("/introspect", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		accessToken := r.FormValue("token")
		ti, err := srv.Manager.LoadAccessToken(ctx, accessToken)

		if duration := ti.GetAccessExpiresIn(); duration <= 0 {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
	})

	log.Fatal(http.ListenAndServe(":9096", nil))
}
