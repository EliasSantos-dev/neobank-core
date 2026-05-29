package api

import (
	"encoding/json"
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/gateway"
	"github.com/google/uuid"
)

func (h handler) deposit(w http.ResponseWriter, r *http.Request)  { h.initiate(w, r, "deposit") }
func (h handler) withdraw(w http.ResponseWriter, r *http.Request) { h.initiate(w, r, "withdraw") }

func (h handler) initiate(w http.ResponseWriter, r *http.Request, kind string) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeErr(w, http.StatusBadRequest, "Idempotency-Key obrigatório")
		return
	}
	var req struct {
		Amount int64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		writeErr(w, http.StatusBadRequest, "valor inválido")
		return
	}
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	var p any
	if kind == "deposit" {
		p, err = h.gw.InitiateDeposit(r.Context(), u.WalletAccountID, req.Amount, key)
	} else {
		p, err = h.gw.InitiateWithdrawal(r.Context(), u.WalletAccountID, req.Amount, key)
	}
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	h.metrics.IncOperation(kind)
	writeJSON(w, http.StatusAccepted, p)
}

func (h handler) listPayments(w http.ResponseWriter, r *http.Request) {
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	items, err := h.s.ListPaymentsByAccount(r.Context(), u.WalletAccountID, 50, 0)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h handler) gatewayWebhook(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EventID   string `json:"event_id"`
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	eventID, err1 := uuid.Parse(body.EventID)
	paymentID, err2 := uuid.Parse(body.PaymentID)
	if err1 != nil || err2 != nil {
		writeErr(w, http.StatusBadRequest, "ids inválidos")
		return
	}
	if err := h.gw.HandleWebhook(r.Context(), gateway.WebhookEvent{
		EventID: eventID, PaymentID: paymentID, Status: body.Status,
	}); err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
