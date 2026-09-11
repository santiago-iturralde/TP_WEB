package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

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
			slog.WarnContext(ctx, "operacion de usuario", "operacion", "crear", "resultado", "email_duplicado")
			return db.Usuario{}, ErrEmailExistente
		}
		logErrorUsuario(ctx, "crear", err)
		return db.Usuario{}, fmt.Errorf("crear usuario: %w", err)
	}

	slog.InfoContext(ctx, "operacion de usuario", "operacion", "crear", "resultado", "ok", "id_usuario", usuario.IDUsuario)
	return usuario, nil
}

func (r *usuarioRepository) BuscarPorEmail(ctx context.Context, email string) (db.Usuario, error) {
	usuario, err := r.queries.GetUsuarioByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.InfoContext(ctx, "operacion de usuario", "operacion", "buscar_por_email", "resultado", "no_encontrado")
		} else {
			logErrorUsuario(ctx, "buscar_por_email", err)
		}
		return db.Usuario{}, fmt.Errorf("buscar usuario por email: %w", err)
	}
	slog.InfoContext(ctx, "operacion de usuario", "operacion", "buscar_por_email", "resultado", "ok", "id_usuario", usuario.IDUsuario)
	return usuario, nil
}

// No se registra el error completo: PostgreSQL puede incluir datos sensibles
// de la fila en su detalle. El tipo y SQLSTATE permiten identificar el fallo.
func logErrorUsuario(ctx context.Context, operacion string, err error) {
	attrs := []any{"operacion", operacion, "resultado", "error", "tipo_error", fmt.Sprintf("%T", err)}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		attrs = append(attrs, "sqlstate", pgErr.Code)
	}
	slog.ErrorContext(ctx, "operacion de usuario", attrs...)
}
