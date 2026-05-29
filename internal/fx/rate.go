package fx

import (
	"context"
	"errors"
)

var ErrRateUnavailable = errors.New("fx: par de moedas sem taxa")

// RateProvider devolve a taxa efetiva (rateE8) do par base→quote, já com spread.
type RateProvider interface {
	Rate(ctx context.Context, base, quote string) (int64, error)
}
