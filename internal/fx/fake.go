package fx

import "context"

// midRates: taxa mid (rateE8) de base→quote. 1 base = (rateE8/1e8) quote.
var midRates = map[string]map[string]int64{
	"BRL": {"USD": 20_000_000, "EUR": 18_000_000},  // 1 BRL = 0.20 USD / 0.18 EUR
	"USD": {"BRL": 500_000_000, "EUR": 90_000_000}, // 1 USD = 5.00 BRL / 0.90 EUR
	"EUR": {"BRL": 555_000_000, "USD": 111_000_000},
}

type FakeRateProvider struct{ spreadBps int64 }

func NewFakeRateProvider(spreadBps int64) *FakeRateProvider {
	return &FakeRateProvider{spreadBps: spreadBps}
}

func (p *FakeRateProvider) Mid(base, quote string) (int64, error) {
	m, ok := midRates[base]
	if !ok {
		return 0, ErrRateUnavailable
	}
	r, ok := m[quote]
	if !ok {
		return 0, ErrRateUnavailable
	}
	return r, nil
}

// Rate aplica o spread contra o cliente (recebe menos da moeda quote).
func (p *FakeRateProvider) Rate(_ context.Context, base, quote string) (int64, error) {
	mid, err := p.Mid(base, quote)
	if err != nil {
		return 0, err
	}
	return mid * (10_000 - p.spreadBps) / 10_000, nil
}
