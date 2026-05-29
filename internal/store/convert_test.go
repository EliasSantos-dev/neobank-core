package store_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestConvertCurrency(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	u, _ := s.CreateUser(ctx, "a@b.com", "h")
	brl, _ := s.GetOrCreateWallet(ctx, u.ID, "BRL")
	_, err := s.Transfer(ctx, store.TransferParams{IdempotencyKey: "fund", FromAccountID: ledger.FxBRL, ToAccountID: brl, Amount: 100_000, Currency: "BRL"})
	require.NoError(t, err)

	out, err := s.ConvertCurrency(ctx, store.ConvertParams{
		IdempotencyKey: "cv1", UserID: u.ID, From: "BRL", To: "USD", Amount: 10_000, RateE8: 20_000_000,
	})
	require.NoError(t, err)
	require.EqualValues(t, 2_000, out.Converted)

	balBRL, _ := s.Balance(ctx, brl)
	require.EqualValues(t, 90_000, balBRL)

	usd, _ := s.GetOrCreateWallet(ctx, u.ID, "USD")
	balUSD, _ := s.Balance(ctx, usd)
	require.EqualValues(t, 2_000, balUSD)

	// idempotência
	out2, err := s.ConvertCurrency(ctx, store.ConvertParams{
		IdempotencyKey: "cv1", UserID: u.ID, From: "BRL", To: "USD", Amount: 10_000, RateE8: 20_000_000,
	})
	require.NoError(t, err)
	require.EqualValues(t, 2_000, out2.Converted)
	balBRL, _ = s.Balance(ctx, brl)
	require.EqualValues(t, 90_000, balBRL)
}

func TestConvertInsufficient(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	u, _ := s.CreateUser(ctx, "a@b.com", "h")
	_, err := s.ConvertCurrency(ctx, store.ConvertParams{
		IdempotencyKey: "cv1", UserID: u.ID, From: "BRL", To: "USD", Amount: 10_000, RateE8: 20_000_000,
	})
	require.ErrorIs(t, err, ledger.ErrInsufficientFunds)
}

func TestConvertConservationPerCurrency(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()
	u, _ := s.CreateUser(ctx, "a@b.com", "h")
	brl, _ := s.GetOrCreateWallet(ctx, u.ID, "BRL")
	_, _ = s.Transfer(ctx, store.TransferParams{IdempotencyKey: "fund", FromAccountID: ledger.FxBRL, ToAccountID: brl, Amount: 100_000, Currency: "BRL"})
	_, err := s.ConvertCurrency(ctx, store.ConvertParams{
		IdempotencyKey: "cv1", UserID: u.ID, From: "BRL", To: "USD", Amount: 10_000, RateE8: 20_000_000,
	})
	require.NoError(t, err)

	// conservação por moeda: o spread fica na conta fx_USD (paga 2000), wallet recebe 2000
	usd, _ := s.GetOrCreateWallet(ctx, u.ID, "USD")
	balUSD, _ := s.Balance(ctx, usd)
	balFxUSD, _ := s.Balance(ctx, ledger.FxUSD)
	require.EqualValues(t, 0, balUSD+balFxUSD) // Σ USD = 0
	balFxBRL, _ := s.Balance(ctx, ledger.FxBRL)
	// Σ BRL: fundeamento veio de fx_BRL (-100000), wallet tem 90000, fx_BRL recebeu +10000 da conversão
	balBRL, _ := s.Balance(ctx, brl)
	require.EqualValues(t, 0, balBRL+balFxBRL) // Σ BRL = 0
}
