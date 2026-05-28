package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/EliasSantos-dev/neobank-core/internal/auth"
	"github.com/EliasSantos-dev/neobank-core/internal/user"
)

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h handler) register(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if err := user.Validate(c.Email, c.Password); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	hash, err := auth.HashPassword(c.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "falha ao processar senha")
		return
	}
	u, err := h.s.CreateUser(r.Context(), c.Email, hash)
	if err != nil {
		if errors.Is(err, user.ErrEmailTaken) {
			writeErr(w, http.StatusConflict, err.Error())
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"id": u.ID, "email": u.Email, "wallet_account_id": u.WalletAccountID,
	})
}

func (h handler) login(w http.ResponseWriter, r *http.Request) {
	var c credentials
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	u, hash, err := h.s.GetUserByEmail(r.Context(), c.Email)
	if err != nil || !auth.ComparePassword(hash, c.Password) {
		writeErr(w, http.StatusUnauthorized, "credenciais inválidas")
		return
	}
	tok, err := auth.Issue(h.secret, u.ID, 24*time.Hour)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "falha ao emitir token")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"access_token": tok})
}

func (h handler) me(w http.ResponseWriter, r *http.Request) {
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	bal, err := h.s.Balance(r.Context(), u.WalletAccountID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": u.ID, "email": u.Email, "balance": bal})
}
