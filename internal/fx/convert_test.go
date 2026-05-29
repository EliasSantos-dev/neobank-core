package fx

import "testing"

func TestConvert(t *testing.T) {
	if got := Convert(100, 20_000_000); got != 20 {
		t.Fatalf("esperava 20, veio %d", got)
	}
	if got := Convert(101, 20_000_000); got != 20 {
		t.Fatalf("esperava 20 (floor), veio %d", got)
	}
	if got := Convert(100_000_000_000, 500_000_000); got != 500_000_000_000 {
		t.Fatalf("esperava 500000000000, veio %d", got)
	}
}
