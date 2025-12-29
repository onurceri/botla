package db_test

import (
	"context"
	"testing"

	"github.com/onurceri/botla-co/internal/db"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/testdb"
	"github.com/stretchr/testify/assert"
)

func TestChatbot_Update_PreservesSuggestedQuestions(t *testing.T) {
	dbConn := testdb.OpenTestDB(t)
	uid := createUser(t, dbConn)
	t.Cleanup(func() {
		_, _ = dbConn.Exec(`DELETE FROM users WHERE id = $1`, uid)
	})

	initialQuestions := []string{"How are you?", "Who are you?"}
	manualQuestions := []string{"What is your return policy?"}

	b := &models.Chatbot{
		UserID:             uid,
		Name:               "Test Bot",
		LanguageCode:       "en",
		Model:              "gpt-3.5-turbo",
		SuggestedQuestions: initialQuestions,
		ManualQuestions:    manualQuestions,
		SuggestionsEnabled: true,
	}

	id, err := db.CreateChatbot(context.Background(), dbConn, b)
	if err != nil {
		t.Fatalf("create chatbot: %v", err)
	}
	t.Cleanup(func() {
		_, _ = dbConn.Exec(`DELETE FROM chatbots WHERE id = $1`, id)
	})

	// Verify initial state
	initial, err := db.GetChatbotByID(context.Background(), dbConn, id)
	assert.NoError(t, err)
	assert.Equal(t, initialQuestions, initial.SuggestedQuestions)

	// Simulate AI regeneration via UpdateChatbotSuggestedQuestions
	newAIQuestions := []string{"What can you help with?", "Tell me about pricing"}
	err = db.UpdateChatbotSuggestedQuestions(context.Background(), dbConn, id, newAIQuestions)
	assert.NoError(t, err)

	// Re-verify: AI questions should change, but manual should persist
	preUpdate, err := db.GetChatbotByID(context.Background(), dbConn, id)
	assert.NoError(t, err)
	assert.Equal(t, newAIQuestions, preUpdate.SuggestedQuestions)
	assert.Equal(t, manualQuestions, preUpdate.ManualQuestions, "ManualQuestions should persist through regeneration")

	// Now simulate a UI update where we modify something else (e.g. ThemeColor)
	// AND we pass nil for SuggestedQuestions
	updateReq := &models.Chatbot{
		ID:                 id,
		UserID:             uid,
		Name:               "Test Bot Updated",
		ThemeColor:         "#ff0000",
		SuggestedQuestions: nil, // Simulate missing field
		LanguageCode:       "en",
		Model:              "gpt-4",
	}

	err = db.UpdateChatbot(context.Background(), dbConn, updateReq)
	assert.NoError(t, err)

	// Verify post update
	postUpdate, err := db.GetChatbotByID(context.Background(), dbConn, id)
	assert.NoError(t, err)

	// Name should change
	assert.Equal(t, "Test Bot Updated", postUpdate.Name)

	// SuggestedQuestions should be PRESERVED (not nil/empty) due to COALESCE
	assert.NotNil(t, postUpdate.SuggestedQuestions)
	assert.Equal(t, newAIQuestions, postUpdate.SuggestedQuestions, "SuggestedQuestions should be preserved")

	// ManualQuestions should also be preserved
	assert.NotNil(t, postUpdate.ManualQuestions)
	assert.Equal(t, manualQuestions, postUpdate.ManualQuestions, "ManualQuestions should be preserved")
}

