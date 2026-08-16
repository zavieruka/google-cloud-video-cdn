package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
)

func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware allows browser clients from the configured origins to call the
// API. A single "*" entry allows any origin; otherwise the request's Origin is
// reflected only when it is on the allowlist. Preflight (OPTIONS) requests are
// answered here so they never fall through to the router.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(allowedOrigins))
	wildcard := false
	for _, origin := range allowedOrigins {
		if origin == "*" {
			wildcard = true
		}
		allowed[origin] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				switch {
				case wildcard:
					w.Header().Set("Access-Control-Allow-Origin", "*")
				case allowed[origin]:
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
				}

				if _, ok := w.Header()["Access-Control-Allow-Origin"]; ok {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					w.Header().Set("Access-Control-Max-Age", "3600")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
