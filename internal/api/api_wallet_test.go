package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

var keyCounter int64

func authedPost(t *testing.T, base, path, tok string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", base+path, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Idempotency-Key", fmt.Sprintf("k-%d", atomic.AddInt64(&keyCounter, 1)))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func meBalance(t *testing.T, base, tok string) int64 {
	t.Helper()
	req, _ := http.NewRequest("GET", base+"/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var me struct {
		Balance int64 `json:"balance"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&me))
	return me.Balance
}

// TestDepositWithdraw cobre depósito/saque via gateway (assíncronos): ambos
// respondem 202 e só refletem no saldo após o webhook do provedor.
func TestDepositWithdraw(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")

	fund(t, srv.URL, tok, 5000) // deposita + confirma webhook
	require.EqualValues(t, 5000, meBalance(t, srv.URL, tok))

	wid := initiatePayment(t, srv.URL, "/me/withdraw", tok, 500)
	require.EqualValues(t, 4500, meBalance(t, srv.URL, tok)) // hold reservou na hora
	confirmWebhook(t, srv.URL, wid, "succeeded")
	require.EqualValues(t, 4500, meBalance(t, srv.URL, tok))
}

// TestTransferCreatesIntent verifica que a transferência vira intenção (202),
// sem efetivar na hora.
func TestTransferCreatesIntent(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	register(t, srv.URL, "c@d.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 5000)

	resp := authedPost(t, srv.URL, "/transfers", tok, map[string]any{"to_email": "c@d.com", "amount": 1000})
	defer resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode) // 202, ainda não efetivou
	require.EqualValues(t, 5000, meBalance(t, srv.URL, tok))
}

func TestWithdrawInsufficient(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	resp := authedPost(t, srv.URL, "/me/withdraw", tok, map[string]int64{"amount": 100})
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestStatement(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 5000)

	req, _ := http.NewRequest("GET", srv.URL+"/me/statement?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var entries []struct {
		Direction string `json:"Direction"`
		Amount    int64  `json:"Amount"`
		Currency  string `json:"Currency"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&entries))
	require.Len(t, entries, 1)
	require.Equal(t, "credit", entries[0].Direction)
	require.EqualValues(t, 5000, entries[0].Amount)
	require.Equal(t, "BRL", entries[0].Currency)
}

func TestTransferUnknownRecipient(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 100)
	resp := authedPost(t, srv.URL, "/transfers", tok, map[string]any{"to_email": "ninguem@x.com", "amount": 50})
	defer resp.Body.Close()
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
