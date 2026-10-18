package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
	"TP_WEB/service"
)

// Verifica el contrato HTTP y la persistencia real de los ítems del carrito.
func TestProductosCarritoCRUD(t *testing.T) {
	database := abrirDBTest(t)
	lectura := abrirDBTest(t)
	h := NewProductosCarritoHandler(service.NewProductosCarritoService(repository.NewProductosCarritoRepository(db.New(database))))
	mux := http.NewServeMux()
	mux.HandleFunc("/productos_carrito", h.Coleccion)
	mux.HandleFunc("/productos_carrito/{id}", h.Elemento)
	pedir := func(metodo, ruta, cuerpo string, esperado int) *httptest.ResponseRecorder {
		t.Helper()
		res := httptest.NewRecorder()
		req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
		req.Header.Set("Content-Type", "application/json")
		mux.ServeHTTP(res, req)
		if res.Code != esperado {
			t.Fatalf("%s %s: status=%d, esperado=%d: %s", metodo, ruta, res.Code, esperado, res.Body.String())
		}
		return res
	}
	var productoID, carritoID int32
	if err := database.QueryRow("INSERT INTO producto (nombre_prod, precio_prod, categoria) VALUES ('Perfume test', 100, 'Perfumes') RETURNING id_producto").Scan(&productoID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := database.Exec("DELETE FROM producto WHERE id_producto=$1", productoID); err != nil {
			t.Error(err)
		}
	})
	if err := database.QueryRow("INSERT INTO carrito DEFAULT VALUES RETURNING id_carrito").Scan(&carritoID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := database.Exec("DELETE FROM carrito WHERE id_carrito=$1", carritoID); err != nil {
			t.Error(err)
		}
	})
	cuerpo := fmt.Sprintf(`{"id_producto":%d,"id_carrito":%d,"cantidad":2}`, productoID, carritoID)
	res := pedir("POST", "/productos_carrito", cuerpo, 201)
	var item db.ProductosCarrito
	if err := json.Unmarshal(res.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if item.IDProductosCarrito <= 0 || item.IDProducto != productoID || item.IDCarrito != carritoID || item.Cantidad != 2 {
		t.Fatalf("item inesperado: %+v", item)
	}
	ruta := fmt.Sprintf("/productos_carrito/%d", item.IDProductosCarrito)
	verificarCantidad := func(esperada int32) {
		t.Helper()
		var producto, carrito, cantidad int32
		if err := lectura.QueryRow("SELECT id_producto, id_carrito, cantidad FROM productos_carrito WHERE id_productos_carrito=$1", item.IDProductosCarrito).Scan(&producto, &carrito, &cantidad); err != nil {
			t.Fatal(err)
		}
		if producto != productoID || carrito != carritoID || cantidad != esperada {
			t.Fatalf("fila inesperada: producto=%d carrito=%d cantidad=%d", producto, carrito, cantidad)
		}
	}
	verificarCantidad(2)
	detalle := pedir("GET", ruta, "", 200)
	var obtenido db.ProductosCarrito
	if err := json.Unmarshal(detalle.Body.Bytes(), &obtenido); err != nil || obtenido != item {
		t.Fatalf("detalle incorrecto: %+v, %v", obtenido, err)
	}
	lista := pedir("GET", "/productos_carrito", "", 200)
	var items []db.ProductosCarrito
	if err := json.Unmarshal(lista.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	encontrado := false
	for _, i := range items {
		if i == item {
			encontrado = true
		}
	}
	if !encontrado {
		t.Fatal("el listado no incluye el item creado")
	}
	actualizado := pedir("PUT", ruta, `{"cantidad":3}`, 200)
	if err := json.Unmarshal(actualizado.Body.Bytes(), &obtenido); err != nil || obtenido.Cantidad != 3 || obtenido.IDProducto != productoID || obtenido.IDCarrito != carritoID {
		t.Fatalf("actualizacion incorrecta: %+v, %v", obtenido, err)
	}
	verificarCantidad(3)
	for _, caso := range []struct {
		nombre, metodo, ruta, cuerpo string
		estado                       int
	}{
		{"producto cero", "POST", "/productos_carrito", fmt.Sprintf(`{"id_producto":0,"id_carrito":%d,"cantidad":1}`, carritoID), 400},
		{"producto negativo", "POST", "/productos_carrito", fmt.Sprintf(`{"id_producto":-1,"id_carrito":%d,"cantidad":1}`, carritoID), 400},
		{"carrito cero", "POST", "/productos_carrito", fmt.Sprintf(`{"id_producto":%d,"id_carrito":0,"cantidad":1}`, productoID), 400},
		{"producto inexistente", "POST", "/productos_carrito", fmt.Sprintf(`{"id_producto":2147483647,"id_carrito":%d,"cantidad":1}`, carritoID), 400},
		{"carrito inexistente", "POST", "/productos_carrito", fmt.Sprintf(`{"id_producto":%d,"id_carrito":2147483647,"cantidad":1}`, productoID), 400},
		{"cantidad cero", "POST", "/productos_carrito", fmt.Sprintf(`{"id_producto":%d,"id_carrito":%d,"cantidad":0}`, productoID, carritoID), 400},
		{"cantidad negativa", "PUT", ruta, `{"cantidad":-1}`, 400},
		{"cantidad omitida", "PUT", ruta, `{}`, 400},
		{"cantidad desbordada", "PUT", ruta, `{"cantidad":2147483648}`, 400},
		{"JSON invalido alta", "POST", "/productos_carrito", `{`, 400},
		{"JSON invalido actualizacion", "PUT", ruta, `{`, 400},
		{"ID cero", "GET", "/productos_carrito/0", "", 400},
		{"ID desbordado", "GET", "/productos_carrito/2147483648", "", 400},
		{"ID texto", "GET", "/productos_carrito/abc", "", 400},
		{"metodo elemento", "PATCH", ruta, "", 405},
		{"metodo coleccion", "DELETE", "/productos_carrito", "", 405},
	} {
		t.Run(caso.nombre, func(t *testing.T) { pedir(caso.metodo, caso.ruta, caso.cuerpo, caso.estado) })
	}
	verificarCantidad(3)
	pedir("DELETE", ruta, "", 204)
	var cantidad int
	if err := lectura.QueryRow("SELECT count(*) FROM productos_carrito WHERE id_productos_carrito=$1", item.IDProductosCarrito).Scan(&cantidad); err != nil || cantidad != 0 {
		t.Fatalf("borrado no persistido: %d, %v", cantidad, err)
	}
	pedir("GET", ruta, "", 404)
	pedir("PUT", ruta, `{"cantidad":1}`, 404)
	pedir("DELETE", ruta, "", 404)
	lista = pedir("GET", "/productos_carrito", "", 200)
	if strings.TrimSpace(lista.Body.String()) != "[]" {
		t.Fatalf("listado vacio = %s", lista.Body.String())
	}
}
