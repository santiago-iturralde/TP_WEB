package handler

import (
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
	"TP_WEB/service"
)

func TestAltaProductoFormulario(t *testing.T) {
	database := abrirDBTest(t)
	lectura := abrirDBTest(t)
	h := NewProductoHandler(service.NewProductoService(repository.NewProductoRepository(db.New(database))))
	nombre := fmt.Sprintf("Perfume SSR %d", time.Now().UnixNano())
	t.Cleanup(func() {
		if _, err := database.Exec("DELETE FROM producto WHERE nombre_prod=$1", nombre); err != nil {
			t.Error(err)
		}
	})
	campos := url.Values{"nombre_prod": {nombre}, "descripcion_prod": {"Fragancia de prueba"}, "precio_prod": {"150.50"}, "stock_prod": {"3"}, "categoria": {"Prueba"}}
	enviar := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "/productos", strings.NewReader(campos.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()
		h.Coleccion(res, req)
		return res
	}
	res := enviar()
	if res.Code != 303 || res.Header().Get("Location") != "/" {
		t.Fatalf("alta: %d %s", res.Code, res.Body.String())
	}
	var cantidad int
	if err := lectura.QueryRow("SELECT count(*) FROM producto WHERE nombre_prod=$1 AND precio_prod=150.50 AND stock_prod=3", nombre).Scan(&cantidad); err != nil || cantidad != 1 {
		t.Fatalf("persistencia: %d %v", cantidad, err)
	}
	res = httptest.NewRecorder()
	h.Index(res, httptest.NewRequest("GET", "/", nil))
	if res.Code != 200 || !strings.Contains(res.Body.String(), nombre) || !strings.Contains(res.Body.String(), `action="/productos"`) {
		t.Fatalf("catálogo: %d %s", res.Code, res.Body.String())
	}
	for _, precio := range []string{"NaN", "abc", "-1"} {
		campos.Set("precio_prod", precio)
		res = enviar()
		if res.Code != 400 || !strings.Contains(res.Body.String(), `value="`+nombre+`"`) {
			t.Fatalf("validación: %d %s", res.Code, res.Body.String())
		}
	}
	if err := lectura.QueryRow("SELECT count(*) FROM producto WHERE nombre_prod=$1", nombre).Scan(&cantidad); err != nil || cantidad != 1 {
		t.Fatalf("alta inválida: %d %v", cantidad, err)
	}
	res = httptest.NewRecorder()
	h.Coleccion(res, httptest.NewRequest("GET", "/productos", nil))
	if res.Code != 200 || !strings.Contains(res.Header().Get("Content-Type"), "application/json") {
		t.Fatal("API JSON no disponible")
	}
}
