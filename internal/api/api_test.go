package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/api"
	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestAPI_TransferAndBalance(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	srv := httptest.NewServer(api.NewServer(s))
	defer srv.Close()
	ctx := context.Background()

	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	_, err := s.Transfer(ctx, store.TransferParams{
		IdempotencyKey: "f", FromAccountID: ext.ID, ToAccountID: a.ID, Amount: 5000, Currency: "BRL",
	})
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]any{
		"from_account_id": a.ID, "to_account_id": b.ID, "amount": 2000, "currency": "BRL",
	})
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/transfers", bytes.NewReader(body))
	req.Header.Set("Idempotency-Key", "api-1")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	resp2, err := http.Get(srv.URL + "/accounts/" + b.ID.String() + "/balance")
	require.NoError(t, err)
	defer resp2.Body.Close()
	var out struct {
		Balance int64 `json:"balance"`
	}
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&out))
	require.EqualValues(t, 2000, out.Balance)
}

func TestAPI_TransferMissingIdempotencyKey(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	srv := httptest.NewServer(api.NewServer(s))
	defer srv.Close()

	body, _ := json.Marshal(map[string]any{"amount": 1})
	resp, err := http.Post(srv.URL+"/transfers", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
