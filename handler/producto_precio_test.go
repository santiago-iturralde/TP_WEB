package handler

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
	"TP_WEB/service"
)

func TestPrecioProductoJSON(t *testing.T) {
	database := abrirDBTest(t)
	lectura := abrirDBTest(t)
	h := NewProductoHandler(service.NewProductoService(repository.NewProductoRepository(db.New(database))))
	nombre := fmt.Sprintf("Precio API %d", time.Now().UnixNano())
	t.Cleanup(func() {
		if _, err := database.Exec("DELETE FROM producto WHERE nombre_prod=$1", nombre); err != nil {
			t.Error(err)
		}
	})
	enviar := func(metodo, precio string, id int32) *httptest.ResponseRecorder {
		body := fmt.Sprintf(`{"nombre_prod":%q,"precio_prod":%s,"stock_prod":1,"categoria":"Prueba"}`, nombre, precio)
		req := httptest.NewRequest(metodo, "/productos", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		if metodo == "PUT" {
			req.SetPathValue("id", fmt.Sprint(id))
			h.Elemento(res, req)
		} else {
			h.Coleccion(res, req)
		}
		return res
	}
	// Valores de dos decimales que pueden tener representación binaria inexacta.
	for _, precio := range []string{"0.01", "0.29", "19.99", "99999999.99"} {
		res := enviar("POST", precio, 0)
		if res.Code != 201 {
			t.Fatalf("alta precio %s: %d %s", precio, res.Code, res.Body.String())
		}
		var producto db.Producto
		if err := json.Unmarshal(res.Body.Bytes(), &producto); err != nil {
			t.Fatal(err)
		}
		// La actualización acepta también los límites y los centavos válidos.
		res = enviar("PUT", precio, producto.IDProducto)
		if res.Code != 200 {
			t.Fatalf("actualizar precio %s: %d %s", precio, res.Code, res.Body.String())
		}
		for _, invalido := range []string{"0", "-1", "0.001", "100000000", "1.234"} {
			for _, metodo := range []string{"POST", "PUT"} {
				res = enviar(metodo, invalido, producto.IDProducto)
				if res.Code != 400 || !strings.Contains(res.Body.String(), service.ErrPrecioInvalido.Error()) {
					t.Fatalf("%s precio %s: %d %s", metodo, invalido, res.Code, res.Body.String())
				}
			}
		}
		var coincide bool
		if err := lectura.QueryRow("SELECT precio_prod = $1::numeric FROM producto WHERE id_producto=$2", precio, producto.IDProducto).Scan(&coincide); err != nil || !coincide {
			t.Fatalf("precio persistido %s: coincide=%v error=%v", precio, coincide, err)
		}
	}
	var cantidad int
	if err := lectura.QueryRow("SELECT count(*) FROM producto WHERE nombre_prod=$1", nombre).Scan(&cantidad); err != nil || cantidad != 4 {
		t.Fatalf("las altas inválidas no deben guardar productos: cantidad=%d error=%v", cantidad, err)
	}
}
