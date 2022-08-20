package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
)

func (s *Server) routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(s.loggerInjector)
	r.Use(middleware.Recoverer)
	r.Use(csrf.Protect([]byte(s.cfg.Secret)))

	r.Get("/", s.index)

	r.Get("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))).ServeHTTP)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.Logger)

		r.Route("/v1", func(r chi.Router) {
			r.Route("/session", func(r chi.Router) {
				r.Post("/", s.createSession)
				r.Delete("/", s.destroySession)
			})

			r.Route("/user", func(r chi.Router) {
				r.With(s.requireUser).Get("/", s.showCurrentUser)
			})

			r.Route("/users", func(r chi.Router) {
				r.Post("/", s.createUser)

				r.Group(func(r chi.Router) {
					r.Use(s.requireUser)

					r.Get("/search", s.searchUsers)

					r.Route("/{user_id}", func(r chi.Router) {
						r.Use(injectResourceIDLog("user"))

						r.Get("/", s.showUser)
					})
				})
			})

			r.Group(func(r chi.Router) {
				r.Use(s.requireUser)

				r.Post("/chats", s.createChat)

				r.Route("/channels", func(r chi.Router) {
					r.Get("/", s.indexChannels)
					r.Post("/", s.createChannel)
					r.Get("/search", s.searchChannels)

					r.Route("/{channel_id}", func(r chi.Router) {
						r.Use(injectResourceIDLog("channel"))

						r.Get("/", s.showChannel)
						r.Delete("/", s.destroySubscription)

						r.Route("/subscriptions", func(r chi.Router) {
							r.Post("/", s.createSubscription)
						})

						r.Route("/messages", func(r chi.Router) {
							r.Get("/", s.indexMessages)
							r.Post("/", s.createMessage)
							r.Get("/subscribe", s.channelSocket)
						})
					})
				})
			})
		})
	})

	if s.cfg.IsDevelopment() {
		chi.Walk(r, func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			if handler != nil {
				fmt.Println(method, route)
			}

			return nil
		})
	}

	return r
}
