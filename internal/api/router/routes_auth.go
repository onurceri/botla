package router

import (
	"net/http"

	"github.com/onurceri/botla-app/internal/api/handlers"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func registerAuthRoutes(mux *http.ServeMux, ah *handlers.AuthHandlers, secret string) {
	mux.HandleFunc("/api/v1/auth/register", ah.RegisterHandler)
	mux.HandleFunc("/api/v1/auth/login", ah.LoginHandler)
	mux.HandleFunc("/api/v1/auth/refresh", ah.RefreshHandler)
	mux.HandleFunc("/api/v1/auth/logout", ah.LogoutHandler)
	mux.Handle("/api/v1/protected", middleware.AuthMiddleware(secret)(http.HandlerFunc(handlers.ProtectedHandler)))
}
