package handler

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
	"TP_WEB/service"
)

func TestLongitudProductoCompartida(t *testing.T) {
	for _, caracter := range []string{"a", "ñ"} {
		t.Run(caracter, func(t *testing.T) {
			database := abrirDBTest(t)
			lectura := abrirDBTest(t)
			h := NewProductoHandler(service.NewProductoService(repository.NewProductoRepository(db.New(database))))
			valido := service.CrearProductoDTO{
				Nombre: strings.Repeat(caracter, 255), Descripcion: strings.Repeat(caracter, 1000),
				Categoria: strings.Repeat(caracter, 100), Precio: 19.99, Stock: 3,
			}
			t.Cleanup(func() {
				if _, err := database.Exec("DELETE FROM producto WHERE nombre_prod=$1 OR nombre_prod=$2", valido.Nombre, valido.Nombre+caracter); err != nil {
					t.Error(err)
				}
			})
			enviar := func(metodo string, formulario bool, dto service.CrearProductoDTO, id int32) *httptest.ResponseRecorder {
				body, err := json.Marshal(dto)
				if err != nil {
					t.Fatal(err)
				}
				contentType := "application/json"
				if formulario {
					body = []byte(url.Values{
						"nombre_prod": {dto.Nombre}, "descripcion_prod": {dto.Descripcion},
						"categoria": {dto.Categoria}, "precio_prod": {"19.99"}, "stock_prod": {"3"},
					}.Encode())
					contentType = "application/x-www-form-urlencoded"
				}
				req := httptest.NewRequest(metodo, "/productos", strings.NewReader(string(body)))
				req.Header.Set("Content-Type", contentType)
				res := httptest.NewRecorder()
				if metodo == "PUT" {
					req.SetPathValue("id", fmt.Sprint(id))
					h.Elemento(res, req)
				} else {
					h.Coleccion(res, req)
				}
				return res
			}
			res := enviar("POST", false, valido, 0)
			if res.Code != 201 {
				t.Fatalf("alta en límite: %d %s", res.Code, res.Body.String())
			}
			var creado db.Producto
			if err := json.Unmarshal(res.Body.Bytes(), &creado); err != nil {
				t.Fatal(err)
			}
			for _, entrada := range []struct {
				metodo     string
				formulario bool
				estado     int
			}{
				{"PUT", false, 200}, {"POST", true, 303},
			} {
				res = enviar(entrada.metodo, entrada.formulario, valido, creado.IDProducto)
				if res.Code != entrada.estado {
					t.Fatalf("límite válido: %d %s", res.Code, res.Body.String())
				}
			}
			var cantidadAntes int
			if err := lectura.QueryRow("SELECT count(*) FROM producto").Scan(&cantidadAntes); err != nil {
				t.Fatal(err)
			}
			for _, campo := range []string{"nombre", "categoria", "descripcion"} {
				invalido := valido
				var texto string
				switch campo {
				case "nombre":
					invalido.Nombre += caracter
					texto = invalido.Nombre
				case "categoria":
					invalido.Categoria += caracter
					texto = invalido.Categoria
				case "descripcion":
					invalido.Descripcion += caracter
					texto = invalido.Descripcion
				}
				for _, entrada := range []struct {
					metodo     string
					formulario bool
				}{
					{"POST", false}, {"PUT", false}, {"POST", true},
				} {
					res = enviar(entrada.metodo, entrada.formulario, invalido, creado.IDProducto)
					if res.Code != 400 || !strings.Contains(res.Body.String(), service.ErrLongitudInvalida.Error()) {
						t.Fatalf("%s %s formulario=%v: %d %s", campo, entrada.metodo, entrada.formulario, res.Code, res.Body.String())
					}
					if entrada.formulario && !strings.Contains(res.Body.String(), texto) {
						t.Fatal("el formulario no conserva el texto ingresado")
					}
				}
			}
			var cantidadDespues int
			if err := lectura.QueryRow("SELECT count(*) FROM producto").Scan(&cantidadDespues); err != nil || cantidadDespues != cantidadAntes {
				t.Fatalf("un alta inválida modificó la cantidad: antes=%d después=%d error=%v", cantidadAntes, cantidadDespues, err)
			}
			var nombre, descripcion, categoria string
			if err := lectura.QueryRow("SELECT nombre_prod, descripcion_prod, categoria FROM producto WHERE id_producto=$1", creado.IDProducto).Scan(&nombre, &descripcion, &categoria); err != nil {
				t.Fatal(err)
			}
			if nombre != valido.Nombre || descripcion != valido.Descripcion || categoria != valido.Categoria {
				t.Fatal("una actualización inválida modificó el producto")
			}
		})
	}
}
