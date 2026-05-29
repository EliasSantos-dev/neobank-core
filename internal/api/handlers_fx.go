package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/fx"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

func (h handler) listWallets(w http.ResponseWriter, r *http.Request) {
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	ws, err := h.s.ListWallets(r.Context(), u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ws)
}

func (h handler) createWallet(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Currency string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if !fxSupported(body.Currency) {
		writeErr(w, http.StatusBadRequest, "moeda não suportada")
		return
	}
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	if _, err := h.s.GetOrCreateWallet(r.Context(), u.ID, body.Currency); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"currency": body.Currency})
}

func (h handler) convert(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeErr(w, http.StatusBadRequest, "Idempotency-Key obrigatório")
		return
	}
	var body struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if !fxSupported(body.From) || !fxSupported(body.To) {
		writeErr(w, http.StatusBadRequest, "moeda não suportada")
		return
	}
	if body.From == body.To {
		writeErr(w, http.StatusBadRequest, "from e to devem diferir")
		return
	}
	rate, err := h.rates.Rate(r.Context(), body.From, body.To)
	if err != nil {
		if errors.Is(err, fx.ErrRateUnavailable) {
			writeErr(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	out, err := h.s.ConvertCurrency(r.Context(), store.ConvertParams{
		IdempotencyKey: key, UserID: u.ID, From: body.From, To: body.To, Amount: body.Amount, RateE8: rate,
	})
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	h.metrics.IncOperation("convert")
	writeJSON(w, http.StatusOK, map[string]int64{"converted": out.Converted, "rate_e8": out.RateE8})
}

func fxSupported(currency string) bool {
	switch currency {
	case "BRL", "USD", "EUR":
		return true
	default:
		return false
	}
}
