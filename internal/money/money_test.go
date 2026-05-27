package money

import "testing"

func TestNew(t *testing.T) {
	cases := []struct {
		name     string
		amount   int64
		currency string
		wantErr  error
	}{
		{"válido", 1000, "BRL", nil},
		{"negativo", -1, "BRL", ErrNegativeAmount},
		{"moeda inválida", 100, "brl", ErrInvalidCurrency},
		{"moeda curta", 100, "BR", ErrInvalidCurrency},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := New(c.amount, c.currency)
			if err != c.wantErr {
				t.Fatalf("New() err = %v, want %v", err, c.wantErr)
			}
		})
	}
}
