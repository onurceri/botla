package integration

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
)

func TestAnalytics_SeriesLength30(t *testing.T) {
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	token := authToken(t, te.Server.URL, "series@example.com")
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/analytics", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var data []map[string]any
	json.NewDecoder(res.Body).Decode(&data)
	res.Body.Close()
	if len(data) != 30 {
		t.Fatalf("expected 30 points, got %d", len(data))
	}
}
