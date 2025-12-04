package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strings"
    "regexp"

    "github.com/onurceri/botla-co/internal/db"
    "github.com/onurceri/botla-co/internal/models"
    "github.com/onurceri/botla-co/pkg/middleware"
)

type ChatbotHandlers struct {
    DB *sql.DB
}

type createChatbotRequest struct {
    Name                 string   `json:"name"`
    Description          *string  `json:"description"`
    SystemPrompt         *string  `json:"system_prompt"`
    Language             *string  `json:"language"`
    Model                *string  `json:"model"`
    Temperature          *float32 `json:"temperature"`
    MaxTokens            *int     `json:"max_tokens"`
    ThemeColor           *string  `json:"theme_color"`
    WelcomeMessage       *string  `json:"welcome_message"`
    Position             *string  `json:"position"`
    BotMessageColor      *string  `json:"bot_message_color"`
    UserMessageColor     *string  `json:"user_message_color"`
    BotMessageTextColor  *string  `json:"bot_message_text_color"`
    UserMessageTextColor *string  `json:"user_message_text_color"`
    ChatFontFamily       *string  `json:"chat_font_family"`
    ChatHeaderColor      *string  `json:"chat_header_color"`
    ChatHeaderTextColor  *string  `json:"chat_header_text_color"`
    ChatBackgroundColor  *string  `json:"chat_background_color"`
    BotIcon              *string  `json:"bot_icon"`
    BotDisplayName       *string  `json:"bot_display_name"`
    SecureEmbedEnabled   *bool    `json:"secure_embed_enabled"`
    AllowedDomains       *string  `json:"allowed_domains"`
    EmbedSecret          *string  `json:"embed_secret"`
    SuggestedQuestions   *[]string `json:"suggested_questions"`
    SuggestionsEnabled   *bool    `json:"suggestions_enabled"`
}

func (h *ChatbotHandlers) ListOrCreate(w http.ResponseWriter, r *http.Request) {
    id, ok := middleware.UserIDFromContext(r.Context())
    if !ok || id == "" {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    switch r.Method {
    case http.MethodGet:
        bots, err := db.GetChatbotsByUserID(r.Context(), h.DB, id)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(bots)
    case http.MethodPost:
        var req createChatbotRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        req.Name = strings.TrimSpace(req.Name)
        if req.Name == "" {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        bot := &models.Chatbot{
            UserID:               id,
            Name:                 req.Name,
            Description:          req.Description,
            SystemPrompt:         defaultString(req.SystemPrompt, "Sen yararlı, kibar ve bilgili bir yapay zeka asistanısın."),
            Language:             defaultString(req.Language, "tr"),
            Model:                defaultString(req.Model, "gpt-3.5-turbo"),
            Temperature:          defaultFloat32(req.Temperature, 0.7),
            MaxTokens:            defaultInt(req.MaxTokens, 512),
            ThemeColor:           defaultString(req.ThemeColor, "#3b82f6"),
            WelcomeMessage:       defaultString(req.WelcomeMessage, "Merhaba! Size nasıl yardımcı olabilirim?"),
            Position:             defaultString(req.Position, "bottom-right"),
            BotMessageColor:      defaultString(req.BotMessageColor, "#fcfcfd"),
            UserMessageColor:     defaultString(req.UserMessageColor, "#2e408a"),
            BotMessageTextColor:  defaultString(req.BotMessageTextColor, "#030303"),
            UserMessageTextColor: defaultString(req.UserMessageTextColor, "#ffffff"),
            ChatFontFamily:       defaultString(req.ChatFontFamily, "Inter, sans-serif"),
            ChatHeaderColor:      defaultString(req.ChatHeaderColor, "#3b82f6"),
            ChatHeaderTextColor:  defaultString(req.ChatHeaderTextColor, "#ffffff"),
            ChatBackgroundColor:  defaultString(req.ChatBackgroundColor, "#fff5e6"),
            BotIcon:              req.BotIcon,
            BotDisplayName:       req.BotDisplayName,
            SecureEmbedEnabled:   func() bool { if req.SecureEmbedEnabled != nil { return *req.SecureEmbedEnabled }; return false }(),
            AllowedDomains:       req.AllowedDomains,
            EmbedSecret:          req.EmbedSecret,
            SuggestedQuestions:   func() []string { if req.SuggestedQuestions != nil { return normalizeSuggestions(*req.SuggestedQuestions) }; return nil }(),
            SuggestionsEnabled:   func() bool { if req.SuggestionsEnabled != nil { return *req.SuggestionsEnabled }; return false }(),
        }
        newID, err := db.CreateChatbot(r.Context(), h.DB, bot)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        c, err := db.GetChatbotByID(r.Context(), h.DB, newID)
        if err != nil || c == nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(c)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func (h *ChatbotHandlers) ByID(w http.ResponseWriter, r *http.Request) {
    id, ok := middleware.UserIDFromContext(r.Context())
    if !ok || id == "" {
        w.WriteHeader(http.StatusUnauthorized)
        return
    }
    const prefix = "/api/v1/chatbots/"
    path := r.URL.Path
    if !strings.HasPrefix(path, prefix) {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    botID := strings.TrimPrefix(path, prefix)
    if botID == "" {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    if botID == "new" {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    c, err := db.GetChatbotByID(r.Context(), h.DB, botID)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    if c == nil {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    if c.UserID != id {
        w.WriteHeader(http.StatusForbidden)
        return
    }
    switch r.Method {
    case http.MethodGet:
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(c)
    case http.MethodPut:
        var req createChatbotRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            w.WriteHeader(http.StatusBadRequest)
            return
        }
        if req.ChatBackgroundColor != nil {
            s := strings.TrimSpace(*req.ChatBackgroundColor)
            if s != "" && !isValidHexColor(s) {
                w.WriteHeader(http.StatusBadRequest)
                return
            }
        }
        if req.Name != "" {
            c.Name = strings.TrimSpace(req.Name)
        }
        if req.Description != nil {
            c.Description = req.Description
        }
        if req.SystemPrompt != nil {
            c.SystemPrompt = *req.SystemPrompt
        }
        if req.Language != nil {
            c.Language = *req.Language
        }
        if req.Model != nil {
            c.Model = *req.Model
        }
        if req.Temperature != nil {
            c.Temperature = *req.Temperature
        }
        if req.MaxTokens != nil {
            c.MaxTokens = *req.MaxTokens
        }
        if req.ThemeColor != nil {
            c.ThemeColor = *req.ThemeColor
        }
        if req.WelcomeMessage != nil {
            c.WelcomeMessage = *req.WelcomeMessage
        }
        if req.Position != nil {
            c.Position = *req.Position
        }
        if req.BotMessageColor != nil {
            c.BotMessageColor = *req.BotMessageColor
        }
        if req.UserMessageColor != nil {
            c.UserMessageColor = *req.UserMessageColor
        }
        if req.BotMessageTextColor != nil {
            c.BotMessageTextColor = *req.BotMessageTextColor
        }
        if req.UserMessageTextColor != nil {
            c.UserMessageTextColor = *req.UserMessageTextColor
        }
        if req.ChatFontFamily != nil {
            c.ChatFontFamily = *req.ChatFontFamily
        }
        if req.ChatHeaderColor != nil {
            c.ChatHeaderColor = *req.ChatHeaderColor
        }
        if req.ChatHeaderTextColor != nil {
            c.ChatHeaderTextColor = *req.ChatHeaderTextColor
        }
        if req.ChatBackgroundColor != nil {
            c.ChatBackgroundColor = *req.ChatBackgroundColor
        }
        if req.BotIcon != nil {
            c.BotIcon = req.BotIcon
        }
        if req.BotDisplayName != nil {
            c.BotDisplayName = req.BotDisplayName
        }
        if req.SecureEmbedEnabled != nil {
            c.SecureEmbedEnabled = *req.SecureEmbedEnabled
        }
        if req.AllowedDomains != nil {
            c.AllowedDomains = req.AllowedDomains
        }
        if req.EmbedSecret != nil {
            c.EmbedSecret = req.EmbedSecret
        }
        if req.SuggestedQuestions != nil {
            c.SuggestedQuestions = normalizeSuggestions(*req.SuggestedQuestions)
        }
        if req.SuggestionsEnabled != nil {
            c.SuggestionsEnabled = *req.SuggestionsEnabled
        }
        err = db.UpdateChatbot(r.Context(), h.DB, c)
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        // Re-read for updated_at change
        c2, err := db.GetChatbotByID(r.Context(), h.DB, botID)
        if err != nil || c2 == nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(c2)
    case http.MethodDelete:
        if err := db.SoftDeleteChatbot(r.Context(), h.DB, botID, id); err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusNoContent)
    default:
        w.WriteHeader(http.StatusMethodNotAllowed)
    }
}

func normalizeSuggestions(in []string) []string {
    if len(in) == 0 { return []string{} }
    out := make([]string, 0, len(in))
    seen := map[string]struct{}{}
    for _, q := range in {
        t := strings.TrimSpace(q)
        if t == "" { continue }
        if len(t) > 120 { t = t[:120] }
        k := strings.ToLower(t)
        if _, ok := seen[k]; ok { continue }
        seen[k] = struct{}{}
        out = append(out, t)
        if len(out) >= 6 { break }
    }
    return out
}

var hexColorRe = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
func isValidHexColor(s string) bool { return hexColorRe.MatchString(s) }

func defaultString(p *string, d string) string {
    if p != nil {
        s := strings.TrimSpace(*p)
        if s != "" {
            return s
        }
    }
    return d
}

func defaultInt(p *int, d int) int {
    if p != nil {
        return *p
    }
    return d
}

func defaultFloat32(p *float32, d float32) float32 {
    if p != nil {
        return *p
    }
    return d
}
