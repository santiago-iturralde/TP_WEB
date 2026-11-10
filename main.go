package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	db "TP_WEB/db/sqlc"
	"TP_WEB/handler"
	"TP_WEB/repository"
	"TP_WEB/service"
)

func main() {
	dsn, err := databaseURL()
	if err != nil {
		log.Fatalf("Error de configuración: %s", err)
	}

	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("No se pudo abrir app.log: %s", err)
	}
	defer logFile.Close()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(os.Stderr, logFile), nil)))

	// Abre la conexion con la base de datos PostgreSQL usando la URL de conexión
	database, err := sql.Open("pgx", dsn)

	if err != nil {
		log.Fatalf("Error al configurar la base de datos: %s", err)
	}

	// Cierra la conexión a la base de datos al finalizar la ejecución del programa
	defer database.Close()

	// Crea las consultas SQL a partir de la conexión a la base de datos
	queries := db.New(database)

	// Instanciación de Usuario
	usuarioRepo := repository.NewUsuarioRepository(queries)
	usuarioService := service.NewUsuarioService(usuarioRepo)
	usuarioHandler := handler.NewUsuarioHandler(usuarioService)

	// Instanciación de Producto
	productoRepo := repository.NewProductoRepository(queries)
	productoService := service.NewProductoService(productoRepo)
	productoHandler := handler.NewProductoHandler(productoService)

	// Instanciación de Carrito
	carritoRepo := repository.NewCarritoRepository(queries)
	carritoService := service.NewCarritoService(carritoRepo)
	carritoHandler := handler.NewCarritoHandler(carritoService)

	// Instanciación de ProductosCarrito
	productosCarritoRepo := repository.NewProductosCarritoRepository(queries)
	productosCarritoService := service.NewProductosCarritoService(productosCarritoRepo)
	productosCarritoHandler := handler.NewProductosCarritoHandler(productosCarritoService)

	mux := http.NewServeMux()
	mux.HandleFunc("/usuarios", usuarioHandler.Coleccion)
	mux.HandleFunc("/usuarios/{id}", usuarioHandler.Elemento)
	mux.HandleFunc("GET /static/estilos.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/estilos.css")
	})

	mux.HandleFunc("/productos_carrito", productosCarritoHandler.Coleccion)
	mux.HandleFunc("/productos_carrito/{id}", productosCarritoHandler.Elemento)

	mux.HandleFunc("/carritos", carritoHandler.Coleccion)
	mux.HandleFunc("/carritos/{id}", carritoHandler.Elemento)

	mux.HandleFunc("/productos", productoHandler.Coleccion)
	mux.HandleFunc("/productos/{id}", productoHandler.Elemento)

	mux.HandleFunc("GET /{$}", productoHandler.Index)

	port := ":" + envOrDefault("HTTP_PORT", "8080")
	slog.Info("Servidor TP5 SSR iniciado", "url", "http://localhost"+port+"/")

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Error al iniciar el servidor: %v", err)
	}
}

func databaseURL() (string, error) {
	if value := strings.TrimSpace(os.Getenv("DATABASE_URL")); value != "" {
		return value, nil
	}

	for _, key := range []string{"DB_USER", "DB_PASSWORD", "DB_NAME"} {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			return "", fmt.Errorf("falta %s: configurá .env y ejecutá mediante Make", key)
		}
	}

	connection := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")),
		Host:     net.JoinHostPort(envOrDefault("DB_HOST", "localhost"), envOrDefault("DB_PORT", "5432")),
		Path:     "/" + os.Getenv("DB_NAME"),
		RawQuery: "sslmode=disable",
	}
	return connection.String(), nil
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
