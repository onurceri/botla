package router

import (
	"net/http"
	"strings"

	"github.com/onurceri/botla-app/internal/api/handlers"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func registerPublicRoutes(mux *http.ServeMux, secret string, hoh *handlers.HandoffHandlers, ph *handlers.PublicHandlers, chatbotRepo repository.ChatbotRepository) {
	mux.Handle("/api/v1/public/chatbots/", middleware.OptionalAuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const p = "/api/v1/public/chatbots/"
		path := r.URL.Path

		// Public email submission
		if strings.HasPrefix(path, p) && strings.Contains(path, "/handoff/") && strings.HasSuffix(path, "/contact") {
			hoh.PublicSubmitEmail(w, r)
			return
		}
		// Public handoff request
		if strings.HasPrefix(path, p) && strings.HasSuffix(path, "/handoff") {
			hoh.PublicRequestHandoff(w, r)
			return
		}
		// Public chat
		if strings.HasPrefix(path, p) && strings.HasSuffix(path, "/chat") {
			ph.PublicChat(w, r)
			return
		}
		// Feedback
		if strings.HasPrefix(path, p) && strings.HasSuffix(path, "/feedback") {
			ph.SubmitFeedback(w, r)
			return
		}
		// Config
		handlers.PublicChatbotConfig(chatbotRepo)(w, r)
	})))
}
