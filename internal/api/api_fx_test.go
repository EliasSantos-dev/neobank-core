package api_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertFlow(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 100_000) // 1000 BRL via gateway

	resp := authedPost(t, srv.URL, "/me/convert", tok, map[string]any{"from": "BRL", "to": "USD", "amount": 10_000})
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, _ := http.NewRequest("GET", srv.URL+"/me/wallets", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	wresp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer wresp.Body.Close()
	require.Equal(t, http.StatusOK, wresp.StatusCode)
}

func TestConvertInsufficientHTTP(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	resp := authedPost(t, srv.URL, "/me/convert", tok, map[string]any{"from": "BRL", "to": "USD", "amount": 10_000})
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestConvertUnsupportedCurrency(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	resp := authedPost(t, srv.URL, "/me/convert", tok, map[string]any{"from": "BRL", "to": "JPY", "amount": 10_000})
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
