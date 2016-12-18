package janus

import (
	"net/http"

	"fmt"

	"encoding/base64"

	log "github.com/Sirupsen/logrus"
	"github.com/hellofresh/janus/errors"
)

// Oauth2SecretMiddleware prevents requests to an API from exceeding a specified rate limit.
type Oauth2SecretMiddleware struct {
	oauthSpec *OAuthSpec
}

func NewOauth2SecretMiddleware(oauthSpec *OAuthSpec) *Oauth2SecretMiddleware {
	return &Oauth2SecretMiddleware{oauthSpec}
}

// Handler is the middleware method.
func (m *Oauth2SecretMiddleware) Handler(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Starting Oauth2Secret middleware")

		if "" != r.Header.Get("Authorization") {
			log.Debug("Authorization is set, proxying")
			handler.ServeHTTP(w, r)
			return
		}

		clientID := r.URL.Query().Get("client_id")
		if "" == clientID {
			log.Debug("ClientID not set, proxying")
			handler.ServeHTTP(w, r)
			return
		}

		clientSecret, exists := m.oauthSpec.Secrets[clientID]
		if false == exists {
			panic(errors.ErrClientIdNotFound)
		}

		m.ChangeRequest(r, clientID, clientSecret)
		handler.ServeHTTP(w, r)
	})
}

// ChangeRequest modifies the request to add the Authorization headers.
func (m *Oauth2SecretMiddleware) ChangeRequest(req *http.Request, clientID, clientSecret string) {
	log.Debug("Modifying request")
	authString := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", clientID, clientSecret)))
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", authString))
}