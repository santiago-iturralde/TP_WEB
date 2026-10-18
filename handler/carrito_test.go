package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
	"TP_WEB/service"
)

// Comprueba el CRUD HTTP contra PostgreSQL y lee los cambios desde otra conexión.
func TestCarritoCRUD(t *testing.T) {
	database := abrirDBTest(t)
	lectura := abrirDBTest(t)
	h := NewCarritoHandler(service.NewCarritoService(repository.NewCarritoRepository(db.New(database))))
	mux := http.NewServeMux()
	mux.HandleFunc("/carritos", h.Coleccion)
	mux.HandleFunc("/carritos/{id}", h.Elemento)
	pedir := func(metodo, ruta, cuerpo string, esperado int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != esperado {
			t.Fatalf("%s %s: status=%d, esperado=%d: %s", metodo, ruta, res.Code, esperado, res.Body.String())
		}
		return res
	}
	email := fmt.Sprintf("carrito-%d@example.com", time.Now().UnixNano())
	var usuarioID int32
	if err := database.QueryRow("INSERT INTO usuario (email, contrasena) VALUES ($1, 'secreto123') RETURNING id_usuario", email).Scan(&usuarioID); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := database.Exec("DELETE FROM usuario WHERE id_usuario=$1", usuarioID); err != nil {
			t.Error(err)
		}
	})
	res := pedir("POST", "/carritos", fmt.Sprintf(`{"id_usuario":%d,"total":12.50}`, usuarioID), 201)
	var carrito db.Carrito
	if err := json.Unmarshal(res.Body.Bytes(), &carrito); err != nil {
		t.Fatal(err)
	}
	if carrito.IDCarrito <= 0 || carrito.IDUsuario.Int32 != usuarioID || !carrito.IDUsuario.Valid {
		t.Fatalf("carrito inesperado: %+v", carrito)
	}
	ruta := fmt.Sprintf("/carritos/%d", carrito.IDCarrito)
	var total float64
	if err := lectura.QueryRow("SELECT total FROM carrito WHERE id_carrito=$1", carrito.IDCarrito).Scan(&total); err != nil || total != 12.50 {
		t.Fatalf("alta persistida: total=%v, error=%v", total, err)
	}
	pedir("GET", ruta, "", 200)
	lista := pedir("GET", "/carritos", "", 200)
	var carritos []db.Carrito
	if err := json.Unmarshal(lista.Body.Bytes(), &carritos); err != nil {
		t.Fatal(err)
	}
	encontrado := false
	for _, c := range carritos {
		if c.IDCarrito == carrito.IDCarrito {
			encontrado = true
		}
	}
	if !encontrado {
		t.Fatal("el listado no incluye el carrito creado")
	}
	pedir("PUT", ruta, fmt.Sprintf(`{"id_usuario":%d,"total":25.75}`, usuarioID), 200)
	if err := lectura.QueryRow("SELECT total FROM carrito WHERE id_carrito=$1", carrito.IDCarrito).Scan(&total); err != nil || total != 25.75 {
		t.Fatalf("actualizacion persistida: total=%v, error=%v", total, err)
	}
	for _, caso := range []struct {
		metodo, ruta, cuerpo string
		estado               int
	}{
		{"POST", "/carritos", `{"total":-1}`, 400},
		{"POST", "/carritos", `{"id_usuario":0}`, 400},
		{"POST", "/carritos", `{"id_usuario":-1}`, 400},
		{"POST", "/carritos", `{"total":100000000}`, 400},
		{"POST", "/carritos", `{"id_usuario":2147483647}`, 400},
		{"POST", "/carritos", `{`, 400},
		{"PUT", ruta, `{"total":-1}`, 400},
		{"PUT", ruta, `{"id_usuario":2147483647}`, 400},
		{"PUT", ruta, `{`, 400},
		{"GET", "/carritos/0", "", 400},
		{"GET", "/carritos/2147483648", "", 400},
		{"GET", "/carritos/abc", "", 400},
		{"PATCH", ruta, "", 405},
		{"DELETE", "/carritos", "", 405},
	} {
		t.Run(caso.metodo+caso.ruta+caso.cuerpo, func(t *testing.T) { pedir(caso.metodo, caso.ruta, caso.cuerpo, caso.estado) })
	}
	if err := lectura.QueryRow("SELECT total FROM carrito WHERE id_carrito=$1", carrito.IDCarrito).Scan(&total); err != nil || total != 25.75 {
		t.Fatalf("peticiones invalidas cambiaron el total: %v, %v", total, err)
	}
	pedir("DELETE", ruta, "", 204)
	var cantidad int
	if err := lectura.QueryRow("SELECT count(*) FROM carrito WHERE id_carrito=$1", carrito.IDCarrito).Scan(&cantidad); err != nil || cantidad != 0 {
		t.Fatalf("borrado no persistido: %d, %v", cantidad, err)
	}
	pedir("GET", ruta, "", 404)
	pedir("PUT", ruta, `{"total":0}`, 404)
	pedir("DELETE", ruta, "", 404)
	// El esquema permite carritos sin usuario y un total inicial igual a cero.
	res = pedir("POST", "/carritos", `{}`, 201)
	if err := json.Unmarshal(res.Body.Bytes(), &carrito); err != nil {
		t.Fatal(err)
	}
	ruta = fmt.Sprintf("/carritos/%d", carrito.IDCarrito)
	t.Cleanup(func() { database.Exec("DELETE FROM carrito WHERE id_carrito=$1", carrito.IDCarrito) })
	var usuarioNulo bool
	if err := lectura.QueryRow("SELECT id_usuario IS NULL, total FROM carrito WHERE id_carrito=$1", carrito.IDCarrito).Scan(&usuarioNulo, &total); err != nil || !usuarioNulo || total != 0 {
		t.Fatalf("valores iniciales invalidos: usuarioNulo=%v, total=%v, error=%v", usuarioNulo, total, err)
	}
	pedir("DELETE", ruta, "", 204)
	lista = pedir("GET", "/carritos", "", 200)
	if strings.TrimSpace(lista.Body.String()) != "[]" {
		t.Fatalf("listado vacio = %s", lista.Body.String())
	}
}
