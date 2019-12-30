package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/picatz/web"
)

func main() {
	logger := log.New(os.Stderr, "test-web-server: ", log.LstdFlags)

	authenticator, _ := web.NewOauth2GoogleAuthenticator(
		web.WithRedirectToLoginOnAuthFailure(),
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

	mainRoutes := web.Routes{
		"/": helloWorld,
	}

	server, _ := web.NewServer(
		web.WithRoutes(
			web.JoinRoutes(
				web.AuthenticatedRoutes(authenticator, mainRoutes),
				authenticator.Routes(),
			),
			web.MiddlewareLogRequest(logger),
		),
	)

	web.Serve(server, nil, "", "")
}
