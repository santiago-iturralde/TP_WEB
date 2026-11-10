package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"TP_WEB/service"
	"TP_WEB/views"
)

// crearFormulario recibe campos HTML, guarda con el servicio y aplica Post/Redirect/Get.
func (h *ProductoHandler) crearFormulario(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario inválido o demasiado grande", http.StatusBadRequest)
		return
	}
	data := views.ProductFormData{
		Nombre:      strings.TrimSpace(r.PostForm.Get("nombre_prod")),
		Descripcion: strings.TrimSpace(r.PostForm.Get("descripcion_prod")),
		Precio:      strings.TrimSpace(r.PostForm.Get("precio_prod")),
		Stock:       strings.TrimSpace(r.PostForm.Get("stock_prod")),
		Categoria:   strings.TrimSpace(r.PostForm.Get("categoria")),
	}
	precio, err := strconv.ParseFloat(data.Precio, 64)
	if err != nil {
		h.renderIndex(w, r, http.StatusBadRequest, data, service.ErrPrecioInvalido.Error())
		return
	}
	stock, err := strconv.ParseInt(data.Stock, 10, 32)
	if err != nil || stock < 0 {
		h.renderIndex(w, r, http.StatusBadRequest, data, "El stock debe ser un entero entre 0 y 2147483647.")
		return
	}
	_, err = h.service.Crear(r.Context(), service.CrearProductoDTO{
		Nombre: data.Nombre, Descripcion: data.Descripcion, Precio: precio, Stock: int32(stock), Categoria: data.Categoria,
	})
	if err != nil {
		if errors.Is(err, service.ErrNombreVacio) || errors.Is(err, service.ErrCategoriaVacia) || errors.Is(err, service.ErrPrecioInvalido) || errors.Is(err, service.ErrStockNegativo) || errors.Is(err, service.ErrLongitudInvalida) {
			h.renderIndex(w, r, http.StatusBadRequest, data, err.Error())
			return
		}
		h.renderIndex(w, r, http.StatusInternalServerError, data, "No se pudo guardar el perfume. Intentá nuevamente.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
