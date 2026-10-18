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

var ErrProductosCarritoNoEncontrado = errors.New("item del carrito no encontrado")
var ErrReferenciaProductosCarritoInvalida = errors.New("el producto o el carrito no existe")

type ProductosCarritoRepository interface {
	Crear(ctx context.Context, arg db.CreateProductoCarritoParams) (db.ProductosCarrito, error)
	ObtenerPorID(ctx context.Context, id int32) (db.ProductosCarrito, error)
	Listar(ctx context.Context) ([]db.ProductosCarrito, error)
	Actualizar(ctx context.Context, arg db.UpdateProductoCarritoParams) error
	Eliminar(ctx context.Context, id int32) error
}

type productosCarritoRepository struct {
	queries *db.Queries
}

func NewProductosCarritoRepository(queries *db.Queries) ProductosCarritoRepository {
	return &productosCarritoRepository{queries: queries}
}

func (r *productosCarritoRepository) Crear(ctx context.Context, arg db.CreateProductoCarritoParams) (db.ProductosCarrito, error) {
	item, err := r.queries.CreateProductoCarrito(ctx, arg)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return db.ProductosCarrito{}, ErrReferenciaProductosCarritoInvalida
		}
		slog.ErrorContext(ctx, "operacion de productos_carrito", "operacion", "crear", "error", err)
		return db.ProductosCarrito{}, fmt.Errorf("crear productos_carrito: %w", err)
	}
	slog.InfoContext(ctx, "operacion de productos_carrito", "operacion", "crear", "resultado", "ok", "id_productos_carrito", item.IDProductosCarrito)
	return item, nil
}

func (r *productosCarritoRepository) ObtenerPorID(ctx context.Context, id int32) (db.ProductosCarrito, error) {
	item, err := r.queries.GetProductoCarrito(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.ProductosCarrito{}, ErrProductosCarritoNoEncontrado
		}
		slog.ErrorContext(ctx, "operacion de productos_carrito", "operacion", "obtener_por_id", "error", err)
		return db.ProductosCarrito{}, fmt.Errorf("obtener productos_carrito por id: %w", err)
	}
	return item, nil
}

func (r *productosCarritoRepository) Listar(ctx context.Context) ([]db.ProductosCarrito, error) {
	items, err := r.queries.ListProductosCarrito(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "operacion de productos_carrito", "operacion", "listar", "error", err)
		return nil, fmt.Errorf("listar items: %w", err)
	}
	if items == nil {
		items = []db.ProductosCarrito{}
	}
	return items, nil
}

func (r *productosCarritoRepository) Actualizar(ctx context.Context, arg db.UpdateProductoCarritoParams) error {
	_, err := r.queries.UpdateProductoCarrito(ctx, arg)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrProductosCarritoNoEncontrado
		}
		slog.ErrorContext(ctx, "operacion de productos_carrito", "operacion", "actualizar", "error", err)
		return fmt.Errorf("actualizar productos_carrito: %w", err)
	}
	return nil
}

func (r *productosCarritoRepository) Eliminar(ctx context.Context, id int32) error {
	err := r.queries.DeleteProductoCarrito(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "operacion de productos_carrito", "operacion", "eliminar", "error", err)
		return fmt.Errorf("eliminar productos_carrito: %w", err)
	}
	return nil
}
