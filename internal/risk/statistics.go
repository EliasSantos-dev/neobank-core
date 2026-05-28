package risk

import "math"

const (
	minHistory = 3
	zThreshold = 3.0
)

// statisticalSignal dispara quando Amount é um outlier (|z| > zThreshold)
// frente ao histórico de valores do remetente.
func statisticalSignal(c Context) (Signal, bool) {
	if len(c.HistoricalAmounts) < minHistory {
		return Signal{}, false
	}
	var sum float64
	for _, a := range c.HistoricalAmounts {
		sum += float64(a)
	}
	mean := sum / float64(len(c.HistoricalAmounts))

	var variance float64
	for _, a := range c.HistoricalAmounts {
		d := float64(a) - mean
		variance += d * d
	}
	variance /= float64(len(c.HistoricalAmounts))
	std := math.Sqrt(variance)
	if std == 0 {
		return Signal{}, false
	}

	z := math.Abs((float64(c.Amount) - mean) / std)
	if z > zThreshold {
		score := int(math.Min(40, z*8))
		return Signal{Name: "amount_anomaly", Score: score, Reason: "valor atípico frente ao histórico do remetente"}, true
	}
	return Signal{}, false
}
