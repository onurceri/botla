package middleware

import (
    "net/http"
    "strings"
)

func CORSMiddleware(origin string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Vary", "Origin")
            w.Header().Set("Access-Control-Allow-Credentials", "true")
            w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

func CORSMiddlewareAllowOrigins(origins []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            o := r.Header.Get("Origin")
            allow := false
            if o != "" {
                for _, allowed := range origins {
                    if strings.TrimSpace(allowed) == o {
                        allow = true
                        break
                    }
                }
            }
            if allow {
                w.Header().Set("Access-Control-Allow-Origin", o)
                w.Header().Set("Vary", "Origin")
                w.Header().Set("Access-Control-Allow-Credentials", "true")
                w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
                w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            }
            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
