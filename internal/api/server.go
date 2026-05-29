package api

import (
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/fx"
	"github.com/EliasSantos-dev/neobank-core/internal/gateway"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

// NewServer monta o roteador da API de produto (net/http, Go 1.22+).
func NewServer(s *store.Store, secret []byte, adminToken string, gw *gateway.Service, rates fx.RateProvider) http.Handler {
	h := handler{s: s, secret: secret, adminToken: adminToken, gw: gw, rates: rates}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
	mux.HandleFunc("GET /me", h.requireAuth(h.me))
	mux.HandleFunc("POST /me/deposit", h.requireAuth(h.deposit))
	mux.HandleFunc("POST /me/withdraw", h.requireAuth(h.withdraw))
	mux.HandleFunc("GET /me/payments", h.requireAuth(h.listPayments))
	mux.HandleFunc("GET /me/wallets", h.requireAuth(h.listWallets))
	mux.HandleFunc("POST /me/wallets", h.requireAuth(h.createWallet))
	mux.HandleFunc("POST /me/convert", h.requireAuth(h.convert))
	mux.HandleFunc("POST /transfers", h.requireAuth(h.createTransferIntent))
	mux.HandleFunc("GET /me/transfers", h.requireAuth(h.listMyIntents))
	mux.HandleFunc("GET /me/statement", h.requireAuth(h.statement))
	mux.HandleFunc("POST /webhooks/gateway", h.gatewayWebhook)
	mux.HandleFunc("POST /admin/transfers/{id}/review", h.requireAdmin(h.reviewIntent))
	mux.HandleFunc("POST /admin/reconcile", h.requireAdmin(h.reconcile))
	return mux
}
