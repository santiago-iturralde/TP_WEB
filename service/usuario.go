package service

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
)

var (
	ErrEmailInvalido   = errors.New("el email no es valido")
	ErrContrasenaCorta = errors.New("la contrasena debe tener al menos 8 caracteres")
	ErrEmailExistente  = errors.New("el email ya esta registrado")
)

// UsuarioService concentra las reglas de negocio del alta de usuarios.
type UsuarioService struct {
	repository repository.UsuarioRepository
}

func NewUsuarioService(repo repository.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repository: repo}
}

func (s *UsuarioService) Crear(ctx context.Context, email, contrasena string) (db.Usuario, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		slog.WarnContext(ctx, "alta de usuario rechazada", "motivo", "email_invalido")
		return db.Usuario{}, ErrEmailInvalido
	}
	if len(contrasena) < 8 {
		slog.WarnContext(ctx, "alta de usuario rechazada", "motivo", "contrasena_corta")
		return db.Usuario{}, ErrContrasenaCorta
	}
	usuario, err := s.repository.Crear(ctx, email, contrasena)
	if errors.Is(err, repository.ErrEmailExistente) {
		return db.Usuario{}, ErrEmailExistente
	}
	return usuario, err
}
