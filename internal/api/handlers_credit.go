package api

import (
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/credit"
)

type creditResponse struct {
	Score   int             `json:"score"`
	Band    string          `json:"band"`
	Factors []credit.Factor `json:"factors"`
}

func (h handler) credit(w http.ResponseWriter, r *http.Request) {
	u, err := h.s.GetUserByID(r.Context(), userID(r))
	if err != nil {
		writeErr(w, statusFor(err), err.Error())
		return
	}
	in, err := h.s.CreditInput(r.Context(), u.ID)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	res := credit.Score(in)
	writeJSON(w, http.StatusOK, creditResponse{Score: res.Score, Band: string(res.Band), Factors: res.Factors})
}
