package server

import (
	"net/http"

	"github.com/broothie/slink.chat/core"
	"github.com/gorilla/sessions"
	"github.com/unrolled/render"
)

type Server struct {
	core.Core
	sessions *sessions.CookieStore
	render   *render.Render
}

func New(core core.Core) (*Server, error) {
	return &Server{
		Core:     core,
		sessions: sessions.NewCookieStore([]byte(core.Config.Secret)),
		render: render.New(render.Options{
			IndentJSON:                  core.Config.IsLocal(),
			IsDevelopment:               core.Config.IsLocal(),
			Layout:                      "layout",
			RenderPartialsWithoutPrefix: true,
			StreamingJSON:               true,
		}),
	}, nil
}

func (s *Server) Handler() http.Handler {
	return s.routes()
}
