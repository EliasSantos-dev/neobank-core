package api

import (
	"encoding/json"
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/google/uuid"
)

type intentReq struct {
	ToEmail string `json:"to_email"`
	Amount  int64  `json:"amount"`
}

// createTransferIntent registra uma intenção de transferência (assíncrona):
// o worker a pontua e decide efetivar ou enviar para revisão.
func (h handler) createTransferIntent(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		writeErr(w, http.StatusBadRequest, "Idempotency-Key obrigatório")
		return
	}
	var req intentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if req.Amount <= 0 {
		writeErr(w, http.StatusBadRequest, "valor inválido")
		return
	}
	sender, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	recipient, _, err := h.s.GetUserByEmail(r.Context(), req.ToEmail)
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	it, err := h.s.CreateIntent(r.Context(), store.IntentParams{
		IdempotencyKey: key, FromAccountID: sender.WalletAccountID,
		ToAccountID: recipient.WalletAccountID, Amount: req.Amount, Currency: "BRL",
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.metrics.IncOperation("transfer")
	writeJSON(w, http.StatusAccepted, map[string]any{"intent_id": it.ID, "status": it.Status})
}

func (h handler) listMyIntents(w http.ResponseWriter, r *http.Request) {
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	items, err := h.s.ListIntentsByAccount(r.Context(), u.WalletAccountID, 50, 0)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

type reviewReq struct {
	Decision string `json:"decision"` // approve | reject
}

func (h handler) reviewIntent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id inválido")
		return
	}
	it, err := h.s.GetIntent(r.Context(), id)
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	if it.Status != "under_review" {
		writeErr(w, http.StatusConflict, "intenção não está em revisão")
		return
	}
	var req reviewReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	switch req.Decision {
	case "approve":
		if err := h.s.CompleteIntent(r.Context(), id); err != nil {
			writeErr(w, statusFor(err), err.Error())
			return
		}
	case "reject":
		if err := h.s.RejectIntent(r.Context(), id); err != nil {
			writeErr(w, statusFor(err), err.Error())
			return
		}
	default:
		writeErr(w, http.StatusBadRequest, "decision deve ser approve ou reject")
		return
	}
	out, _ := h.s.GetIntent(r.Context(), id)
	writeJSON(w, http.StatusOK, out)
}
