package service

import (
	"context"
	"errors"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
)

var (
	ErrProductosCarritoNoEncontrado       = repository.ErrProductosCarritoNoEncontrado
	ErrReferenciaProductosCarritoInvalida = repository.ErrReferenciaProductosCarritoInvalida
	ErrProductoCarritoInvalido            = errors.New("el ID de producto debe ser mayor a cero")
	ErrIDCarritoInvalido                  = errors.New("el ID de carrito debe ser mayor a cero")
	ErrCantidadProductosCarritoInvalida   = errors.New("la cantidad debe ser mayor a cero")
)

type ProductosCarritoService struct {
	repo repository.ProductosCarritoRepository
}

func NewProductosCarritoService(repo repository.ProductosCarritoRepository) *ProductosCarritoService {
	return &ProductosCarritoService{repo: repo}
}

type CrearProductosCarritoDTO struct {
	IDProducto int32 `json:"id_producto"`
	IDCarrito  int32 `json:"id_carrito"`
	Cantidad   int32 `json:"cantidad"`
}

// La consulta existente permite actualizar la cantidad del ítem.
type ActualizarProductosCarritoDTO struct {
	Cantidad int32 `json:"cantidad"`
}

func (s *ProductosCarritoService) Crear(ctx context.Context, dto CrearProductosCarritoDTO) (db.ProductosCarrito, error) {
	if dto.IDProducto <= 0 {
		return db.ProductosCarrito{}, ErrProductoCarritoInvalido
	}
	if dto.IDCarrito <= 0 {
		return db.ProductosCarrito{}, ErrIDCarritoInvalido
	}
	if dto.Cantidad <= 0 {
		return db.ProductosCarrito{}, ErrCantidadProductosCarritoInvalida
	}
	return s.repo.Crear(ctx, db.CreateProductoCarritoParams{IDProducto: dto.IDProducto, IDCarrito: dto.IDCarrito, Cantidad: dto.Cantidad})
}

func (s *ProductosCarritoService) ObtenerPorID(ctx context.Context, id int32) (db.ProductosCarrito, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *ProductosCarritoService) Listar(ctx context.Context) ([]db.ProductosCarrito, error) {
	return s.repo.Listar(ctx)
}

func (s *ProductosCarritoService) Actualizar(ctx context.Context, id int32, dto ActualizarProductosCarritoDTO) (db.ProductosCarrito, error) {
	if _, err := s.repo.ObtenerPorID(ctx, id); err != nil {
		return db.ProductosCarrito{}, err
	}
	if dto.Cantidad <= 0 {
		return db.ProductosCarrito{}, ErrCantidadProductosCarritoInvalida
	}
	if err := s.repo.Actualizar(ctx, db.UpdateProductoCarritoParams{IDProductosCarrito: id, Cantidad: dto.Cantidad}); err != nil {
		return db.ProductosCarrito{}, err
	}
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *ProductosCarritoService) Eliminar(ctx context.Context, id int32) error {
	if _, err := s.repo.ObtenerPorID(ctx, id); err != nil {
		return err
	}
	return s.repo.Eliminar(ctx, id)
}
