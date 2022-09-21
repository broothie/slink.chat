package job

import (
	"context"
	"net/http"

	"github.com/broothie/slink.chat/async"
	"github.com/broothie/slink.chat/core"
	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Server struct {
	core.Core
}

func NewServer(core core.Core) *Server {
	return &Server{Core: core}
}

func (s *Server) Handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(util.ContextLoggerMiddleware(s.Logger))
	r.Use(middleware.RequestLogger(util.NewChiLogFormatter(s.Config)))
	r.Use(middleware.Recoverer)

	r.Handle("/", s.Async.Handler(s.Dispatch))

	return r
}

func (s *Server) Dispatch(ctx context.Context, message async.Message) error {
	if err := s.dispatch(ctx, message); err != nil {
		s.Logger.Error("failed to dispatch job", zap.Error(err))
		return errors.Wrap(err, "failed to dispatch job")
	}

	return nil
}
