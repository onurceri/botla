package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/onurceri/botla-app/internal/integration/fixtures"
	"github.com/onurceri/botla-app/internal/models"
)

type handoffListPage struct {
	Requests []models.HandoffRequest `json:"requests"`
	Total    int                     `json:"total"`
	Page     int                     `json:"page"`
	Limit    int                     `json:"limit"`
}

func listHandoffs(t *testing.T, te *fixtures.TestEnv, token, botID, query string) (int, handoffListPage) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/chatbots/"+botID+"/handoff-requests"+query, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := testHTTPClient().Do(req)
	if err != nil {
		t.Fatalf("list handoffs: %v", err)
	}
	defer drainBody(res)
	var page handoffListPage
	if res.StatusCode == http.StatusOK {
		if err := json.NewDecoder(res.Body).Decode(&page); err != nil {
			t.Fatalf("decode list: %v", err)
		}
	}
	return res.StatusCode, page
}

func TestHandoff_List_PaginationAndStatusFilter(t *testing.T) {
	t.Parallel()
	te, err := fixtures.SetupTestEnv()
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)

	token, botID := setupHandoffBot(t, te, "handoff_list@example.com")

	// 5 requests (one per session); mark 2 of them resolved.
	ids := make([]string, 0, 5)
	for i := 0; i < 5; i++ {
		ids = append(ids, requestPublicHandoff(t, te, botID, fmt.Sprintf("list-session-%d", i)))
	}
	if _, err := te.DB.Exec("UPDATE handoff_requests SET status='resolved' WHERE id = ANY($1::uuid[])",
		fmt.Sprintf("{%s,%s}", ids[0], ids[1])); err != nil {
		t.Fatalf("resolve requests: %v", err)
	}

	code, all := listHandoffs(t, te, token, botID, "")
	if code != http.StatusOK || all.Total != 5 || len(all.Requests) != 5 || all.Page != 1 || all.Limit != 50 {
		t.Fatalf("default list: code=%d page=%+v", code, all)
	}

	code, p1 := listHandoffs(t, te, token, botID, "?limit=2&page=1")
	_, p3 := listHandoffs(t, te, token, botID, "?limit=2&page=3")
	if code != http.StatusOK || p1.Total != 5 || len(p1.Requests) != 2 || len(p3.Requests) != 1 {
		t.Fatalf("paging: p1=%+v p3=%+v", p1, p3)
	}
	seen := map[string]bool{}
	for _, pg := range []string{"?limit=2&page=1", "?limit=2&page=2", "?limit=2&page=3"} {
		_, p := listHandoffs(t, te, token, botID, pg)
		for _, r := range p.Requests {
			if seen[r.ID] {
				t.Fatalf("request %s returned on more than one page", r.ID)
			}
			seen[r.ID] = true
		}
	}
	if len(seen) != 5 {
		t.Fatalf("pages should cover all 5 requests, saw %d", len(seen))
	}

	_, resolved := listHandoffs(t, te, token, botID, "?status=resolved")
	if resolved.Total != 2 || len(resolved.Requests) != 2 {
		t.Fatalf("resolved filter: %+v", resolved)
	}
	for _, r := range resolved.Requests {
		if r.Status != models.HandoffStatusResolved {
			t.Fatalf("filter leaked status %s", r.Status)
		}
	}
	_, pending := listHandoffs(t, te, token, botID, "?status=pending&limit=1")
	if pending.Total != 3 || len(pending.Requests) != 1 {
		t.Fatalf("pending filter: %+v", pending)
	}

	if code, _ := listHandoffs(t, te, token, botID, "?status=bogus"); code != http.StatusBadRequest {
		t.Fatalf("bogus status: expected 400, got %d", code)
	}
	if code, _ := listHandoffs(t, te, token, botID, "?limit=1000"); code != http.StatusBadRequest {
		t.Fatalf("huge limit: expected 400, got %d", code)
	}
}
