package router

import (
	"net/http"
	"strings"

	"github.com/onurceri/botla-co/internal/api/handlers"
	"github.com/onurceri/botla-co/pkg/middleware"
)

func registerSourceRoutes(mux *http.ServeMux, secret string, sh *handlers.SourcesHandlers) {
	mux.Handle("/api/v1/sources/", middleware.AuthMiddleware(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/refresh") {
			sh.RefreshSource(w, r)
			return
		}
		sh.GetSourceStatusOrDelete(w, r)
	})))
}
