package web

import "testing"

func TestNewServer(t *testing.T) {
	server, err := NewServer()

	if err != nil {
		t.Fatal(err)
	}

	t.Log(server)
}
