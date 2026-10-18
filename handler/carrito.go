package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"TP_WEB/service"
)

type CarritoHandler struct {
	service *service.CarritoService
}

func NewCarritoHandler(service *service.CarritoService) *CarritoHandler {
	return &CarritoHandler{service: service}
}

// POST /carritos y GET /carritos
func (h *CarritoHandler) Coleccion(w http.ResponseWriter, r *http.Request) {
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

// GET /carritos/{id}, PUT /carritos/{id} y DELETE /carritos/{id}
func (h *CarritoHandler) Elemento(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		http.Error(w, "ID de carrito invalido", http.StatusBadRequest)
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

func (h *CarritoHandler) crear(w http.ResponseWriter, r *http.Request) {
	var dto service.CrearCarritoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	carrito, err := h.service.Crear(r.Context(), dto)
	if err != nil {
		if errors.Is(err, service.ErrUsuarioCarritoInvalido) ||
			errors.Is(err, service.ErrUsuarioCarritoNoEncontrado) ||
			errors.Is(err, service.ErrTotalCarritoInvalido) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "error al crear el carrito", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(carrito)
}

func (h *CarritoHandler) listar(w http.ResponseWriter, r *http.Request) {
	carritos, err := h.service.Listar(r.Context())
	if err != nil {
		http.Error(w, "error al obtener carritos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(carritos)
}

func (h *CarritoHandler) obtenerPorID(w http.ResponseWriter, r *http.Request, id int32) {
	carrito, err := h.service.ObtenerPorID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCarritoNoEncontrado) {
			http.Error(w, "carrito no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error al obtener el carrito", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(carrito)
}

func (h *CarritoHandler) actualizar(w http.ResponseWriter, r *http.Request, id int32) {
	var dto service.ActualizarCarritoDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON invalido", http.StatusBadRequest)
		return
	}

	carrito, err := h.service.Actualizar(r.Context(), id, dto)
	if err != nil {
		if errors.Is(err, service.ErrCarritoNoEncontrado) {
			http.Error(w, "carrito no encontrado", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrUsuarioCarritoInvalido) ||
			errors.Is(err, service.ErrUsuarioCarritoNoEncontrado) ||
			errors.Is(err, service.ErrTotalCarritoInvalido) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "error al actualizar el carrito", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(carrito)
}

func (h *CarritoHandler) eliminar(w http.ResponseWriter, r *http.Request, id int32) {
	err := h.service.Eliminar(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrCarritoNoEncontrado) {
			http.Error(w, "carrito no encontrado", http.StatusNotFound)
			return
		}
		http.Error(w, "error al eliminar el carrito", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
