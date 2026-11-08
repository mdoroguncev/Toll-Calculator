// Package httpx contains HTTP helpers shared by the services:
// JSON responses, handlers that return errors, and request logging.
package httpx

import (
	"log/slog"
	"net/http"
)

// APIError is an error that is safe to show to the client:
// Handle responds with its Code and its message.
type APIError struct {
	Code int
	Err  error
}

func (e APIError) Error() string {
	return e.Err.Error()
}

// HandlerFunc is an HTTP handler that returns an error instead of writing it itself.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

// Handle converts fn into an http.HandlerFunc that writes returned errors as JSON.
// An APIError is sent with its own code and message. Any other error is logged
// and sent as 500 "internal server error", so internal details do not reach the client.
func Handle(fn HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			if apiErr, ok := err.(APIError); ok {
				WriteJSON(w, apiErr.Code, map[string]string{"error": apiErr.Error()})
			} else {
				slog.Error("internal error", "err", err, "method", r.Method, "path", r.URL.Path)
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}
	}
}
