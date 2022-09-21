package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

func injectResourceIDLog(resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idName := fmt.Sprintf("%s_id", resource)
			ctxzap.AddFields(r.Context(), zap.String(idName, chi.URLParam(r, idName)))

			next.ServeHTTP(w, r)
		})
	}
}
