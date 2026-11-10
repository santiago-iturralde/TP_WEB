package service

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
)

var (
	ErrNombreVacio          = errors.New("el nombre del producto no puede estar vacio")
	ErrCategoriaVacia       = errors.New("la categoria no puede estar vacia")
	ErrPrecioInvalido       = errors.New("el precio debe ser un numero valido mayor a cero, entre 0.01 y 99999999.99 y con hasta dos decimales")
	ErrStockNegativo        = errors.New("el stock no puede ser negativo")
	ErrLongitudInvalida     = errors.New("Revisá las longitudes: nombre hasta 255, categoría hasta 100 y descripción hasta 1000 caracteres.")
	ErrProductoNoEncontrado = repository.ErrProductoNoEncontrado
)

type ProductoService struct {
	repo repository.ProductoRepository
}

func NewProductoService(repo repository.ProductoRepository) *ProductoService {
	return &ProductoService{repo: repo}
}

type CrearProductoDTO struct {
	Nombre      string  `json:"nombre_prod"`
	Descripcion string  `json:"descripcion_prod"`
	Precio      float64 `json:"precio_prod"`
	Stock       int32   `json:"stock_prod"`
	Categoria   string  `json:"categoria"`
}

type ActualizarProductoDTO struct {
	Nombre      string  `json:"nombre_prod"`
	Descripcion string  `json:"descripcion_prod"`
	Precio      float64 `json:"precio_prod"`
	Stock       int32   `json:"stock_prod"`
	Categoria   string  `json:"categoria"`
}

func (s *ProductoService) Crear(ctx context.Context, dto CrearProductoDTO) (db.Producto, error) {
	if err := validarDatosProducto(dto.Nombre, dto.Descripcion, dto.Categoria, dto.Precio, dto.Stock); err != nil {
		return db.Producto{}, err
	}

	params := db.CreateProductoParams{
		NombreProd:      strings.TrimSpace(dto.Nombre),
		DescripcionProd: sql.NullString{String: strings.TrimSpace(dto.Descripcion), Valid: strings.TrimSpace(dto.Descripcion) != ""},
		PrecioProd:      dto.Precio,
		StockProd:       dto.Stock,
		Categoria:       strings.TrimSpace(dto.Categoria),
	}

	return s.repo.Crear(ctx, params)
}

func (s *ProductoService) ObtenerPorID(ctx context.Context, id int32) (db.Producto, error) {
	return s.repo.ObtenerPorID(ctx, id)
}

func (s *ProductoService) Listar(ctx context.Context) ([]db.Producto, error) {
	return s.repo.Listar(ctx)
}

func (s *ProductoService) Actualizar(ctx context.Context, id int32, dto ActualizarProductoDTO) (db.Producto, error) {
	// Verificar existencia previa
	_, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return db.Producto{}, err
	}

	if err := validarDatosProducto(dto.Nombre, dto.Descripcion, dto.Categoria, dto.Precio, dto.Stock); err != nil {
		return db.Producto{}, err
	}

	params := db.UpdateProductoParams{
		IDProducto:      id,
		NombreProd:      strings.TrimSpace(dto.Nombre),
		DescripcionProd: sql.NullString{String: strings.TrimSpace(dto.Descripcion), Valid: strings.TrimSpace(dto.Descripcion) != ""},
		PrecioProd:      dto.Precio,
		StockProd:       dto.Stock,
		Categoria:       strings.TrimSpace(dto.Categoria),
	}

	if err := s.repo.Actualizar(ctx, params); err != nil {
		return db.Producto{}, err
	}

	return s.repo.ObtenerPorID(ctx, id)
}

func (s *ProductoService) Eliminar(ctx context.Context, id int32) error {
	// Verificar existencia previa para retornar 404 si no existe
	_, err := s.repo.ObtenerPorID(ctx, id)
	if err != nil {
		return err
	}
	return s.repo.Eliminar(ctx, id)
}

func validarDatosProducto(nombre, descripcion, categoria string, precio float64, stock int32) error {
	if strings.TrimSpace(nombre) == "" {
		return ErrNombreVacio
	}
	if strings.TrimSpace(categoria) == "" {
		return ErrCategoriaVacia
	}

	if utf8.RuneCountInString(strings.TrimSpace(nombre)) > 255 ||
		utf8.RuneCountInString(strings.TrimSpace(categoria)) > 100 ||
		utf8.RuneCountInString(strings.TrimSpace(descripcion)) > 1000 {
		return ErrLongitudInvalida
	}
	if math.IsNaN(precio) || math.IsInf(precio, 0) || precio < 0.01 || precio > 99999999.99 || precio != math.Round(precio*100)/100 {
		return ErrPrecioInvalido
	}
	if stock < 0 {
		return ErrStockNegativo
	}
	return nil
}
