package handlers

import (
    "database/sql"
    "encoding/json"
    "net/http"
    "strings"

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
    BotIcon              *string  `json:"bot_icon"`
    BotDisplayName       *string  `json:"bot_display_name"`
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
            Model:                defaultString(req.Model, "gpt-3.5-turbo"),
            Temperature:          defaultFloat32(req.Temperature, 0.7),
            MaxTokens:            defaultInt(req.MaxTokens, 512),
            ThemeColor:           defaultString(req.ThemeColor, "#3b82f6"),
            WelcomeMessage:       defaultString(req.WelcomeMessage, "Merhaba! Size nasıl yardımcı olabilirim?"),
            Position:             defaultString(req.Position, "bottom-right"),
            BotMessageColor:      defaultString(req.BotMessageColor, "#3b82f6"),
            UserMessageColor:     defaultString(req.UserMessageColor, "#f3f4f6"),
            BotMessageTextColor:  defaultString(req.BotMessageTextColor, "#ffffff"),
            UserMessageTextColor: defaultString(req.UserMessageTextColor, "#1f2937"),
            ChatFontFamily:       defaultString(req.ChatFontFamily, "Inter, sans-serif"),
            ChatHeaderColor:      defaultString(req.ChatHeaderColor, "#3b82f6"),
            ChatHeaderTextColor:  defaultString(req.ChatHeaderTextColor, "#ffffff"),
            BotIcon:              req.BotIcon,
            BotDisplayName:       req.BotDisplayName,
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
        if req.Name != "" {
            c.Name = strings.TrimSpace(req.Name)
        }
        if req.Description != nil {
            c.Description = req.Description
        }
        if req.SystemPrompt != nil {
            c.SystemPrompt = *req.SystemPrompt
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
        if req.BotIcon != nil {
            c.BotIcon = req.BotIcon
        }
        if req.BotDisplayName != nil {
            c.BotDisplayName = req.BotDisplayName
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
