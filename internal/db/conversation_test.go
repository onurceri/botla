package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
)

func TestConversation_Messages_DB(t *testing.T) {
	db := testdb.OpenTestDB(t)
	defer db.Close()
	uid := createUser(t, db)
	// create bot
	b := &models.Chatbot{UserID: uid, Name: "Conv Bot", SystemPrompt: "p", LanguageCode: "en-US", Model: "gpt-3.5-turbo", Temperature: 0.1, MaxTokens: 64, ThemeColor: "#000000", WelcomeMessage: "hi", Position: "bottom-right", BotMessageColor: "#000000", UserMessageColor: "#ffffff", BotMessageTextColor: "#ffffff", UserMessageTextColor: "#000000", ChatFontFamily: "Inter", ChatHeaderColor: "#000000", ChatHeaderTextColor: "#ffffff", ChatBackgroundColor: "#ffffff"}
	bid, err := CreateChatbot(context.Background(), db, b)
	if err != nil {
		t.Fatalf("create bot: %v", err)
	}
	// conversation by session
	sid := fmt.Sprintf("s-%d", time.Now().UnixNano())
	conv, err := GetOrCreateConversationBySessionID(context.Background(), db, bid, sid)
	if err != nil || conv == nil {
		t.Fatalf("conv: %v", err)
	}
	// create message
	mid, err := CreateMessage(context.Background(), db, &models.Message{ConversationID: conv.ID, Role: "user", Content: "hello", TokensUsed: 1})
	if err != nil || mid == "" {
		t.Fatalf("create msg: %v", err)
	}
	// increment count
	if err2 := IncrementConversationMessageCount(context.Background(), db, conv.ID); err2 != nil {
		t.Fatalf("inc count: %v", err2)
	}
	// list recent
	msgs, err := ListRecentMessages(context.Background(), db, conv.ID, 5)
	if err != nil || len(msgs) == 0 {
		t.Fatalf("list msgs: %v", err)
	}
	// feedback
	_, _, err = UpdateMessageFeedback(context.Background(), db, msgs[len(msgs)-1].ID, true)
	if err != nil {
		t.Fatalf("update feedback: %v", err)
	}
}
