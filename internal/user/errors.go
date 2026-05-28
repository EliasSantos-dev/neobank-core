package user

import "errors"

var (
	ErrInvalidEmail       = errors.New("user: email inválido")
	ErrWeakPassword       = errors.New("user: senha deve ter ao menos 8 caracteres")
	ErrEmailTaken         = errors.New("user: email já cadastrado")
	ErrUserNotFound       = errors.New("user: usuário não encontrado")
	ErrInvalidCredentials = errors.New("user: credenciais inválidas")
)
