package web

import (
	"io"
	"log"
	"net/http"

	"golang.org/x/crypto/ssh"
)

type RouteMiddleware func(http.HandlerFunc) http.HandlerFunc

func MiddlewareLogRequest(logger *log.Logger) RouteMiddleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			logger.Println(r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			h(w, r)
		}
	}
}

var DefaultRequestBodySize = int64(1024)

func MiddlewareLimitRequestBody(numBytes int64) RouteMiddleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, numBytes)
			h(w, r)
		}
	}
}

func MiddlewareHeader(headerKey, headerValue string) RouteMiddleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(headerKey, headerValue)
			h(w, r)
		}
	}
}

func MiddlewareHSTS() RouteMiddleware {
	return MiddlewareHeader("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
}

func MiddlewareContentType(contentType string) RouteMiddleware {
	return MiddlewareHeader("Content-Type", contentType)
}

func MiddlewareContentTypeUTF8(contentType string) RouteMiddleware {
	return MiddlewareHeader("Content-Type", contentType+"; charset=UTF-8")
}

func MiddlewareContentTypeJSON() RouteMiddleware {
	return MiddlewareContentTypeUTF8("application/json")
}

func MiddlewareSSHProxy(sshAddr string, cfg *ssh.ClientConfig, remoteAddr string) RouteMiddleware {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Establish connection with SSH server
			conn, err := ssh.Dial("tcp", sshAddr, cfg)
			if err != nil {
				log.Fatalln(err)
			}
			defer conn.Close()

			// Establish connection with remote server
			remote, err := conn.Dial("tcp", remoteAddr)
			if err != nil {
				log.Fatalln(err)
			}
			defer remote.Close()

			io.Copy(remote, r.Body)
			h(w, r)
		}
	}
}
