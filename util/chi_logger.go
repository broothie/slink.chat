package util

import (
	"fmt"
	"net/http"
	"time"

	"github.com/broothie/slink.chat/config"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

func NewChiLogFormatter(cfg *config.Config) ChiLogFormatter {
	return ChiLogFormatter{config: cfg}
}

type ChiLogFormatter struct {
	config *config.Config
}

func (c ChiLogFormatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return chiLogEntry{request: r, config: c.config}
}

type chiLogEntry struct {
	request *http.Request
	config  *config.Config
}

func (e chiLogEntry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ any) {
	fields := []zap.Field{
		zap.String("bytes", fmt.Sprintf("%dB", bytes)),
		zap.Duration("elapsed", elapsed),
		zap.Int("status", status),
	}

	if e.config.IsHosted() {
		fields = append(fields,
			zap.Any("httpRequest", Map{
				"requestMethod": e.request.Method,
				"requestUrl":    e.request.URL.String(),
				"status":        status,
			}),
		)
	}

	ctxzap.Extract(e.request.Context()).Info(fmt.Sprintf("%s %s", e.request.Method, e.request.URL.Path), fields...)
}

func (e chiLogEntry) Panic(value any, stack []byte) {
	ctxzap.Extract(e.request.Context()).Error("request logger panic",
		zap.Any("value", value),
		zap.String("stack", string(stack)),
	)
}
