package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/integration/fixtures"
	"github.com/onurceri/botla-co/pkg/config"
)

func TestAnalytics_ThumbsUpAfterFeedback(t *testing.T) {
	oai := fixtures.NewLLMMock(t)
	qd := startQdrantStub()
	te, err := fixtures.SetupTestEnvWithConfigAndMocks(func(cfg *config.Config) {
		cfg.OPENAI_API_BASE = oai.URL
		cfg.OPENROUTER_API_BASE = oai.URL + "/v1"
		cfg.QDRANT_URL = qd.URL
	}, false)
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	defer fixtures.TeardownTestEnv(te)
	defer oai.Close()
	defer qd.Close()

	token, err := te.AuthToken("fbupd@example.com")
	if err != nil {
		t.Fatalf("auth token failed: %v", err)
	}

	// create chatbot
	create := map[string]any{"name": "FB Upd Bot"}
	cbj, _ := json.Marshal(create)
	reqC, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots", bytes.NewReader(cbj))
	reqC.Header.Set("Authorization", "Bearer "+token)
	reqC.Header.Set("Content-Type", "application/json")
	resC, _ := testHTTPClient().Do(reqC)
	var bot chatbot
	json.NewDecoder(resC.Body).Decode(&bot)
	resC.Body.Close()

	// perform chat to create messages (and conversation)
	sess := "s7"
	cr := chatReq{Message: "merhaba", SessionID: sess}
	crb, _ := json.Marshal(cr)
	reqCh, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/chatbots/"+bot.ID+"/chat", bytes.NewReader(crb))
	reqCh.Header.Set("Authorization", "Bearer "+token)
	reqCh.Header.Set("Content-Type", "application/json")
	resCh, _ := testHTTPClient().Do(reqCh)
	if resCh.StatusCode != http.StatusOK {
		var errB bytes.Buffer
		errB.ReadFrom(resCh.Body)
		t.Fatalf("expected 200, got %d. Body: %s", resCh.StatusCode, errB.String())
	}
	resCh.Body.Close()

	// find conversation by session
	var convID string
	err = te.DB.QueryRow("SELECT id FROM conversations WHERE chatbot_id=$1 AND session_id=$2", bot.ID, sess).Scan(&convID)
	if err != nil {
		t.Fatalf("conv query error: %v", err)
	}

	// find a message id
	var msgID string
	err = te.DB.QueryRow("SELECT id FROM messages WHERE conversation_id=$1 ORDER BY created_at DESC LIMIT 1", convID).Scan(&msgID)
	if err != nil {
		t.Fatalf("msg query error: %v", err)
	}

	// send feedback thumbs_up=true
	fb := map[string]any{"thumbs_up": true}
	fbj, _ := json.Marshal(fb)
	reqF, _ := http.NewRequest(http.MethodPost, te.Server.URL+"/api/v1/messages/"+msgID+"/feedback", bytes.NewReader(fbj))
	reqF.Header.Set("Authorization", "Bearer "+token)
	reqF.Header.Set("Content-Type", "application/json")
	resF, _ := testHTTPClient().Do(reqF)
	if resF.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resF.StatusCode)
	}
	resF.Body.Close()

	// verify analytics thumbs_up_count increased (sum across series >=1) with retry
	var sumThumbs int
	found := false
	for i := 0; i < 10; i++ {
		reqA, _ := http.NewRequest(http.MethodGet, te.Server.URL+"/api/v1/analytics", nil)
		reqA.Header.Set("Authorization", "Bearer "+token)
		resA, _ := testHTTPClient().Do(reqA)
		if resA.StatusCode == http.StatusOK {
			var series []map[string]any
			json.NewDecoder(resA.Body).Decode(&series)
			resA.Body.Close()
			sumThumbs = 0
			for _, p := range series {
				if v, ok := p["thumbs_up_count"]; ok {
					switch vv := v.(type) {
					case float64:
						sumThumbs += int(vv)
					case int:
						sumThumbs += vv
					}
				}
			}
		}

		if sumThumbs == 0 {
			var cnt sql.NullInt64
			err = te.DB.QueryRow("SELECT COALESCE(SUM(thumbs_up_count),0) FROM analytics WHERE chatbot_id=$1 AND analytics_date=CURRENT_DATE", bot.ID).Scan(&cnt)
			if err == nil {
				sumThumbs = int(cnt.Int64)
			}
		}

		if sumThumbs >= 1 {
			found = true
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	if !found {
		t.Fatalf("expected thumbs_up_count >=1, got %d", sumThumbs)
	}
}
