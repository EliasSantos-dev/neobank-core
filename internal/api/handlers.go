package api

import (
	"encoding/json"
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/google/uuid"
)

type handler struct{ s *store.Store }

type createAccountReq struct {
	Currency string `json:"currency"`
	Type     string `json:"type"`
}

func (h handler) createAccount(w http.ResponseWriter, r *http.Request) {
	var req createAccountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	acc, err := h.s.CreateAccount(r.Context(), req.Currency, ledger.AccountType(req.Type))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, acc)
}

type transferReq struct {
	FromAccountID uuid.UUID `json:"from_account_id"`
	ToAccountID   uuid.UUID `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
}

func (h handler) transfer(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeErr(w, http.StatusBadRequest, "Idempotency-Key obrigatório")
		return
	}
	var req transferReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	tr, err := h.s.Transfer(r.Context(), store.TransferParams{
		IdempotencyKey: key,
		FromAccountID:  req.FromAccountID,
		ToAccountID:    req.ToAccountID,
		Amount:         req.Amount,
		Currency:       req.Currency,
	})
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, tr)
}

func (h handler) balance(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	bal, err := h.s.Balance(r.Context(), id)
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]int64{"balance": bal})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
