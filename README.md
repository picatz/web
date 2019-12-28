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
  server, _ := web.NewServer()

  web.Serve(server, nil, "", "")
}

```
