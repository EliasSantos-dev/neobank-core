package risk

import "testing"

func TestStatisticalSignal(t *testing.T) {
	hist := []int64{100, 110, 90, 105, 95} // ~estável em 100
	if sig, ok := statisticalSignal(Context{Amount: 5000, HistoricalAmounts: hist}); !ok || sig.Score == 0 {
		t.Fatal("outlier deveria disparar sinal estatístico")
	}
	if _, ok := statisticalSignal(Context{Amount: 102, HistoricalAmounts: hist}); ok {
		t.Fatal("valor típico não deveria disparar")
	}
	if _, ok := statisticalSignal(Context{Amount: 5000, HistoricalAmounts: []int64{100}}); ok {
		t.Fatal("histórico insuficiente não deveria disparar")
	}
}
