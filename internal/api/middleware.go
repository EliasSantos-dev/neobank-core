package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/EliasSantos-dev/neobank-core/internal/auth"
	"github.com/google/uuid"
)

type ctxKey int

const userIDKey ctxKey = 0

// requireAuth valida o JWT do header Authorization e injeta o user id no contexto.
func (h handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || token == "" {
			writeErr(w, http.StatusUnauthorized, "token ausente")
			return
		}
		id, err := auth.Parse(h.secret, token)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "token inválido")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userIDKey, id)))
	}
}

func userID(r *http.Request) uuid.UUID {
	id, _ := r.Context().Value(userIDKey).(uuid.UUID)
	return id
}
