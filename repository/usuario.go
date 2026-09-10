package repository

import (
	"context"
	"errors"
	"fmt"

	db "TP_WEB/db/sqlc"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrEmailExistente = errors.New("el email ya esta registrado")

// UsuarioRepository concentra las consultas a la tabla usuario.
type UsuarioRepository interface {
	Crear(ctx context.Context, email, contrasena string) (db.Usuario, error)
	BuscarPorEmail(ctx context.Context, email string) (db.Usuario, error)
}

type usuarioRepository struct {
	queries *db.Queries
}

func NewUsuarioRepository(queries *db.Queries) UsuarioRepository {
	return &usuarioRepository{queries: queries}
}

func (r *usuarioRepository) Crear(ctx context.Context, email, contrasena string) (db.Usuario, error) {
	usuario, err := r.queries.CreateUsuario(ctx, db.CreateUsuarioParams{
		Email:      email,
		Contrasena: contrasena,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return db.Usuario{}, ErrEmailExistente
		}
		return db.Usuario{}, fmt.Errorf("crear usuario: %w", err)
	}

	return usuario, nil
}

func (r *usuarioRepository) BuscarPorEmail(ctx context.Context, email string) (db.Usuario, error) {
	usuario, err := r.queries.GetUsuarioByEmail(ctx, email)
	if err != nil {
		return db.Usuario{}, fmt.Errorf("buscar usuario por email: %w", err)
	}
	return usuario, nil
}
