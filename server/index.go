package server

import (
	"net/http"

	"github.com/broothie/slink.chat/util"
	"github.com/gorilla/csrf"
)

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.render.HTML(w, http.StatusOK, "index", util.Map{
		"csrf_token":    csrf.Token(r),
		"is_production": s.Config.IsProduction(),
	})
}
