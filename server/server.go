package server

import (
	"net/http"

	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/db"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/unrolled/render"
)

type Server struct {
	cfg      *config.Config
	sessions *sessions.CookieStore
	db       *db.DB
	render   *render.Render
}

func New(cfg *config.Config) (*Server, error) {
	client, err := db.New(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new db")
	}

	return &Server{
		cfg:      cfg,
		sessions: sessions.NewCookieStore([]byte(cfg.Secret)),
		db:       client,
		render: render.New(render.Options{
			IndentJSON:                  cfg.IsLocal(),
			IsDevelopment:               cfg.IsLocal(),
			Layout:                      "layout",
			RenderPartialsWithoutPrefix: true,
			StreamingJSON:               true,
		}),
	}, nil
}

func (s *Server) Handler() http.Handler {
	return s.routes()
}
