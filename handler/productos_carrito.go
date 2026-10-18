package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"TP_WEB/service"
)

type ProductosCarritoHandler struct {
	service *service.ProductosCarritoService
}

func NewProductosCarritoHandler(service *service.ProductosCarritoService) *ProductosCarritoHandler {
	return &ProductosCarritoHandler{service: service}
}

// POST /productos_carrito y GET /productos_carrito
func (h *ProductosCarritoHandler) Coleccion(w http.ResponseWriter, r *http.Request) {
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

// GET /productos_carrito/{id}, PUT /productos_carrito/{id} y DELETE /productos_carrito/{id}
func (h *ProductosCarritoHandler) Elemento(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		http.Error(w, "ID de item invalido", http.StatusBadRequest)
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

func (h *ProductosCarritoHandler) crear(w http.ResponseWriter, r *http.Request) {
	var dto service.CrearProductosCarritoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	item, err := h.service.Crear(r.Context(), dto)
	if err != nil {
		if errors.Is(err, service.ErrProductoCarritoInvalido) ||
			errors.Is(err, service.ErrIDCarritoInvalido) ||
			errors.Is(err, service.ErrCantidadProductosCarritoInvalida) ||
			errors.Is(err, service.ErrReferenciaProductosCarritoInvalida) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "error al crear el item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(item)
}

func (h *ProductosCarritoHandler) listar(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Listar(r.Context())
	if err != nil {
		http.Error(w, "error al obtener los items del carrito", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(items)
}

func (h *ProductosCarritoHandler) obtenerPorID(w http.ResponseWriter, r *http.Request, id int32) {
	item, err := h.service.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductosCarritoNoEncontrado) {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error al obtener el item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(item)
}

func (h *ProductosCarritoHandler) actualizar(w http.ResponseWriter, r *http.Request, id int32) {
	var dto service.ActualizarProductosCarritoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	item, err := h.service.Actualizar(r.Context(), id, dto)
	if err != nil {
		if errors.Is(err, service.ErrProductosCarritoNoEncontrado) {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrProductoCarritoInvalido) ||
			errors.Is(err, service.ErrIDCarritoInvalido) ||
			errors.Is(err, service.ErrCantidadProductosCarritoInvalida) ||
			errors.Is(err, service.ErrReferenciaProductosCarritoInvalida) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "error al actualizar el item", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(item)
}

func (h *ProductosCarritoHandler) eliminar(w http.ResponseWriter, r *http.Request, id int32) {
	err := h.service.Eliminar(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductosCarritoNoEncontrado) {
			http.Error(w, "item no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error al eliminar el item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
