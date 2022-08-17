package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

func (s *Server) loggerInjector(next http.Handler) http.Handler {
	logger, err := s.cfg.NewLogger()
	if err != nil {
		logger = zap.NewExample()
		logger.Error("failed to create logger", zap.Error(err))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(ctxzap.ToContext(r.Context(), logger.With(
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", middleware.GetReqID(r.Context())),
			zap.String("remote_addr", r.RemoteAddr),
		))))
	})
}

func injectResourceLog(resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idName := fmt.Sprintf("%s_id", resource)
			ctxzap.AddFields(r.Context(), zap.String(idName, chi.URLParam(r, idName)))

			next.ServeHTTP(w, r)
		})
	}
}
