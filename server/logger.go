package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func (s *Server) loggerInjector(next http.Handler) http.Handler {
	logger, err := s.cfg.NewLogger()
	if err != nil {
		logger = zap.NewExample()
		logger.Error("failed to create logger", zap.Error(err))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := []zap.Field{
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("request_id", middleware.GetReqID(r.Context())),
			zap.String("remote_addr", r.RemoteAddr),
			zap.Any("httpRequest", util.Map{
				"requestMethod": r.Method,
				"requestUrl":    r.URL.String(),
				"userAgent":     r.UserAgent(),
				"referer":       r.Referer(),
			}),
		}

		if traceParent := r.Header.Get("Traceparent"); traceParent != "" {
			fields = append(fields, zap.String("traceparent", traceParent))

			sections := strings.Split(traceParent, "-")

			if traceID, err := trace.TraceIDFromHex(sections[1]); err != nil {
				logger.Error("failed to parse trace id", zap.Error(err))
			} else {
				fields = append(fields, zap.String("trace", fmt.Sprintf("projects/%s/traces/%s", s.cfg.ProjectID, traceID.String())))
			}

			if spanID, err := trace.SpanIDFromHex(sections[2]); err != nil {
				logger.Error("failed to parse span id", zap.Error(err))
			} else {
				fields = append(fields, zap.String("spanId", spanID.String()))
			}
		}

		next.ServeHTTP(w, r.WithContext(ctxzap.ToContext(r.Context(), logger.With(fields...))))
	})
}

func injectResourceIDLog(resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idName := fmt.Sprintf("%s_id", resource)
			ctxzap.AddFields(r.Context(), zap.String(idName, chi.URLParam(r, idName)))

			next.ServeHTTP(w, r)
		})
	}
}
