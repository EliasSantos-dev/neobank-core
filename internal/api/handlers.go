// Package api é o transporte HTTP do neobank-core.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

type handler struct {
	s      *store.Store
	secret []byte
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
