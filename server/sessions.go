package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/model"
	"github.com/broothie/slink.chat/util"
	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const authSessionName = "auth"

func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	var params userParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		logger.Error("failed to decode body", zap.Error(err))
		s.render.JSON(w, http.StatusBadRequest, errorMap(err))
		return
	}

	user, err := db.NewFetcher[model.User](s.DB).FetchFirst(r.Context(), func(query *firestore.CollectionRef) firestore.Query {
		return query.Where("screenname", "==", params.Screenname)
	})
	if err != nil {
		if err == db.NotFound {
			logger.Error("user not found", zap.Error(err))
			s.render.JSON(w, http.StatusUnauthorized, errorMap(errors.New("invalid screenname/password combination")))
			return
		}

		logger.Error("failed to search for user", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	if passwordsMatch, err := user.PasswordsMatch(params.Password); err != nil {
		logger.Error("failed to compare passwords", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	} else if !passwordsMatch {
		logger.Error("failed to decode body", zap.Error(err))
		s.render.JSON(w, http.StatusUnauthorized, errorMap(errors.New("invalid screenname/password combination")))
		return
	}

	jwt, err := s.newJWTToken(user.ID)
	if err != nil {
		logger.Error("failed to create jwt", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	session, _ := s.sessions.Get(r, authSessionName)
	session.Values["jwt"] = jwt
	if err := session.Save(r, w); err != nil {
		logger.Error("failed to save session", zap.Error(err))
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusOK, util.Map{"user": user})
}

func (s *Server) destroySession(w http.ResponseWriter, r *http.Request) {
	logger := ctxzap.Extract(r.Context())

	authSession, _ := s.sessions.Get(r, authSessionName)
	authSession.Values = nil
	if err := authSession.Save(r, w); err != nil {
		logger.Info("failed to save auth session")
		s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
		return
	}

	s.render.JSON(w, http.StatusNoContent, nil)
}

func (s *Server) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := ctxzap.Extract(r.Context())

		authSession, _ := s.sessions.Get(r, authSessionName)
		tokenValue, ok := authSession.Values["jwt"]
		if !ok {
			logger.Info("no jwt on session")
			s.render.JSON(w, http.StatusUnauthorized, errorMap(errors.New("no jwt on session")))
			return
		}

		token, err := jwt.Parse(tokenValue.(string), func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(s.Config.Secret), nil
		})
		if err != nil {
			logger.Error("jwt parse error", zap.Error(err))
			s.render.JSON(w, http.StatusUnauthorized, errorMap(err))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			logger.Info("invalid jwt token")
			s.render.JSON(w, http.StatusUnauthorized, errorMap(errors.New("invalid token claims")))
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			logger.Error("no user_id in claims")
			s.render.JSON(w, http.StatusUnauthorized, errorMap(errors.New("no user_id in claims")))
			return
		}

		user, err := db.NewFetcher[model.User](s.DB).Fetch(r.Context(), userID)
		if err != nil {
			logger.Error("failed to get user from db", zap.Error(err))
			s.render.JSON(w, http.StatusInternalServerError, errorMap(err))
			return
		}

		ctxzap.AddFields(r.Context(), zap.Any("user_id", userID))
		next.ServeHTTP(w, r.WithContext(user.OnContext(r.Context())))
	})
}

func (s *Server) newJWTToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID})
	tokenString, err := token.SignedString([]byte(s.Config.Secret))
	if err != nil {
		return "", errors.Wrap(err, "failed to sign JWT")
	}

	return tokenString, nil
}
