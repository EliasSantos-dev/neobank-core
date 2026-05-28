package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/google/uuid"
)

type amountReq struct {
	Amount int64 `json:"amount"`
}

func (h handler) deposit(w http.ResponseWriter, r *http.Request) {
	var req amountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	h.doTransfer(w, r, ledger.TreasuryBRL, u.WalletAccountID, req.Amount)
}

func (h handler) withdraw(w http.ResponseWriter, r *http.Request) {
	var req amountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	h.doTransfer(w, r, u.WalletAccountID, ledger.TreasuryBRL, req.Amount)
}

// doTransfer encapsula a chamada ao ledger com a Idempotency-Key do header.
func (h handler) doTransfer(w http.ResponseWriter, r *http.Request, from, to uuid.UUID, amount int64) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeErr(w, http.StatusBadRequest, "Idempotency-Key obrigatório")
		return
	}
	tr, err := h.s.Transfer(r.Context(), store.TransferParams{
		IdempotencyKey: key, FromAccountID: from, ToAccountID: to, Amount: amount, Currency: "BRL",
	})
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, tr)
}

func (h handler) statement(w http.ResponseWriter, r *http.Request) {
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	limit := atoiDefault(r.URL.Query().Get("limit"), 50)
	offset := atoiDefault(r.URL.Query().Get("offset"), 0)
	entries, err := h.s.ListEntries(r.Context(), u.WalletAccountID, int32(limit), int32(offset))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func atoiDefault(s string, def int) int {
	if n, err := strconv.Atoi(s); err == nil && n >= 0 {
		return n
	}
	return def
}
