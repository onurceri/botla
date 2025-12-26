package db_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
)

func TestConversation_Messages_DB(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	uid := createUser(t, dbConn)
	// create bot
	b := &models.Chatbot{UserID: uid, Name: "Conv Bot", SystemPrompt: "p", LanguageCode: "en-US", Model: "gpt-3.5-turbo", Temperature: 0.1, MaxTokens: 64, ThemeColor: "#000000", WelcomeMessage: "hi", Position: "bottom-right", BotMessageColor: "#000000", UserMessageColor: "#ffffff", BotMessageTextColor: "#ffffff", UserMessageTextColor: "#000000", ChatFontFamily: "Inter", ChatHeaderColor: "#000000", ChatHeaderTextColor: "#ffffff", ChatBackgroundColor: "#ffffff"}
	bid, err := db.CreateChatbot(context.Background(), dbConn, b)
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}
	// conversation by session
	sid := fmt.Sprintf("s-%d", time.Now().UnixNano())
	conv, err := db.GetOrCreateConversationBySessionID(context.Background(), dbConn, bid, sid)
	if err != nil || conv == nil {
		t.Fatalf("conv: %v", err)
	}
	// create message
	mid, err := db.CreateMessage(context.Background(), dbConn, &models.Message{ConversationID: conv.ID, Role: "user", Content: "hello", TokensUsed: 1})
	if err != nil || mid == "" {
		t.Fatalf("create msg: %v", err)
	}
	// increment count
	if err2 := db.IncrementConversationMessageCount(context.Background(), dbConn, conv.ID); err2 != nil {
		t.Fatalf("inc count: %v", err2)
	}
	// list recent
	msgs, err := db.ListRecentMessages(context.Background(), dbConn, conv.ID, 5)
	if err != nil || len(msgs) == 0 {
		t.Fatalf("list msgs: %v", err)
	}
	// feedback
	_, _, err = db.UpdateMessageFeedback(context.Background(), dbConn, msgs[len(msgs)-1].ID, true)
	if err != nil {
		t.Fatalf("update feedback: %v", err)
	}
}

func TestGetOrCreateConversation_Concurrency(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	uid := createUser(t, dbConn)
	// create bot
	b := &models.Chatbot{UserID: uid, Name: "Race Bot", SystemPrompt: "p", LanguageCode: "en-US", Model: "gpt-3.5-turbo", Temperature: 0.1, MaxTokens: 64, ThemeColor: "#000000", WelcomeMessage: "hi", Position: "bottom-right", BotMessageColor: "#000000", UserMessageColor: "#ffffff", BotMessageTextColor: "#ffffff", UserMessageTextColor: "#000000", ChatFontFamily: "Inter", ChatHeaderColor: "#000000", ChatHeaderTextColor: "#ffffff", ChatBackgroundColor: "#ffffff"}
	bid, err := db.CreateChatbot(context.Background(), dbConn, b)
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}

	sessionID := fmt.Sprintf("race-session-%d", time.Now().UnixNano())
	concurrency := 10
	errCh := make(chan error, concurrency)
	idCh := make(chan string, concurrency)

	// Launch concurrent requests
	for i := 0; i < concurrency; i++ {
		go func() {
			c, err := db.GetOrCreateConversationBySessionID(context.Background(), dbConn, bid, sessionID)
			if err != nil {
				errCh <- err
				return
			}
			idCh <- c.ID
			errCh <- nil
		}()
	}

	// Collect results
	ids := make([]string, 0)
	for i := 0; i < concurrency; i++ {
		if err := <-errCh; err != nil {
			t.Errorf("concurrent request failed: %v", err)
		}
		if id := <-idCh; id != "" {
			ids = append(ids, id)
		}
	}

	// Verify all returned same ID
	if len(ids) == 0 {
		t.Fatal("no ids returned")
	}
	firstID := ids[0]
	for _, id := range ids {
		if id != firstID {
			t.Errorf("got different conversation IDs for same session: %s vs %s", firstID, id)
		}
	}

	// Verify only 1 record exists
	var count int
	err = dbConn.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM conversations WHERE session_id=$1", sessionID).Scan(&count)
	if err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 conversation, got %d", count)
	}
}
