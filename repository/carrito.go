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

var ErrCarritoNoEncontrado = errors.New("carrito no encontrado")
var ErrUsuarioCarritoNoEncontrado = errors.New("el usuario del carrito no existe")

type CarritoRepository interface {
	Crear(ctx context.Context, arg db.CreateCarritoParams) (db.Carrito, error)
	ObtenerPorID(ctx context.Context, id int32) (db.Carrito, error)
	Listar(ctx context.Context) ([]db.Carrito, error)
	Actualizar(ctx context.Context, arg db.UpdateCarritoParams) error
	Eliminar(ctx context.Context, id int32) error
}

type carritoRepository struct {
	queries *db.Queries
}

func NewCarritoRepository(queries *db.Queries) CarritoRepository {
	return &carritoRepository{queries: queries}
}

func (r *carritoRepository) Crear(ctx context.Context, arg db.CreateCarritoParams) (db.Carrito, error) {
	carrito, err := r.queries.CreateCarrito(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return db.Carrito{}, ErrUsuarioCarritoNoEncontrado
		}
		slog.ErrorContext(ctx, "operacion de carrito", "operacion", "crear", "error", err)
		return db.Carrito{}, fmt.Errorf("crear carrito: %w", err)
	}
	slog.InfoContext(ctx, "operacion de carrito", "operacion", "crear", "resultado", "ok", "id_carrito", carrito.IDCarrito)
	return carrito, nil
}

func (r *carritoRepository) ObtenerPorID(ctx context.Context, id int32) (db.Carrito, error) {
	carrito, err := r.queries.GetCarrito(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Carrito{}, ErrCarritoNoEncontrado
		}
		slog.ErrorContext(ctx, "operacion de carrito", "operacion", "obtener_por_id", "error", err)
		return db.Carrito{}, fmt.Errorf("obtener carrito por id: %w", err)
	}
	return carrito, nil
}

func (r *carritoRepository) Listar(ctx context.Context) ([]db.Carrito, error) {
	carritos, err := r.queries.ListCarritos(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "operacion de carrito", "operacion", "listar", "error", err)
		return nil, fmt.Errorf("listar carritos: %w", err)
	}
	if carritos == nil {
		carritos = []db.Carrito{}
	}
	return carritos, nil
}

func (r *carritoRepository) Actualizar(ctx context.Context, arg db.UpdateCarritoParams) error {
	err := r.queries.UpdateCarrito(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrUsuarioCarritoNoEncontrado
		}
		slog.ErrorContext(ctx, "operacion de carrito", "operacion", "actualizar", "error", err)
		return fmt.Errorf("actualizar carrito: %w", err)
	}
	return nil
}

func (r *carritoRepository) Eliminar(ctx context.Context, id int32) error {
	err := r.queries.DeleteCarrito(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "operacion de carrito", "operacion", "eliminar", "error", err)
		return fmt.Errorf("eliminar carrito: %w", err)
	}
	return nil
}
