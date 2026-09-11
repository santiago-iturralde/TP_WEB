package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	db "TP_WEB/db/sqlc"
	"TP_WEB/repository"
	"TP_WEB/service"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func separarTest(t *testing.T) {
	t.Helper()
	linea := strings.Repeat("=", 72)
	t.Logf("\n\n%s\nTEST: %s\n%s", linea, t.Name(), linea)
	t.Cleanup(func() {
		t.Logf("\n%s\n", linea)
	})
}

// Cada llamada abre un pool independiente: las lecturas verifican datos confirmados.
func abrirDBTest(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("falta TEST_DATABASE_URL: ejecutar make test para levantar PostgreSQL de tests")
	}
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("abrir PostgreSQL de tests: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var name string
	if err := database.QueryRowContext(ctx, "SELECT current_database()").Scan(&name); err != nil {
		t.Fatalf("conectar a PostgreSQL de tests: %v", err)
	}
	if name != "perfumes_test" {
		t.Fatalf("se requiere perfumes_test, se recibio %q", name)
	}
	return database
}

func handlerConDB(t *testing.T) *UsuarioHandler {
	t.Helper()
	return NewUsuarioHandler(service.NewUsuarioService(repository.NewUsuarioRepository(db.New(abrirDBTest(t)))), "../static/usuario.html")
}

func contarUsuarios(t *testing.T, database *sql.DB) int {
	t.Helper()
	var count int
	if err := database.QueryRow("SELECT count(*) FROM usuario").Scan(&count); err != nil {
		t.Fatalf("contar usuarios: %v", err)
	}
	return count
}

func TestCrearUsuarioDesdeFormulario(t *testing.T) {
	separarTest(t)
	h := handlerConDB(t)
	lectura := abrirDBTest(t)
	email := "formulario-" + time.Now().Format("20060102150405.000000000") + "@example.com"
	t.Cleanup(func() {
		if _, err := lectura.Exec("DELETE FROM usuario WHERE email = $1", email); err != nil {
			t.Errorf("limpiar usuario: %v", err)
		}
	})
	form := url.Values{"email": {"  " + strings.ToUpper(email) + "  "}, "contrasena": {"secreto123"}}
	req := httptest.NewRequest(http.MethodPost, "/usuario", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()

	h.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("status = %d, se esperaba %d; respuesta: %s", res.Code, http.StatusCreated, res.Body.String())
	}

	var usuarioGuardado struct {
		ID    int32  `json:"id_usuario"`
		Email string `json:"email"`
	}
	respuesta := res.Body.String()
	if err := json.NewDecoder(strings.NewReader(respuesta)).Decode(&usuarioGuardado); err != nil {
		t.Fatalf("la respuesta no contiene un usuario valido: %v", err)
	}
	var persistido db.Usuario
	if err := lectura.QueryRow("SELECT id_usuario, email, contrasena FROM usuario WHERE email = $1", email).
		Scan(&persistido.IDUsuario, &persistido.Email, &persistido.Contrasena); err != nil {
		t.Fatalf("leer usuario persistido desde otra conexion: %v", err)
	}
	if persistido.IDUsuario <= 0 || usuarioGuardado.ID != persistido.IDUsuario || usuarioGuardado.Email != persistido.Email || persistido.Email != email {
		t.Fatalf("respuesta y fila persistida no coinciden: respuesta=%+v, id=%d, email=%s", usuarioGuardado, persistido.IDUsuario, persistido.Email)
	}
	if persistido.Contrasena != "secreto123" {
		t.Fatal("la contrasena persistida no coincide")
	}
	t.Logf("persistencia verificada en PostgreSQL desde otra conexion: id_usuario=%d, email=%s", persistido.IDUsuario, persistido.Email)

	if strings.Contains(respuesta, "secreto123") {
		t.Fatal("la respuesta expuso la contrasena")
	}
}

func TestCrearUsuarioRechazaDatosInvalidos(t *testing.T) {
	separarTest(t)
	h := handlerConDB(t)
	lectura := abrirDBTest(t)
	antes := contarUsuarios(t, lectura)
	for _, caso := range []struct {
		nombre string
		cuerpo string
	}{
		{"email invalido", `{"email":"invalido","contrasena":"secreto123"}`},
		{"contrasena corta", `{"email":"valido@example.com","contrasena":"corta"}`},
	} {
		t.Run(caso.nombre, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/usuario", strings.NewReader(caso.cuerpo))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			if res.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, se esperaba %d", res.Code, http.StatusBadRequest)
			}
			if despues := contarUsuarios(t, lectura); despues != antes {
				t.Fatalf("datos invalidos alteraron la DB: antes=%d, despues=%d", antes, despues)
			}
		})
	}
}

func TestMostrarFormulario(t *testing.T) {
	separarTest(t)
	h := handlerConDB(t)
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

func TestCrearUsuarioRechazaEmailDuplicado(t *testing.T) {
	separarTest(t)
	h := handlerConDB(t)
	lectura := abrirDBTest(t)
	email := "duplicado-" + time.Now().Format("20060102150405.000000000") + "@example.com"
	t.Cleanup(func() {
		if _, err := lectura.Exec("DELETE FROM usuario WHERE email = $1", email); err != nil {
			t.Errorf("limpiar usuario: %v", err)
		}
	})
	for _, esperado := range []int{http.StatusCreated, http.StatusConflict} {
		form := url.Values{"email": {email}, "contrasena": {"secreto123"}}
		req := httptest.NewRequest(http.MethodPost, "/usuario", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res := httptest.NewRecorder()
		h.ServeHTTP(res, req)
		if res.Code != esperado {
			t.Fatalf("status = %d, se esperaba %d; respuesta: %s", res.Code, esperado, res.Body.String())
		}
	}
	var count int
	if err := lectura.QueryRow("SELECT count(*) FROM usuario WHERE email = $1", email).Scan(&count); err != nil {
		t.Fatalf("consultar email duplicado: %v", err)
	}
	if count != 1 {
		t.Fatalf("filas persistidas para el mismo email = %d, se esperaba 1", count)
	}
}
