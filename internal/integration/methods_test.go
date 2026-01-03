package integration

import (
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestMethods_NotAllowed(t *testing.T) {
t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	res, _ := testHTTPGet(te.Server.URL + "/api/v1/chatbots/any/chat")
	if res.StatusCode != http.StatusUnauthorized && res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	drainBody(res)
}
