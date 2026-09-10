package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	db "TP_WEB/db/sqlc"
)

type usuarioRepoFalso struct {
	usuario    db.Usuario
	contrasena string
	err        error
}

func (f *usuarioRepoFalso) Crear(_ context.Context, email, contrasena string) (db.Usuario, error) {
	f.contrasena = contrasena
	f.usuario.Email = email
	return f.usuario, f.err
}

func (f *usuarioRepoFalso) BuscarPorEmail(context.Context, string) (db.Usuario, error) {
	return db.Usuario{}, nil
}

func TestCrearUsuarioDesdeFormulario(t *testing.T) {
	repo := &usuarioRepoFalso{usuario: db.Usuario{IDUsuario: 7}}
	h := NewUsuarioHandler(repo, "../static/usuario.html")
	form := url.Values{"email": {"PERSONA@EXAMPLE.COM"}, "contrasena": {"secreto123"}}
	req := httptest.NewRequest(http.MethodPost, "/usuario", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba %d; respuesta: %s", res.Code, http.StatusCreated, res.Body.String())
	}
	if repo.usuario.Email != "persona@example.com" {
		t.Fatalf("email guardado = %q", repo.usuario.Email)
	}
	if repo.contrasena != "secreto123" {
		t.Fatalf("contrasena guardada = %q", repo.contrasena)
	}

	var usuarioGuardado struct {
		ID    int32  `json:"id_usuario"`
		Email string `json:"email"`
	}
	respuesta := res.Body.String()
	if err := json.NewDecoder(strings.NewReader(respuesta)).Decode(&usuarioGuardado); err != nil {
		t.Fatalf("la respuesta no contiene un usuario valido: %v", err)
	}
	if usuarioGuardado.ID != repo.usuario.IDUsuario || usuarioGuardado.Email != repo.usuario.Email {
		t.Fatalf("usuario retornado = %+v; usuario guardado = %+v", usuarioGuardado, repo.usuario)
	}

	t.Logf("usuario guardado en la DB: id_usuario=%d, email=%s", usuarioGuardado.ID, usuarioGuardado.Email)

	if strings.Contains(respuesta, "secreto123") {
		t.Fatal("la respuesta expuso la contrasena")
	}
}

func TestCrearUsuarioRechazaDatosInvalidos(t *testing.T) {
	repo := &usuarioRepoFalso{}
	h := NewUsuarioHandler(repo, "../static/usuario.html")
	req := httptest.NewRequest(http.MethodPost, "/usuario", strings.NewReader(`{"email":"invalido","contrasena":"corta"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, se esperaba %d", res.Code, http.StatusBadRequest)
	}
}

func TestMostrarFormulario(t *testing.T) {
	h := NewUsuarioHandler(&usuarioRepoFalso{}, "../static/usuario.html")
	req := httptest.NewRequest(http.MethodGet, "/usuario", nil)
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, se esperaba %d", res.Code, http.StatusOK)
	}
	if !strings.Contains(res.Body.String(), `action="/usuario"`) {
		t.Fatal("la respuesta no contiene el formulario de usuario")
	}
}
