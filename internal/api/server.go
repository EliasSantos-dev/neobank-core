package api

import (
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

// NewServer monta o roteador da API de produto (net/http, Go 1.22+).
func NewServer(s *store.Store, secret []byte) http.Handler {
	h := handler{s: s, secret: secret}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("GET /me", h.requireAuth(h.me))
	mux.HandleFunc("POST /me/deposit", h.requireAuth(h.deposit))
	mux.HandleFunc("POST /me/withdraw", h.requireAuth(h.withdraw))
	mux.HandleFunc("POST /transfers", h.requireAuth(h.transferToUser))
	mux.HandleFunc("GET /me/statement", h.requireAuth(h.statement))
	return mux
}
