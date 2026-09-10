package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	db "TP_WEB/db/sqlc"
	"TP_WEB/handler"
	"TP_WEB/repository"
)

func main() {
	database, err := sql.Open("pgx", databaseURL())
	if err != nil {
		log.Fatalf("Error al configurar la base de datos: %s", err)
	}
	defer database.Close()

	queries := db.New(database)
	usuarioRepo := repository.NewUsuarioRepository(queries)
	usuarioHandler := handler.NewUsuarioHandler(usuarioRepo, "static/usuario.html")

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

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		http.ServeFile(w, r, "static/index.html")
	})

	port := ":8080"
	fmt.Printf("Servidor TP2 escuchando en http://localhost%s\n", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Printf("Error al iniciar el servidor: %s\n", err)
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
