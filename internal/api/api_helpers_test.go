package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// initiatePayment faz um POST autenticado de depósito/saque e devolve o payment id.
func initiatePayment(t *testing.T, base, path, tok string, amount int64) string {
	t.Helper()
	b, _ := json.Marshal(map[string]int64{"amount": amount})
	req, _ := http.NewRequest("POST", base+path, bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Idempotency-Key", fmt.Sprintf("pay-%d", atomic.AddInt64(&keyCounter, 1)))
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
	var p struct {
		ID string `json:"ID"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&p))
	return p.ID
}

// confirmWebhook simula o callback do provedor para um pagamento.
func confirmWebhook(t *testing.T, base, paymentID, status string) {
	t.Helper()
	b, _ := json.Marshal(map[string]string{
		"event_id": uuid.NewString(), "payment_id": paymentID, "status": status,
	})
	resp, err := http.Post(base+"/webhooks/gateway", "application/json", bytes.NewReader(b))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// fund deposita e confirma via webhook, deixando a wallet com saldo.
func fund(t *testing.T, base, tok string, amount int64) {
	t.Helper()
	pid := initiatePayment(t, base, "/me/deposit", tok, amount)
	confirmWebhook(t, base, pid, "succeeded")
}
