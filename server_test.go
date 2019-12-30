package web

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"testing"
)

func TestNewServer(t *testing.T) {
	logger := log.New(os.Stderr, "test-web-server: ", log.LstdFlags)

	helloWorld := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	}

	server, err := NewServer(
		WithRoutes(
			Routes{"/": helloWorld},
			MiddlewareLogRequest(logger),
			MiddlewareLimitRequestBody(DefaultRequestBodySize),
		),
	)

	if err != nil {
		t.Fatal(err)
	}

	t.Log(server)
}

func TestNewServerWithGoogleOauth(t *testing.T) {
	logger := log.New(os.Stderr, "test-web-server: ", log.LstdFlags)

	authenticator, _ := NewOauth2GoogleAuthenticator(
		WithRedirectToLoginOnAuthFailure(),
		WithRedirectOnLogout("/goodbye"),
	)

	tmpl, _ := template.New("homePage").Parse(`
		<!DOCTYPE html>
		<html lang="en">
		<meta charset="utf-8">

		<body>
			Hello {{.}}

			<a href="/auth/google/logout">Logout</a>
		</body>
	`)

	helloWorld := func(w http.ResponseWriter, r *http.Request) {
		v, ok := authenticator.ReadSessionValue(w, r, "name")
		if ok {
			tmpl.Execute(w, v)
		}
	}

	mainRoutes := Routes{
		"/": helloWorld,
	}

	server, err := NewServer(
		WithRoutes(
			JoinRoutes(
				AuthenticatedRoutes(authenticator, mainRoutes),
				authenticator.Routes(),
			),
			MiddlewareLogRequest(logger),
		),
	)

	if err != nil {
		t.Fatal(err)
	}

	// go test -v -timeout 30m github.com/picatz/web -run TestNewServerWithGoogleOauth
	// Serve(server, nil, "", "")
}
