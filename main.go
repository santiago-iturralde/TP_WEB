package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	db "TP_WEB/db/sqlc"
	"TP_WEB/handler"
	"TP_WEB/repository"
	"TP_WEB/service"
)

func main() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("No se pudo abrir app.log: %s", err)
	}
	defer logFile.Close()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(os.Stderr, logFile), nil)))

	// Abre la conexion con la base de datos PostgreSQL usando la URL de conexión
	database, err := sql.Open("pgx", databaseURL())

	if err != nil {
		log.Fatalf("Error al configurar la base de datos: %s", err)
	}

	// Cierra la conexión a la base de datos al finalizar la ejecución del programa
	defer database.Close()

	// Crea las consultas SQL a partir de la conexión a la base de datos
	queries := db.New(database)
	usuarioRepo := repository.NewUsuarioRepository(queries)
	usuarioService := service.NewUsuarioService(usuarioRepo)
	usuarioHandler := handler.NewUsuarioHandler(usuarioService, "static/usuario.html")

	mux := http.NewServeMux()
	mux.Handle("/usuario", usuarioHandler)
	mux.HandleFunc("/usuario.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/usuario.css")
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, "/usuario", http.StatusFound)
	})

	port := ":8080"
	slog.Info("Servidor TP2 iniciado", "url", "http://localhost"+port+"/usuario")

	if err := http.ListenAndServe(port, mux); err != nil {
		slog.Error("Error al iniciar el servidor", "error", err)
	}
}

func databaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	user := envOrDefault("DB_USER", "postgres")
	password := envOrDefault("DB_PASSWORD", "postgres")
	name := envOrDefault("DB_NAME", "perfumes_db")
	port := envOrDefault("DB_PORT", "5432")
	host := envOrDefault("DB_HOST", "localhost")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, name)
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
