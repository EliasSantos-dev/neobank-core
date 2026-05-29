package api_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGatewayDepositReturns202(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	resp := authedPost(t, srv.URL, "/me/deposit", tok, map[string]int64{"amount": 5000})
	defer resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	require.EqualValues(t, 0, meBalance(t, srv.URL, tok)) // ainda não creditou
}

func TestE2E_GatewayDepositCredits(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")

	pid := initiatePayment(t, srv.URL, "/me/deposit", tok, 5000)
	require.EqualValues(t, 0, meBalance(t, srv.URL, tok))
	confirmWebhook(t, srv.URL, pid, "succeeded")
	require.EqualValues(t, 5000, meBalance(t, srv.URL, tok))
}

func TestE2E_GatewayWithdrawHoldThenFailRefunds(t *testing.T) {
	srv := newSrv(t)
	register(t, srv.URL, "a@b.com", "segredo123")
	tok := login(t, srv.URL, "a@b.com", "segredo123")
	fund(t, srv.URL, tok, 5000)

	wid := initiatePayment(t, srv.URL, "/me/withdraw", tok, 2000)
	require.EqualValues(t, 3000, meBalance(t, srv.URL, tok)) // hold reservou
	confirmWebhook(t, srv.URL, wid, "failed")
	require.EqualValues(t, 5000, meBalance(t, srv.URL, tok)) // estornou
}
