# web

> 🕸 Your friendly neighborhood HTTP client and server

## Download

```console
$ go get -u -v github.com/picatz/web
...
```

## Client Usage

```go
package main

import (
  "fmt"

  "github.com/picatz/web"
)

func main() {
  client, _ := web.NewClient()

  resp, err := client.Get("https://www.google.com")

  if err != nil {
    panic(err)
  }

  fmt.Println(resp.StatusCode)
}
```

## Server Usage

```go
package main

import (
  "log"
  "net/http"
  "os"

  "github.com/picatz/web"
)

func main() {
  logger := log.New(os.Stderr, "example-web-server: ", log.LstdFlags)

  helloWorld := func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello World!"))
  }

  server, _ := web.NewServer(
    web.WithRoutes(
      web.Routes{"/": helloWorld},
      web.MiddlewareLogRequest(logger),
      web.MiddlewareLimitRequestBody(web.DefaultRequestBodySize),
    ),
  )

  web.Serve(server, nil, "", "")
}
```

## Server Authentication

A simple Google Oauth2 authenticator implementation is built in:

```console
$

```

```go
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
```
