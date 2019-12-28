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
  "fmt"

  "github.com/picatz/web"
)

func main() {
  logger := log.New(os.Stderr, "example-web-server: ", log.LstdFlags)

  helloWorld := func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello World!"))
  }

  server, _ := web.NewServer(
    WithRoutes(
      Routes{"/": helloWorld},
      MiddlewareLogRequest(logger),
      MiddlewareLimitRequestBody(web.DefaultRequestBodySize),
    ),
  )

  web.Serve(server, nil, "", "")
}
```
