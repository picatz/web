package web

import "net/http"

type Authenticator interface {
	Routes() Routes
	RequireAuthentication(http.HandlerFunc) http.HandlerFunc
	IsAuthenticated(w http.ResponseWriter, r *http.Request) bool
	Authenticate(w http.ResponseWriter, r *http.Request)
	Deauthenticate(w http.ResponseWriter, r *http.Request)
	ReadSessionValue(w http.ResponseWriter, r *http.Request, key string) (interface{}, bool)
}

type exampleAuthenticator struct{}

func (a *exampleAuthenticator) RequireAuthentication(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !a.IsAuthenticated(w, r) {
			http.NotFound(w, r)
			return
		}
		h(w, r)
	}
}

func (a *exampleAuthenticator) IsAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	return false
}

func (a *exampleAuthenticator) ReadSessionValue(w http.ResponseWriter, r *http.Request, key string) (interface{}, bool) {
	return nil, false
}

func (a *exampleAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) {}

func (a *exampleAuthenticator) Deauthenticate(w http.ResponseWriter, r *http.Request) {}

func (a *exampleAuthenticator) Routes() Routes {
	return Routes{
		"/auth/example/login":  a.Authenticate,
		"/auth/example/logout": a.Deauthenticate,
	}
}
