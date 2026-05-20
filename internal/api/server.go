package api

import (
	"net/http"

	"github.com/dacolabs/daco/internal/api/handlers"
	"github.com/dacolabs/daco/internal/api/middleware"
)

func NewServer() http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux)

	var handler http.Handler = mux
	handler = middleware.Logging(handler)
	handler = middleware.Recovery(handler)
	return handler
}

func addRoutes(
	mux *http.ServeMux,
) {
	mux.Handle("/health", handlers.Health())
}
