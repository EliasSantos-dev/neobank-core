package fx_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/fx"
	"github.com/stretchr/testify/require"
)

func TestFakeRateProvider(t *testing.T) {
	p := fx.NewFakeRateProvider(50) // 50 bps
	ctx := context.Background()

	r, err := p.Rate(ctx, "BRL", "USD")
	require.NoError(t, err)
	require.Greater(t, r, int64(0))

	mid, err := p.Mid("BRL", "USD")
	require.NoError(t, err)
	require.Less(t, r, mid) // spread desfavorável

	_, err = p.Rate(ctx, "BRL", "JPY")
	require.ErrorIs(t, err, fx.ErrRateUnavailable)
}
