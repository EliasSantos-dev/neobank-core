package api

import (
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

// NewServer monta o roteador (net/http, roteamento por método do Go 1.22+).
func NewServer(s *store.Store) http.Handler {
	h := handler{s: s}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /accounts", h.createAccount)
	mux.HandleFunc("POST /transfers", h.transfer)
	mux.HandleFunc("GET /accounts/{id}/balance", h.balance)
	return mux
}
