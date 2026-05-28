// Package auth cuida de hashing de senha e tokens JWT (sem IO de banco/HTTP).
package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword gera um hash bcrypt da senha em claro.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// ComparePassword retorna true se a senha em claro corresponde ao hash.
func ComparePassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
