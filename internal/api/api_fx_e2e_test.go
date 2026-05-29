package api_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2E_DepositConvertReconcile(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 100_000) // 1000 BRL

	resp := authedPost(t, srv.URL, "/me/convert", tok, map[string]any{"from": "BRL", "to": "USD", "amount": 50_000})
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var out struct {
		Converted int64 `json:"converted"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))
	require.Greater(t, out.Converted, int64(0))

	rr, _ := http.NewRequest("POST", srv.URL+"/admin/reconcile", nil)
	rr.Header.Set("X-Admin-Token", "admin-token")
	rresp, err := http.DefaultClient.Do(rr)
	require.NoError(t, err)
	defer rresp.Body.Close()
	var rep struct {
		Healthy bool `json:"Healthy"`
	}
	require.NoError(t, json.NewDecoder(rresp.Body).Decode(&rep))
	require.True(t, rep.Healthy)
}
