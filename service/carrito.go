package service

import (
	"context"
	"database/sql"
	"errors"
	"math"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
)

var (
	ErrCarritoNoEncontrado        = repository.ErrCarritoNoEncontrado
	ErrUsuarioCarritoNoEncontrado = repository.ErrUsuarioCarritoNoEncontrado
	ErrUsuarioCarritoInvalido     = errors.New("el ID de usuario debe ser mayor a cero")
	ErrTotalCarritoInvalido       = errors.New("el total debe ser un numero valido entre 0 y 99999999.99")
)

type CarritoService struct {
	repo repository.CarritoRepository
}

func NewCarritoService(repo repository.CarritoRepository) *CarritoService {
	return &CarritoService{repo: repo}
}

type CrearCarritoDTO struct {
	IDUsuario *int32  `json:"id_usuario"`
	Total     float64 `json:"total"`
}

type ActualizarCarritoDTO struct {
	IDUsuario *int32  `json:"id_usuario"`
	Total     float64 `json:"total"`
}

func (s *CarritoService) Crear(ctx context.Context, dto CrearCarritoDTO) (db.Carrito, error) {
	if err := validarDatosCarrito(dto.IDUsuario, dto.Total); err != nil {
		return db.Carrito{}, err
	}
	return s.repo.Crear(ctx, db.CreateCarritoParams{IDUsuario: usuarioCarrito(dto.IDUsuario), Total: dto.Total})
}

func (s *CarritoService) ObtenerPorID(ctx context.Context, id int32) (db.Carrito, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *CarritoService) Listar(ctx context.Context) ([]db.Carrito, error) {
	return s.repo.Listar(ctx)
}

func (s *CarritoService) Actualizar(ctx context.Context, id int32, dto ActualizarCarritoDTO) (db.Carrito, error) {
	if _, err := s.repo.ObtenerPorID(ctx, id); err != nil {
		return db.Carrito{}, err
	}
	if err := validarDatosCarrito(dto.IDUsuario, dto.Total); err != nil {
		return db.Carrito{}, err
	}
	if err := s.repo.Actualizar(ctx, db.UpdateCarritoParams{IDCarrito: id, IDUsuario: usuarioCarrito(dto.IDUsuario), Total: dto.Total}); err != nil {
		return db.Carrito{}, err
	}
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *CarritoService) Eliminar(ctx context.Context, id int32) error {
	if _, err := s.repo.ObtenerPorID(ctx, id); err != nil {
		return err
	}
	return s.repo.Eliminar(ctx, id)
}

func validarDatosCarrito(idUsuario *int32, total float64) error {
	if idUsuario != nil && *idUsuario <= 0 {
		return ErrUsuarioCarritoInvalido
	}
	if math.IsNaN(total) || math.IsInf(total, 0) || total < 0 || total > 99999999.99 {
		return ErrTotalCarritoInvalido
	}
	return nil
}

func usuarioCarrito(id *int32) sql.NullInt32 {
	if id == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *id, Valid: true}
}
