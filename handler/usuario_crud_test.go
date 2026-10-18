package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestUsuarioCRUD(t *testing.T) {
	h := handlerConDB(t)
	lectura := abrirDBTest(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/usuarios", h.Coleccion)
	mux.HandleFunc("/usuarios/{id}", h.Elemento)
	pedir := func(metodo, ruta, cuerpo string, esperado int) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		mux.ServeHTTP(res, req)
		if res.Code != esperado {
			t.Fatalf("%s %s: status=%d, esperado=%d: %s", metodo, ruta, res.Code, esperado, res.Body.String())
		}
		if (strings.Contains(res.Header().Get("Content-Type"), "application/json") && strings.Contains(res.Body.String(), "contrasena")) || strings.Contains(res.Body.String(), "secreto123") || strings.Contains(res.Body.String(), "nuevaclave123") {
			t.Fatalf("respuesta expone contrasena: %s", res.Body.String())
		}
		return res
	}
	email := fmt.Sprintf("crud-%d@example.com", time.Now().UnixNano())
	nuevoEmail := "actualizado-" + email
	duplicadoEmail := "duplicado-" + email
	t.Cleanup(func() {
		if _, err := lectura.Exec("DELETE FROM usuario WHERE email IN ($1,$2,$3)", email, nuevoEmail, duplicadoEmail); err != nil {
			t.Error(err)
		}
	})
	res := pedir("POST", "/usuarios", fmt.Sprintf(`{"email":"  %s  ","contrasena":"secreto123"}`, strings.ToUpper(email)), 201)
	var usuario struct {
		ID    int32  `json:"id_usuario"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &usuario); err != nil {
		t.Fatal(err)
	}
	if usuario.ID <= 0 || usuario.Email != email {
		t.Fatalf("usuario inesperado: %+v", usuario)
	}
	ruta := fmt.Sprintf("/usuarios/%d", usuario.ID)
	verificarPersistencia := func(emailEsperado, claveEsperada string) {
		t.Helper()
		var emailGuardado, claveGuardada string
		if err := lectura.QueryRow("SELECT email, contrasena FROM usuario WHERE id_usuario=$1", usuario.ID).Scan(&emailGuardado, &claveGuardada); err != nil {
			t.Fatal(err)
		}
		if emailGuardado != emailEsperado || claveGuardada != claveEsperada {
			t.Fatal("la fila persistida no coincide con los datos esperados")
		}
	}
	verificarPersistencia(email, "secreto123")
	detalle := pedir("GET", ruta, "", 200)
	var obtenido struct {
		ID    int32  `json:"id_usuario"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(detalle.Body.Bytes(), &obtenido); err != nil || obtenido != usuario {
		t.Fatalf("detalle inesperado: %+v, %v", obtenido, err)
	}
	lista := pedir("GET", "/usuarios", "", 200)
	var usuarios []struct {
		ID    int32  `json:"id_usuario"`
		Email string `json:"email"`
	}
	if err := json.Unmarshal(lista.Body.Bytes(), &usuarios); err != nil {
		t.Fatal(err)
	}
	encontrado := false
	for _, u := range usuarios {
		if u == usuario {
			encontrado = true
		}
	}
	if !encontrado {
		t.Fatal("el listado no incluye al usuario creado")
	}
	res = pedir("PUT", ruta, fmt.Sprintf(`{"email":"  %s  ","contrasena":"nuevaclave123"}`, strings.ToUpper(nuevoEmail)), 200)
	if err := json.Unmarshal(res.Body.Bytes(), &obtenido); err != nil || obtenido.ID != usuario.ID || obtenido.Email != nuevoEmail {
		t.Fatalf("actualizacion inesperada: %+v, %v", obtenido, err)
	}
	verificarPersistencia(nuevoEmail, "nuevaclave123")
	pedir("POST", "/usuarios", fmt.Sprintf(`{"email":"%s","contrasena":"secreto123"}`, duplicadoEmail), 201)
	for _, caso := range []struct {
		nombre, metodo, ruta, cuerpo string
		estado                       int
	}{
		{"alta email duplicado", "POST", "/usuarios", fmt.Sprintf(`{"email":"%s","contrasena":"secreto123"}`, nuevoEmail), 409},
		{"actualizacion email duplicado", "PUT", ruta, fmt.Sprintf(`{"email":"%s","contrasena":"secreto123"}`, duplicadoEmail), 409},
		{"email invalido", "PUT", ruta, `{"email":"invalido","contrasena":"secreto123"}`, 400},
		{"clave corta", "PUT", ruta, fmt.Sprintf(`{"email":"%s","contrasena":"corta"}`, nuevoEmail), 400},
		{"JSON invalido", "PUT", ruta, `{`, 400},
		{"ID cero", "GET", "/usuarios/0", "", 400},
		{"ID negativo", "GET", "/usuarios/-1", "", 400},
		{"ID texto", "GET", "/usuarios/abc", "", 400},
		{"ID desbordado", "DELETE", "/usuarios/4294967297", "", 400},
		{"metodo elemento", "PATCH", ruta, "", 405},
		{"metodo coleccion", "DELETE", "/usuarios", "", 405},
	} {
		t.Run(caso.nombre, func(t *testing.T) { pedir(caso.metodo, caso.ruta, caso.cuerpo, caso.estado) })
	}
	verificarPersistencia(nuevoEmail, "nuevaclave123")
	pedir("DELETE", ruta, "", 204)
	var cantidad int
	if err := lectura.QueryRow("SELECT count(*) FROM usuario WHERE id_usuario=$1", usuario.ID).Scan(&cantidad); err != nil || cantidad != 0 {
		t.Fatalf("borrado no persistido: %d, %v", cantidad, err)
	}
	pedir("GET", ruta, "", 404)
	pedir("PUT", ruta, fmt.Sprintf(`{"email":"%s","contrasena":"secreto123"}`, email), 404)
	pedir("DELETE", ruta, "", 404)
	if _, err := lectura.Exec("DELETE FROM usuario WHERE email=$1", duplicadoEmail); err != nil {
		t.Fatal(err)
	}
	lista = pedir("GET", "/usuarios", "", 200)
	if strings.TrimSpace(lista.Body.String()) != "[]" {
		t.Fatalf("listado vacio = %s", lista.Body.String())
	}
}
