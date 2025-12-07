package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/internal/scraper"
)

// ChatbotCache provides caching for chatbot lookups
type ChatbotCache struct {
	cache scraper.Cache
	db    *sql.DB
	ttl   time.Duration
}

// NewChatbotCache creates a new cached chatbot repository
func NewChatbotCache(db *sql.DB) *ChatbotCache {
	return &ChatbotCache{
		cache: scraper.NewCache(),
		db:    db,
		ttl:   5 * time.Minute,
	}
}

// GetByID retrieves a chatbot by ID, using cache when available
func (c *ChatbotCache) GetByID(ctx context.Context, id string) (*models.Chatbot, error) {
	key := "chatbot:" + id

	// Try cache first
	if cached, ok := c.cache.Get(key); ok {
		var chatbot models.Chatbot
		if err := json.Unmarshal([]byte(cached), &chatbot); err == nil {
			return &chatbot, nil
		}
	}

	// Fallback to database
	chatbot, err := GetChatbotByID(ctx, c.db, id)
	if err != nil || chatbot == nil {
		return chatbot, err
	}

	// Cache the result
	if data, err := json.Marshal(chatbot); err == nil {
		_ = c.cache.Set(key, string(data), c.ttl)
	}

	return chatbot, nil
}

// Invalidate removes a chatbot from cache (call after updates)
func (c *ChatbotCache) Invalidate(id string) {
	// Cache doesn't support delete, but we can set with very short TTL
	// In practice, this will be a no-op for memory cache (TTL will expire)
	// For Redis, we'd want a Delete method, but this is good enough for now
	_ = c.cache.Set("chatbot:"+id, "", 1*time.Millisecond)
}
