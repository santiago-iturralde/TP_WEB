package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"TP_WEB/service"
)

type ProductoHandler struct {
	service *service.ProductoService
}

func NewProductoHandler(service *service.ProductoService) *ProductoHandler {
	return &ProductoHandler{service: service}
}

// POST /productos y GET /productos
func (h *ProductoHandler) Coleccion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.crear(w, r)
	case http.MethodGet:
		h.listar(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
	}
}

// GET /productos/{id}, PUT /productos/{id} y DELETE /productos/{id}
func (h *ProductoHandler) Elemento(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		http.Error(w, "ID de producto invalido", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.obtenerPorID(w, r, int32(id))
	case http.MethodPut:
		h.actualizar(w, r, int32(id))
	case http.MethodDelete:
		h.eliminar(w, r, int32(id))
	default:
		w.Header().Set("Allow", "GET, PUT, DELETE")
		http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *ProductoHandler) crear(w http.ResponseWriter, r *http.Request) {
	var dto service.CrearProductoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	prod, err := h.service.Crear(r.Context(), dto)
	if err != nil {
		if errors.Is(err, service.ErrNombreVacio) ||
			errors.Is(err, service.ErrCategoriaVacia) ||
			errors.Is(err, service.ErrPrecioInvalido) ||
			errors.Is(err, service.ErrStockNegativo) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "error al crear el producto", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(prod)
}

func (h *ProductoHandler) listar(w http.ResponseWriter, r *http.Request) {
	productos, err := h.service.Listar(r.Context())
	if err != nil {
		http.Error(w, "error al obtener productos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(productos)
}

func (h *ProductoHandler) obtenerPorID(w http.ResponseWriter, r *http.Request, id int32) {
	prod, err := h.service.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductoNoEncontrado) {
			http.Error(w, "producto no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error al obtener el producto", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(prod)
}

func (h *ProductoHandler) actualizar(w http.ResponseWriter, r *http.Request, id int32) {
	var dto service.ActualizarProductoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	prod, err := h.service.Actualizar(r.Context(), id, dto)
	if err != nil {
		if errors.Is(err, service.ErrProductoNoEncontrado) {
			http.Error(w, "producto no encontrado", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrNombreVacio) ||
			errors.Is(err, service.ErrCategoriaVacia) ||
			errors.Is(err, service.ErrPrecioInvalido) ||
			errors.Is(err, service.ErrStockNegativo) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "error al actualizar el producto", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(prod)
}

func (h *ProductoHandler) eliminar(w http.ResponseWriter, r *http.Request, id int32) {
	err := h.service.Eliminar(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductoNoEncontrado) {
			http.Error(w, "producto no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error al eliminar el producto", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
