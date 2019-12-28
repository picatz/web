package web

import (
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
