package middleware

import (
	"log/slog"
	"net/http"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					"error", err,
					"method", r.Method,
					"path", r.URL.Path,
				)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				if _, err := w.Write([]byte(`{"error":"Internal server error"}`)); err != nil {
					slog.Error("failed to write response", "error", err)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}