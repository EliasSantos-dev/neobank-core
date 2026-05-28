package risk

import "testing"

func TestLevelString(t *testing.T) {
	if Low != "low" || High != "high" {
		t.Fatalf("níveis inesperados: %s %s", Low, High)
	}
}
