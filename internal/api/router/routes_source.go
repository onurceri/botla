package router

import (
	"net/http"

	"github.com/onurceri/botla-app/internal/api/handlers"
	"github.com/onurceri/botla-app/internal/repository"
	"github.com/onurceri/botla-app/pkg/logger"
	"github.com/onurceri/botla-app/pkg/middleware"
)

func RegisterSourceRoutes(mux *http.ServeMux, secret string, userRepo *repository.PostgresUserRepo, log *logger.Logger, sh *handlers.SourcesHandlers, tjh *handlers.TrainingJobHandlers) {
	auth := middleware.AuthMiddleware(secret)
	deletedCheck := middleware.DeletedAccountMiddleware(userRepo, log)

	mux.Handle("POST /api/v1/sources/{id}/refresh", auth(deletedCheck(http.HandlerFunc(sh.RefreshSource))))
	mux.Handle("GET /api/v1/sources/{id}/chunks", auth(deletedCheck(http.HandlerFunc(sh.GetSourceChunks))))
	mux.Handle("GET /api/v1/sources/{id}/job", auth(deletedCheck(http.HandlerFunc(tjh.GetJobStatus))))
	mux.Handle("POST /api/v1/sources/{id}/job/retry", auth(deletedCheck(http.HandlerFunc(tjh.RetryJob))))
	mux.Handle("GET /api/v1/sources/{id}", auth(deletedCheck(http.HandlerFunc(sh.GetSourceStatusOrDelete))))
	mux.Handle("DELETE /api/v1/sources/{id}", auth(deletedCheck(http.HandlerFunc(sh.GetSourceStatusOrDelete))))
}
