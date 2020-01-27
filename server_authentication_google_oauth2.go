package web

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Google Oauth2 middleware
//
// References:
// - https://support.google.com/cloud/answer/6158849?hl=en

type googleAuthData struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	//GivenName     string `json:"given_name"`
	//FamilyName    string `json:"family_name"`
	//Link          string `json:"link"`
	//Picture       string `json:"picture"`
	//Locale        string `json:"locale"`
	//Hd            string `json:"hd"`
}

var googleOauthConfig = &oauth2.Config{
	RedirectURL:  "http://localhost:8080/auth/google/callback",
	ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	Scopes:       []string{"email", "profile"},
	Endpoint:     google.Endpoint,
}

const oauthGoogleURLAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(365 * 24 * time.Hour)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}

func getUserDataFromGoogle(code string) ([]byte, error) {
	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleURLAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}

type oauth2GoogleAuthenticatorRedirects struct {
	authFailure string
	logout      string
}

type Oauth2GoogleAuthenticator struct {
	store *sessions.CookieStore

	redirects    *oauth2GoogleAuthenticatorRedirects
	authCallback func(http.ResponseWriter, *http.Request) error
}

type Oauth2GoogleOption func(*Oauth2GoogleAuthenticator) error

func NewOauth2GoogleAuthenticator(opts ...Oauth2GoogleOption) (Authenticator, error) {
	a := &Oauth2GoogleAuthenticator{
		store:     Store,
		redirects: &oauth2GoogleAuthenticatorRedirects{},
	}

	for _, opt := range opts {
		err := opt(a)
		if err != nil {
			return nil, err
		}
	}

	return a, nil
}

func WithAuthRedirectURL(url string) Oauth2GoogleOption {
	return func(a *Oauth2GoogleAuthenticator) error {
		googleOauthConfig.RedirectURL = url
		return nil
	}
}

func WithAuthCallback(cb func(http.ResponseWriter, *http.Request) error) Oauth2GoogleOption {
	return func(a *Oauth2GoogleAuthenticator) error {
		a.authCallback = cb
		return nil
	}
}

func WithRedirectOnAuthFailure(url string) Oauth2GoogleOption {
	return func(a *Oauth2GoogleAuthenticator) error {
		a.redirects.authFailure = url
		return nil
	}
}

func WithRedirectToLoginOnAuthFailure() Oauth2GoogleOption {
	return func(a *Oauth2GoogleAuthenticator) error {
		a.redirects.authFailure = "/auth/google/login"
		return nil
	}
}

func WithRedirectOnLogout(url string) Oauth2GoogleOption {
	return func(a *Oauth2GoogleAuthenticator) error {
		a.redirects.logout = url
		return nil
	}
}

func WithCookieStoreForOauth2Google(givenStore *sessions.CookieStore) Oauth2GoogleOption {
	if givenStore == nil {
		givenStore = Store // default store
	}

	return func(a *Oauth2GoogleAuthenticator) error {
		a.store = givenStore
		return nil
	}
}

func (a *Oauth2GoogleAuthenticator) Routes() Routes {
	return Routes{
		"/auth/google/login":    a.Authenticate,
		"/auth/google/callback": a.callback,
		"/auth/google/logout":   a.Deauthenticate,
	}
}

func (a *Oauth2GoogleAuthenticator) ReadSessionValue(w http.ResponseWriter, r *http.Request, key string) (interface{}, bool) {
	session, err := a.store.Get(r, "oauthstate")
	log.Println(err)

	if err != nil {
		return nil, false
	}

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)

	if !ok || !auth {
		return nil, false
	}

	// Check if user is authenticated
	v, ok := session.Values[key]

	if ok {
		return v, true
	}

	return nil, false
}

func (a *Oauth2GoogleAuthenticator) RequireAuthentication(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("verifying request is authenticated")
		if !a.IsAuthenticated(w, r) {
			if a.redirects != nil && a.redirects.authFailure != "" {
				log.Println("redirecting due to auth failure to:", a.redirects.authFailure)
				w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
				w.Header().Set("Expires", time.Unix(0, 0).Format(http.TimeFormat))
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("X-Accel-Expires", "0")
				http.Redirect(w, r, a.redirects.authFailure, http.StatusPermanentRedirect)
			} else {
				http.NotFound(w, r)
			}
			return
		}
		h(w, r)
	}
}

func (a *Oauth2GoogleAuthenticator) IsAuthenticated(w http.ResponseWriter, r *http.Request) bool {
	log.Println("checking is request is authenticated")
	session, err := a.store.Get(r, "oauthstate")

	if err != nil {
		log.Println(err)
		return false
	}

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)

	if !ok || !auth {
		log.Println("request is NOT authenticated")
		return false
	}

	log.Println("request is authenticated")
	return true
}

func (a *Oauth2GoogleAuthenticator) Authenticate(w http.ResponseWriter, r *http.Request) {
	log.Println("attempting to authenticate:", r.RemoteAddr, "from", r.Referer())
	oauthState := generateStateOauthCookie(w)
	u := googleOauthConfig.AuthCodeURL(oauthState)
	log.Println("redirecting to google")
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
	return
}

func (a *Oauth2GoogleAuthenticator) Deauthenticate(w http.ResponseWriter, r *http.Request) {
	log.Println("attempting to deauthenticate:", r.RemoteAddr, "from", r.Referer())
	session, err := a.store.Get(r, "oauthstate")
	if err != nil {
		log.Println(err)
	}

	session.Values["authenticated"] = false
	session.Options.MaxAge = -1

	err = session.Save(r, w)

	if err != nil {
		http.Error(w, "Session save failure", http.StatusInternalServerError)
		return
	}

	log.Println("deauthenticated:", r.RemoteAddr, "from", r.Referer())

	if a.redirects != nil && a.redirects.logout != "" {
		http.Redirect(w, r, a.redirects.logout, http.StatusPermanentRedirect)
		return
	}

	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	return
}

func (a *Oauth2GoogleAuthenticator) callback(w http.ResponseWriter, r *http.Request) {
	log.Println("hit callback")
	oauthState, err := r.Cookie("oauthstate")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	log.Println("getting user data from google")
	data, err := getUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(data))

	authData := googleAuthData{}
	err = decoder.Decode(&authData)
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// never allow unverified emails to access the application
	if !authData.VerifiedEmail {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	session, _ := a.store.Get(r, "oauthstate")

	// setup the session
	log.Println("saving session")
	session.Values["authenticated"] = true
	session.Values["name"] = authData.Name
	session.Values["email"] = authData.Email
	log.Println(session)

	session.Save(r, w)

	log.Println("authenticated:", r.RemoteAddr)

	if a.authCallback != nil {
		log.Println("running auth call back")
		err = a.authCallback(w, r)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}

	log.Println("redirecting after authentication")
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	w.Header().Set("Expires", time.Unix(0, 0).Format(http.TimeFormat))
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Accel-Expires", "0")

	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	return
}
