package web

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
)

// Default server timeout values.
var (
	DefaultServerReadTimeout  = 5 * time.Second
	DefaultServerWriteTimeout = 5 * time.Second
	DefaultServerIdleTimeout  = 10 * time.Second
)

// DefaultServerTLSConfigCipherSuites contains TLS cipher configuration.
var DefaultServerTLSConfigCipherSuites = []uint16{
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
}

// DefaultServerTLSConfig defines the default TLS configuration for a server.
var DefaultServerTLSConfig = &tls.Config{
	MinVersion:               tls.VersionTLS12,
	CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
	CipherSuites:             DefaultServerTLSConfigCipherSuites,
	PreferServerCipherSuites: true,
}

// DefaultServerTLSNextProto defines the default TLS protocol logic for a server.
var DefaultServerTLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)

// ServerOption is a function which always for server configurations.
type ServerOption = func(*http.Server) error

// NewServer simplifies the creation of a new HTTP server using an optional list
// of ServerOptions for configuration.
func NewServer(opts ...ServerOption) (*http.Server, error) {
	var mux = http.NewServeMux()

	var server = &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      mux,
		ReadTimeout:  DefaultServerReadTimeout,
		WriteTimeout: DefaultServerWriteTimeout,
		IdleTimeout:  DefaultServerIdleTimeout,
	}

	for _, opt := range opts {
		err := opt(server)
		if err != nil {
			return nil, err
		}
	}

	return server, nil
}

// WithDefaultServerOptions provides an example method to use
// with the NewServer options using the default timeout values.
func WithDefaultServerOptions() []ServerOption {
	return []ServerOption{
		WithServerReadTimeout(DefaultServerReadTimeout),
		WithServerWriteTimeout(DefaultServerWriteTimeout),
		WithServerIdleTimeout(DefaultServerIdleTimeout),
	}
}

// WithServerReadTimeout can change the server's read timeout.
func WithServerReadTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) error {
		s.ReadTimeout = d
		return nil
	}
}

// WithServerWriteTimeout can change the server's write timeout.
func WithServerWriteTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) error {
		s.WriteTimeout = d
		return nil
	}
}

// WithServerIdleTimeout can change the server's idle timeout.
func WithServerIdleTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) error {
		s.IdleTimeout = d
		return nil
	}
}

// Routes contains a { key -> value } mapping of { pathString -> http.HandlerFunc }
type Routes = map[string]http.HandlerFunc

// JoinRoutes takes multiple Routes objects and merges them into one Routes object.
func JoinRoutes(r ...Routes) Routes {
	newRoutes := Routes{}

	for _, routes := range r {
		for path, handlerFunc := range routes {
			newRoutes[path] = handlerFunc
		}
	}

	return newRoutes
}

// AuthenticatedRoutes takes multiple Routes and requires them all to be authenticated.
func AuthenticatedRoutes(a Authenticator, r ...Routes) Routes {
	newRoutes := Routes{}

	for _, routes := range r {
		for path, handlerFunc := range routes {
			newRoutes[path] = a.RequireAuthentication(handlerFunc)
		}
	}

	return newRoutes
}

// WithRoutes allows for custom routes to be added to a server.
//
// Note: you can only use this method once within a NewServer method.
func WithRoutes(r Routes, m ...RouteMiddleware) ServerOption {
	return func(s *http.Server) error {
		var mux = http.NewServeMux()

		for path, handleFunc := range r {
			for _, middleware := range m {
				handleFunc = middleware(handleFunc)
			}
			mux.HandleFunc(path, handleFunc)
		}

		s.Handler = mux

		return nil
	}
}

// WithServerTLSConfig can change the server's TLS configurations.
func WithServerTLSConfig(c *tls.Config) ServerOption {
	return func(s *http.Server) error {
		s.TLSConfig = c
		return nil
	}
}

// WithServerDefaultTLSConfig configures the server to use the DefaultServerTLSConfig.
func WithServerDefaultTLSConfig() ServerOption {
	return func(s *http.Server) error {
		s.TLSConfig = DefaultServerTLSConfig
		return nil
	}
}

// WithAddr configures the IP address and port to sserve requests.
func WithAddr(addr string) ServerOption {
	return func(s *http.Server) error {
		s.Addr = addr
		return nil
	}
}

// WithInMemoryCookieStore configures an in-memory cookie store.
func WithInMemoryCookieStore(keyPairs ...[]byte) ServerOption {
	return func(s *http.Server) error {
		Store = sessions.NewCookieStore(keyPairs...)
		return nil
	}
}

// Serve starts the listener.
func Serve(s *http.Server, l *log.Logger, certFile, keyFile string) {
	if l == nil {
		l = log.New(os.Stderr, "web-server-error: ", log.LstdFlags)
	}
	if s.TLSConfig != nil {
		l.Fatal(s.ListenAndServeTLS(certFile, keyFile))
	} else {
		l.Fatal(s.ListenAndServe())
	}
}
