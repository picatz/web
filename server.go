package web

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	DefaultServerReadTimeout  = 5 * time.Second
	DefaultServerWriteTimeout = 5 * time.Second
	DefaultServerIdleTimeout  = 10 * time.Second
)

var DefaultServerTLSConfigCipherSuites = []uint16{
	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
	tls.TLS_RSA_WITH_AES_256_CBC_SHA,
}

var DefaultServerTLSConfig = &tls.Config{
	MinVersion:               tls.VersionTLS12,
	CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
	CipherSuites:             DefaultServerTLSConfigCipherSuites,
	PreferServerCipherSuites: true,
}

var DefaultServerTLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)

type ServerOption = func(*http.Server) error

func NewServer(opts ...ServerOption) (*http.Server, error) {
	var mux = http.NewServeMux()

	var server = &http.Server{
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

func WithDefaultServerOptions() []ServerOption {
	return []ServerOption{
		WithServerReadTimeout(DefaultServerReadTimeout),
		WithServerWriteTimeout(DefaultServerWriteTimeout),
		WithServerIdleTimeout(DefaultServerIdleTimeout),
	}
}

func WithServerReadTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) error {
		s.ReadTimeout = d
		return nil
	}
}

func WithServerWriteTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) error {
		s.WriteTimeout = d
		return nil
	}
}

func WithServerIdleTimeout(d time.Duration) ServerOption {
	return func(s *http.Server) error {
		s.IdleTimeout = d
		return nil
	}
}

type Routes = map[string]func(http.ResponseWriter, *http.Request)

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

func WithServerTLSConfig(c *tls.Config) ServerOption {
	return func(s *http.Server) error {
		s.TLSConfig = c
		return nil
	}
}

func WithServerDefaultTLSConfig() ServerOption {
	return func(s *http.Server) error {
		s.TLSConfig = DefaultServerTLSConfig
		return nil
	}
}

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
