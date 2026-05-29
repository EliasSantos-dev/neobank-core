package gateway_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/gateway"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFakeProvider(t *testing.T) {
	p := gateway.NewFakeProvider()
	ctx := context.Background()

	ref, err := p.CreateCharge(ctx, gateway.ChargeRequest{PaymentID: uuid.New(), Amount: 5000, Currency: "BRL"})
	require.NoError(t, err)
	require.NotEmpty(t, ref)

	require.True(t, gateway.FakeApproves(5000))
	require.False(t, gateway.FakeApproves(1013))
}
