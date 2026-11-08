package httpx

import (
	"log/slog"
	"net/http"
	"time"
)

// Logging is a middleware that logs the method, path and duration of every request.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		slog.Info("http request", "method", r.Method, "path", r.URL.Path, "took", time.Since(start))
	})
}
