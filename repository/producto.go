package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	db "TP_WEB/db/sqlc"
)

var ErrProductoNoEncontrado = errors.New("producto no encontrado")

type ProductoRepository interface {
	Crear(ctx context.Context, arg db.CreateProductoParams) (db.Producto, error)
	ObtenerPorID(ctx context.Context, id int32) (db.Producto, error)
	Listar(ctx context.Context) ([]db.Producto, error)
	Actualizar(ctx context.Context, arg db.UpdateProductoParams) error
	Eliminar(ctx context.Context, id int32) error
}

type productoRepository struct {
	queries *db.Queries
}

func NewProductoRepository(queries *db.Queries) ProductoRepository {
	return &productoRepository{queries: queries}
}

func (r *productoRepository) Crear(ctx context.Context, arg db.CreateProductoParams) (db.Producto, error) {
	prod, err := r.queries.CreateProducto(ctx, arg)
	if err != nil {
		slog.ErrorContext(ctx, "operacion de producto", "operacion", "crear", "error", err)
		return db.Producto{}, fmt.Errorf("crear producto: %w", err)
	}
	slog.InfoContext(ctx, "operacion de producto", "operacion", "crear", "resultado", "ok", "id_producto", prod.IDProducto)
	return prod, nil
}

func (r *productoRepository) ObtenerPorID(ctx context.Context, id int32) (db.Producto, error) {
	prod, err := r.queries.GetProducto(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.Producto{}, ErrProductoNoEncontrado
		}
		slog.ErrorContext(ctx, "operacion de producto", "operacion", "obtener_por_id", "error", err)
		return db.Producto{}, fmt.Errorf("obtener producto por id: %w", err)
	}
	return prod, nil
}

func (r *productoRepository) Listar(ctx context.Context) ([]db.Producto, error) {
	productos, err := r.queries.ListProductos(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "operacion de producto", "operacion", "listar", "error", err)
		return nil, fmt.Errorf("listar productos: %w", err)
	}
	if productos == nil {
		productos = []db.Producto{}
	}
	return productos, nil
}

func (r *productoRepository) Actualizar(ctx context.Context, arg db.UpdateProductoParams) error {
	err := r.queries.UpdateProducto(ctx, arg)
	if err != nil {
		slog.ErrorContext(ctx, "operacion de producto", "operacion", "actualizar", "error", err)
		return fmt.Errorf("actualizar producto: %w", err)
	}
	return nil
}

func (r *productoRepository) Eliminar(ctx context.Context, id int32) error {
	err := r.queries.DeleteProducto(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "operacion de producto", "operacion", "eliminar", "error", err)
		return fmt.Errorf("eliminar producto: %w", err)
	}
	return nil
}
