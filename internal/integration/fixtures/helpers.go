package fixtures

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/onurceri/botla-co/internal/models"
	"github.com/onurceri/botla-co/pkg/policy"
)

func (e *TestEnv) CreateUser(email string) (*models.User, error) {
	id := uuid.New().String()
	// simple hash for tests
	passwordHash := "hash"

	var planID string
	err := e.DB.QueryRow("SELECT id FROM plans WHERE code = $1", policy.PlanFree.String()).Scan(&planID)
	if err != nil {
		return nil, fmt.Errorf("get free plan: %w", err)
	}

	user := &models.User{
		ID:                  id,
		Email:               email,
		PlanID:              &planID,
		CreatedAt:           time.Now(),
		IsPlatformAdmin:     false,
		OnboardingCompleted: true,
		OnboardingStep:      0,
		OnboardingSkipped:   false,
	}

	_, err = e.DB.Exec(`
		INSERT INTO users (id, email, password_hash, plan_id, created_at, is_platform_admin, onboarding_completed, onboarding_step, onboarding_skipped)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		user.ID, user.Email, passwordHash, *user.PlanID, user.CreatedAt, user.IsPlatformAdmin, user.OnboardingCompleted, user.OnboardingStep, user.OnboardingSkipped,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (e *TestEnv) CreateChatbot(user *models.User, name string) (*models.Chatbot, error) {
	id := uuid.New().String()

	var languageID string
	err := e.DB.QueryRow("SELECT id FROM languages WHERE code = 'en-US'").Scan(&languageID)
	if err != nil {
		return nil, fmt.Errorf("get language: %w", err)
	}

	bot := &models.Chatbot{
		ID:                   id,
		UserID:               user.ID,
		Name:                 name,
		Model:                policy.ModelGPT4oMini.String(),
		Temperature:          0.7,
		MaxTokens:            4096,
		ThemeColor:           "#000000",
		WelcomeMessage:       "Hello",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		Position:             "bottom-right",
		BotMessageColor:      "#ffffff",
		UserMessageColor:     "#000000",
		BotMessageTextColor:  "#000000",
		UserMessageTextColor: "#ffffff",
		ChatFontFamily:       "Inter",
		ChatHeaderColor:      "#ffffff",
		ChatHeaderTextColor:  "#000000",
		ChatBackgroundColor:  "#ffffff",
		BubbleRadius:         "12px",
		InputBackgroundColor: "#ffffff",
		InputTextColor:       "#000000",
		SendButtonColor:      "#000000",
		DiscoveryMode:        "auto",
		RefreshPolicy:        "manual",
		HideBranding:         false,
		ConfidenceThreshold:  0.5,
		HandoffEnabled:       false,
		HandoffType:          "email",
		LanguageCode:         "en-US",
		SecureEmbedEnabled:   false,
		SuggestionsEnabled:   false,
	}

	_, err = e.DB.Exec(`INSERT INTO chatbots (
		id, user_id, name, model, temperature, max_tokens, 
		theme_color, welcome_message, created_at, updated_at, 
		position, bot_message_color, user_message_color, bot_message_text_color, user_message_text_color, 
		chat_font_family, chat_header_color, chat_header_text_color, chat_background_color, bubble_radius, 
		input_background_color, input_text_color, send_button_color, 
		discovery_mode, refresh_policy, hide_branding, confidence_threshold, handoff_enabled, handoff_type, 
		language_id, secure_embed_enabled, suggestions_enabled
	) VALUES (
		$1, $2, $3, $4, $5, $6, 
		$7, $8, $9, $10, 
		$11, $12, $13, $14, $15, 
		$16, $17, $18, $19, $20, 
		$21, $22, $23, 
		$24, $25, $26, $27, $28, $29, 
		$30, $31, $32
	)`,
		bot.ID, bot.UserID, bot.Name, bot.Model, bot.Temperature, bot.MaxTokens,
		bot.ThemeColor, bot.WelcomeMessage, bot.CreatedAt, bot.UpdatedAt,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor, bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor, bot.ChatBackgroundColor, bot.BubbleRadius,
		bot.InputBackgroundColor, bot.InputTextColor, bot.SendButtonColor,
		bot.DiscoveryMode, bot.RefreshPolicy, bot.HideBranding, bot.ConfidenceThreshold, bot.HandoffEnabled, bot.HandoffType,
		languageID, bot.SecureEmbedEnabled, bot.SuggestionsEnabled,
	)
	if err != nil {
		return nil, fmt.Errorf("insert chatbot: %w", err)
	}

	return bot, nil
}

func (e *TestEnv) CreateSource(bot *models.Chatbot, url string) (*models.DataSource, error) {
	id := uuid.New().String()

	source := &models.DataSource{
		ID:           id,
		ChatbotID:    bot.ID,
		SourceType:   "website",
		SourceURL:    &url,
		Status:       "completed",
		ChunkCount:   1,
		SizeBytes:    1024,
		IsDiscovered: false,
		CreatedAt:    time.Now(),
	}

	_, err := e.DB.Exec(`INSERT INTO data_sources (id, chatbot_id, source_type, source_url, status, chunk_count, size_bytes, is_discovered, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		source.ID, source.ChatbotID, source.SourceType, source.SourceURL, source.Status,
		source.ChunkCount, source.SizeBytes, source.IsDiscovered, source.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert source: %w", err)
	}

	return source, nil
}

func (e *TestEnv) AuthToken(email string) (string, error) {
	regBody := map[string]string{
		"email":     email,
		"password":  TestPassword,
		"full_name": "Test User",
	}
	regJSON, _ := json.Marshal(regBody)
	_, _ = http.Post(e.Server.URL+"/api/v1/auth/register", "application/json", bytes.NewBuffer(regJSON))

	loginBody := map[string]string{
		"email":    email,
		"password": TestPassword,
	}
	loginJSON, _ := json.Marshal(loginBody)
	loginResp, err := http.Post(e.Server.URL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		return "", fmt.Errorf("login failed: %w", err)
	}
	defer func() {
		_ = loginResp.Body.Close()
	}()

	var tokenResp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(loginResp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token failed: %w", err)
	}

	return tokenResp.Token, nil
}

func (e *TestEnv) UpdateUserPlan(email, planCode string) error {
	_, err := e.DB.Exec(`UPDATE users SET plan_id = (SELECT id FROM plans WHERE code = $1) WHERE email = $2`, planCode, email)
	if err != nil {
		return fmt.Errorf("update user plan: %w", err)
	}
	return nil
}

func (e *TestEnv) CreateChatbotWithConfig(user *models.User, name string, opts map[string]any) (*models.Chatbot, error) {
	id := uuid.New().String()

	var languageID string
	err := e.DB.QueryRow("SELECT id FROM languages WHERE code = 'en-US'").Scan(&languageID)
	if err != nil {
		return nil, fmt.Errorf("get language: %w", err)
	}

	bot := &models.Chatbot{
		ID:                   id,
		UserID:               user.ID,
		Name:                 name,
		Model:                policy.ModelGPT4oMini.String(),
		Temperature:          0.7,
		MaxTokens:            4096,
		ThemeColor:           "#000000",
		WelcomeMessage:       "Hello",
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
		Position:             "bottom-right",
		BotMessageColor:      "#ffffff",
		UserMessageColor:     "#000000",
		BotMessageTextColor:  "#000000",
		UserMessageTextColor: "#ffffff",
		ChatFontFamily:       "Inter",
		ChatHeaderColor:      "#ffffff",
		ChatHeaderTextColor:  "#000000",
		ChatBackgroundColor:  "#ffffff",
		BubbleRadius:         "12px",
		InputBackgroundColor: "#ffffff",
		InputTextColor:       "#000000",
		SendButtonColor:      "#000000",
		DiscoveryMode:        "auto",
		RefreshPolicy:        "manual",
		HideBranding:         false,
		ConfidenceThreshold:  0.5,
		HandoffEnabled:       false,
		HandoffType:          "email",
		LanguageCode:         "en-US",
		SecureEmbedEnabled:   false,
		SuggestionsEnabled:   false,
	}

	if dm, ok := opts["discovery_mode"].(string); ok {
		bot.DiscoveryMode = dm
	}
	if he, ok := opts["handoff_enabled"].(bool); ok {
		bot.HandoffEnabled = he
	}
	if ht, ok := opts["handoff_type"].(string); ok {
		bot.HandoffType = ht
	}

	_, err = e.DB.Exec(`INSERT INTO chatbots (
		id, user_id, name, model, temperature, max_tokens, 
		theme_color, welcome_message, created_at, updated_at, 
		position, bot_message_color, user_message_color, bot_message_text_color, user_message_text_color, 
		chat_font_family, chat_header_color, chat_header_text_color, chat_background_color, bubble_radius, 
		input_background_color, input_text_color, send_button_color, 
		discovery_mode, refresh_policy, hide_branding, confidence_threshold, handoff_enabled, handoff_type, 
		language_id, secure_embed_enabled, suggestions_enabled
	) VALUES (
		$1, $2, $3, $4, $5, $6, 
		$7, $8, $9, $10, 
		$11, $12, $13, $14, $15, 
		$16, $17, $18, $19, $20, 
		$21, $22, $23, 
		$24, $25, $26, $27, $28, $29, 
		$30, $31, $32
	)`,
		bot.ID, bot.UserID, bot.Name, bot.Model, bot.Temperature, bot.MaxTokens,
		bot.ThemeColor, bot.WelcomeMessage, bot.CreatedAt, bot.UpdatedAt,
		bot.Position, bot.BotMessageColor, bot.UserMessageColor, bot.BotMessageTextColor, bot.UserMessageTextColor,
		bot.ChatFontFamily, bot.ChatHeaderColor, bot.ChatHeaderTextColor, bot.ChatBackgroundColor, bot.BubbleRadius,
		bot.InputBackgroundColor, bot.InputTextColor, bot.SendButtonColor,
		bot.DiscoveryMode, bot.RefreshPolicy, bot.HideBranding, bot.ConfidenceThreshold, bot.HandoffEnabled, bot.HandoffType,
		languageID, bot.SecureEmbedEnabled, bot.SuggestionsEnabled,
	)
	if err != nil {
		return nil, fmt.Errorf("insert chatbot: %w", err)
	}

	return bot, nil
}
