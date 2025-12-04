package integration

import (
	"net/http"
	"testing"
)

func TestMethods_NotAllowed(t *testing.T) {
	te, err := SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer TeardownTestEnv(te)
	res, _ := http.Get(te.Server.URL + "/api/v1/chatbots/any/chat")
	if res.StatusCode != http.StatusUnauthorized && res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	res.Body.Close()
}
