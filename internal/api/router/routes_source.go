package router

import (
	"net/http"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func RegisterSourceRoutes(mux *http.ServeMux, secret string, sh *handlers.SourcesHandlers) {
	auth := middleware.AuthMiddleware(secret)

	mux.Handle("POST /api/v1/sources/{id}/refresh", auth(http.HandlerFunc(sh.RefreshSource)))
	mux.Handle("GET /api/v1/sources/{id}/chunks", auth(http.HandlerFunc(sh.GetSourceChunks)))
	mux.Handle("GET /api/v1/sources/{id}", auth(http.HandlerFunc(sh.GetSourceStatusOrDelete)))
	mux.Handle("DELETE /api/v1/sources/{id}", auth(http.HandlerFunc(sh.GetSourceStatusOrDelete)))
}
