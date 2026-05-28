package auth

import "testing"

func TestHashAndCompare(t *testing.T) {
	h, err := HashPassword("segredo123")
	if err != nil {
		t.Fatal(err)
	}
	if h == "segredo123" {
		t.Fatal("hash não pode ser a senha em claro")
	}
	if !ComparePassword(h, "segredo123") {
		t.Fatal("deveria casar")
	}
	if ComparePassword(h, "errado") {
		t.Fatal("não deveria casar")
	}
}
