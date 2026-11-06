package handler

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

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
	if err != nil || math.IsNaN(precio) || math.IsInf(precio, 0) || precio < 0.01 || precio > 99999999.99 {
		h.renderIndex(w, r, http.StatusBadRequest, data, "El precio debe estar entre 0.01 y 99999999.99.")
		return
	}
	stock, err := strconv.ParseInt(data.Stock, 10, 32)
	if err != nil || stock < 0 {
		h.renderIndex(w, r, http.StatusBadRequest, data, "El stock debe ser un entero entre 0 y 2147483647.")
		return
	}
	if utf8.RuneCountInString(data.Nombre) > 255 || utf8.RuneCountInString(data.Categoria) > 100 || utf8.RuneCountInString(data.Descripcion) > 1000 {
		h.renderIndex(w, r, http.StatusBadRequest, data, "Revisá las longitudes: nombre hasta 255, categoría hasta 100 y descripción hasta 1000 caracteres.")
		return
	}
	_, err = h.service.Crear(r.Context(), service.CrearProductoDTO{
		Nombre: data.Nombre, Descripcion: data.Descripcion, Precio: precio, Stock: int32(stock), Categoria: data.Categoria,
	})
	if err != nil {
		if errors.Is(err, service.ErrNombreVacio) || errors.Is(err, service.ErrCategoriaVacia) || errors.Is(err, service.ErrPrecioInvalido) || errors.Is(err, service.ErrStockNegativo) {
			h.renderIndex(w, r, http.StatusBadRequest, data, err.Error())
			return
		}
		h.renderIndex(w, r, http.StatusInternalServerError, data, "No se pudo guardar el perfume. Intentá nuevamente.")
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
