package server

import (
	"fmt"
	"net/http"

	"github.com/broothie/slink.chat/util"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func (s *Server) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(util.ContextLoggerMiddleware(s.Logger))
	r.Use(middleware.Recoverer)
	r.Use(csrf.Protect([]byte(s.Config.Secret)))

	r.Get("/", s.index)

	r.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.RequestLogger(util.NewChiLogFormatter(s.Config)))

		r.Route("/v1", func(r chi.Router) {
			r.With(s.requireUser).Get("/user", s.showCurrentUser)

			r.Route("/session", func(r chi.Router) {
				r.Post("/", s.createSession)
				r.Delete("/", s.destroySession)
			})

			r.Route("/users", func(r chi.Router) {
				r.Post("/", s.createUser)

				r.Group(func(r chi.Router) {
					r.Use(s.requireUser)

					r.Get("/", s.showUsers)
					r.Get("/search", s.searchUsers)

					r.Route("/{user_id}", func(r chi.Router) {
						r.Use(injectResourceIDLog("user"))

						r.Get("/", s.showUser)
					})
				})
			})

			r.Route("/channels", func(r chi.Router) {
				r.Use(s.requireUser)

				r.Get("/", s.indexChannels)
				r.Post("/", s.createChannel)
				r.Get("/search", s.searchChannels)

				r.Route("/chats", func(r chi.Router) {
					r.Post("/", s.upsertChat)
					r.Get("/messages", s.channelsSocket)
				})

				r.Route("/{channel_id}", func(r chi.Router) {
					r.Use(injectResourceIDLog("channel"))

					r.Get("/", s.showChannel)
					r.Post("/join", s.joinChannel)
					r.Delete("/leave", s.leaveChannel)
					r.Get("/users", s.indexChannelUsers)

					r.Route("/messages", func(r chi.Router) {
						r.Get("/", s.indexMessages)
						r.Get("/subscribe", s.channelSocket)
					})
				})
			})
		})
	})

	if s.Config.IsDevelopment() {
		chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			if handler != nil {
				fmt.Println(method, route)
			}

			return nil
		})
	}

	return r
}
