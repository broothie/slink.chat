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
			r.Post("/users", s.createUser)
			r.Post("/sessions", s.createSession)

			r.Group(func(r chi.Router) {
				r.Use(s.requireUser)

				r.Get("/user", s.showUser)

				r.Route("/channels", func(r chi.Router) {
					r.Post("/", s.createChannel)

					r.Route("/{channel_id}", func(r chi.Router) {
						r.Use(injectResourceLog("channel"))

						r.Get("/subscribe", s.channelSubscribe)

						r.Route("/messages", func(r chi.Router) {
							r.Get("/", s.indexMessages)
							r.Post("/", s.createMessage)
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
