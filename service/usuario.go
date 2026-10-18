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
	ErrUsuarioNoEncontrado = repository.ErrUsuarioNoEncontrado
	ErrEmailInvalido       = errors.New("el email no es valido")
	ErrContrasenaCorta     = errors.New("la contrasena debe tener al menos 8 caracteres")
	ErrEmailExistente      = errors.New("el email ya esta registrado")
)

// UsuarioService concentra las reglas de negocio de usuarios.
type UsuarioService struct {
	repository repository.UsuarioRepository
}

func NewUsuarioService(repo repository.UsuarioRepository) *UsuarioService {
	return &UsuarioService{repository: repo}
}

func (s *UsuarioService) Crear(ctx context.Context, email, contrasena string) (db.Usuario, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if err := validarDatosUsuario(ctx, email, contrasena); err != nil {
		return db.Usuario{}, err
	}
	usuario, err := s.repository.Crear(ctx, email, contrasena)
	if errors.Is(err, repository.ErrEmailExistente) {
		return db.Usuario{}, ErrEmailExistente
	}
	return usuario, err
}

func validarDatosUsuario(ctx context.Context, email, contrasena string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		slog.WarnContext(ctx, "datos de usuario rechazados", "motivo", "email_invalido")
		return ErrEmailInvalido
	}
	if len(contrasena) < 8 {
		slog.WarnContext(ctx, "datos de usuario rechazados", "motivo", "contrasena_corta")
		return ErrContrasenaCorta
	}
	return nil
}

type ActualizarUsuarioDTO struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
}

func (s *UsuarioService) ObtenerPorID(ctx context.Context, id int32) (db.Usuario, error) {
	return s.repository.ObtenerPorID(ctx, id)
}

func (s *UsuarioService) Listar(ctx context.Context) ([]db.Usuario, error) {
	return s.repository.Listar(ctx)
}

func (s *UsuarioService) Actualizar(ctx context.Context, id int32, dto ActualizarUsuarioDTO) (db.Usuario, error) {
	if _, err := s.repository.ObtenerPorID(ctx, id); err != nil {
		return db.Usuario{}, err
	}
	email := strings.ToLower(strings.TrimSpace(dto.Email))
	if err := validarDatosUsuario(ctx, email, dto.Contrasena); err != nil {
		return db.Usuario{}, err
	}
	if err := s.repository.Actualizar(ctx, db.UpdateUsuarioParams{IDUsuario: id, Email: email, Contrasena: dto.Contrasena}); err != nil {
		if errors.Is(err, repository.ErrEmailExistente) {
			return db.Usuario{}, ErrEmailExistente
		}
		return db.Usuario{}, err
	}
	return s.repository.ObtenerPorID(ctx, id)
}

func (s *UsuarioService) Eliminar(ctx context.Context, id int32) error {
	if _, err := s.repository.ObtenerPorID(ctx, id); err != nil {
		return err
	}
	return s.repository.Eliminar(ctx, id)
}
