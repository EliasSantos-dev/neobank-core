package user

import "testing"

func TestValidate(t *testing.T) {
	if err := Validate("a@b.com", "12345678"); err != nil {
		t.Fatalf("válido: %v", err)
	}
	if Validate("semarroba", "12345678") != ErrInvalidEmail {
		t.Fatal("email inválido")
	}
	if Validate("a@b.com", "curta") != ErrWeakPassword {
		t.Fatal("senha curta")
	}
}
