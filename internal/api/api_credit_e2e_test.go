package api_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2E_CreditAndMetrics(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 100_000) // gera operação de depósito

	req, _ := http.NewRequest("GET", srv.URL+"/me/credit", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	mresp, err := http.Get(srv.URL + "/metrics")
	require.NoError(t, err)
	defer mresp.Body.Close()
	body, _ := io.ReadAll(mresp.Body)
	require.True(t, strings.Contains(string(body), `neobank_operations_total{op="deposit"}`), string(body))
}
