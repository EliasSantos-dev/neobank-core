package api

import (
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/recon"
)

func (h handler) reconcile(w http.ResponseWriter, r *http.Request) {
	rep, err := recon.Run(r.Context(), h.s)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rep)
}
