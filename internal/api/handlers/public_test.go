package handlers

import (
	"testing"
	"time"

	"github.com/onurceri/botla-co/internal/models"
)

func TestPublicSuggestionsCacheKey(t *testing.T) {
	c := &models.Chatbot{ID: "x", UpdatedAt: time.Now()}
	k := publicSuggestionsCacheKey(c)
	if k == "" {
		t.Fatalf("cache key should not be empty")
	}
	if want := "public:chatbot:"; k[:len(want)] != want {
		t.Fatalf("unexpected prefix: %q", k)
	}
}
