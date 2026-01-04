package router

import (
	"net/http"

	"github.com/onurceri/botla-app/internal/api/handlers"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func RegisterSourceRoutes(mux *http.ServeMux, secret string, sh *handlers.SourcesHandlers, tjh *handlers.TrainingJobHandlers) {
	auth := middleware.AuthMiddleware(secret)

	mux.Handle("POST /api/v1/sources/{id}/refresh", auth(http.HandlerFunc(sh.RefreshSource)))
	mux.Handle("GET /api/v1/sources/{id}/chunks", auth(http.HandlerFunc(sh.GetSourceChunks)))
	mux.Handle("GET /api/v1/sources/{id}/job", auth(http.HandlerFunc(tjh.GetJobStatus)))
	mux.Handle("POST /api/v1/sources/{id}/job/retry", auth(http.HandlerFunc(tjh.RetryJob)))
	mux.Handle("GET /api/v1/sources/{id}", auth(http.HandlerFunc(sh.GetSourceStatusOrDelete)))
	mux.Handle("DELETE /api/v1/sources/{id}", auth(http.HandlerFunc(sh.GetSourceStatusOrDelete)))
}
