package server

import (
	"context"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/config"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/unrolled/render"
)

type Server struct {
	cfg      *config.Config
	sessions *sessions.CookieStore
	db       *firestore.Client
	render   *render.Render
}

func New(cfg *config.Config) (*Server, error) {
	client, err := firestore.NewClient(context.Background(), cfg.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new firestore client")
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

func (s *Server) Collection(name string) *firestore.CollectionRef {
	return s.db.Collection(s.cfg.Collection(name))
}
