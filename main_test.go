package main

import (
	"net/url"
	"strings"
	"testing"
)

func TestDatabaseURLRequiresCredentials(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	for _, key := range []string{"DB_USER", "DB_PASSWORD", "DB_NAME"} {
		for _, value := range []string{"", "   "} {
			t.Run(key+"/"+value, func(t *testing.T) {
				t.Setenv("DB_USER", "usuario")
				t.Setenv("DB_PASSWORD", "clave-de-prueba")
				t.Setenv("DB_NAME", "catalogo")
				t.Setenv(key, value)
				dsn, err := databaseURL()
				if err == nil || !strings.Contains(err.Error(), "falta "+key) || dsn != "" {
					t.Fatalf("se esperaba error de configuración para %s", key)
				}
				if strings.Contains(err.Error(), "clave-de-prueba") {
					t.Fatal("el error expone la contraseña")
				}
			})
		}
	}
}

func TestDatabaseURLExplicitConnection(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://prueba:clave@localhost:5433/pruebas?sslmode=disable")
	for _, key := range []string{"DB_USER", "DB_PASSWORD", "DB_NAME"} {
		t.Setenv(key, "")
	}
	dsn, err := databaseURL()
	if err != nil || dsn != "postgres://prueba:clave@localhost:5433/pruebas?sslmode=disable" {
		t.Fatal("debe respetar la conexión explícita del entorno aislado")
	}
}

func TestDatabaseURLConfiguredCredentials(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_USER", "usuario")
	t.Setenv("DB_PASSWORD", "clave@:/?#")
	t.Setenv("DB_NAME", "catalogo")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	dsn, err := databaseURL()
	if err != nil {
		t.Fatal(err)
	}
	connection, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("la conexión no es una URL válida")
	}
	password, _ := connection.User.Password()
	if connection.User.Username() != "usuario" || password != "clave@:/?#" || connection.Host != "localhost:5432" || connection.Path != "/catalogo" || connection.Query().Get("sslmode") != "disable" {
		t.Fatal("la conexión debe conservar las credenciales y los defaults de host y puerto")
	}
}
