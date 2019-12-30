package web

import (
	"crypto/rand"

	"github.com/gorilla/sessions"
)

func GenerateNewRandKey() []byte {
	c := 2048
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func GenerateNewCookiStore() *sessions.CookieStore {
	return sessions.NewCookieStore(GenerateNewRandKey())
}

var (
	store = GenerateNewCookiStore()
)
