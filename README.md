# Trabajo de Cursada - Parte 3: Exponiendo la Lógica como API REST

Este repositorio contiene la implementación de la **Parte 3 (TP3)** de la materia. Consiste en la construcción de una **API REST HTTP** completa para la gestión de productos en un sistema de e-commerce de perfumes, manteniendo la arquitectura en capas y la persistencia en PostgreSQL mediante `sqlc`.

---

## 1. Dominio y Entidades del Sistema

El sistema gestiona cuatro entidades principales dentro del esquema de PostgreSQL:

* **Producto** (Entidad principal del TP3):
  * `id_producto`: Identificador único autogenerado (Serial/PK).
  * `nombre_prod`: Nombre del producto/perfume.
  * `descripcion_prod`: Descripción detallada.
  * `precio_prod`: Precio unitario (Decimal).
  * `stock_prod`: Cantidad de unidades disponibles en inventario.
  * `categoria`: Clasificación del producto.

* **Usuario**:
  * `id_usuario`: Identificador único (Serial/PK).
  * `email`: Correo electrónico (Único).
  * `contrasena`: Clave de acceso.

* **Carrito** y **ProductosCarrito**:
  * Tablas relacionales para la gestión de compras e ítems seleccionados.

---

## 2. Documentación de Endpoints (API REST)

La API REST expone las operaciones CRUD sobre la entidad `Producto` en la ruta `/productos`, respetando las convenciones HTTP:

| Método | Endpoint | Descripción | Código Éxito | Códigos Error |
| :--- | :--- | :--- | :--- | :--- |
| **POST** | `/productos` | Crea un nuevo producto recibiendo JSON. | `201 Created` | `400 Bad Request` |
| **GET** | `/productos` | Retorna la lista completa de productos en JSON. | `200 OK` | `500 Internal Error` |
| **GET** | `/productos/{id}` | Obtiene el detalle de un producto por su ID. | `200 OK` | `404 Not Found`, `400 Bad Request` |
| **PUT** | `/productos/{id}` | Actualiza todos los datos de un producto por su ID. | `200 OK` | `404 Not Found`, `400 Bad Request` |
| **DELETE** | `/productos/{id}` | Elimina el producto especificado por su ID. | `204 No Content` | `404 Not Found`, `400 Bad Request` |

### Rutas Adicionales del Proyecto:
* `GET /usuario`: Muestra el formulario HTML de registro de usuarios.
* `POST /usuario`: Procesa el alta de usuarios (acepta Form o JSON).
* `GET /`: Redirige automáticamente a `/usuario`.

---

## 3. Requisitos del Entorno

Para compilar, ejecutar y probar el proyecto es necesario contar con:
* **Go** 1.25 o superior.
* **Docker** y **Docker Compose** (con soporte para `--wait`).
* **Make** (herramienta de automatización).
* **curl** (para la ejecución de los tests de la API).

---

## 4. Guía Paso a Paso para Ejecutar el Proyecto

### Paso 1: Clonar y configurar variables de entorno
Copiar el archivo de ejemplo `.env.example` a `.env` (si se desean modificar los puertos o credenciales por defecto):
```bash
cp .env.example .env
```

### Paso 2: Inicializar la Base de Datos de Desarrollo
Ejecutar el siguiente comando **solo la primera vez** para levantar PostgreSQL en Docker y crear el esquema de tablas:
```bash
make db-init
```

### Paso 3: Iniciar el Servidor Web
Para levantar la base de datos y la aplicación en modo desarrollo:
```bash
make run
```
La aplicación quedará escuchando en `http://localhost:8080`.

### Paso 4: Detener el Entorno de Desarrollo
Para apagar el contenedor de PostgreSQL conservando los datos:
```bash
make db-down
```

---

## 5. Pruebas Automatizadas de la API REST (`make test-api`)

El proyecto incluye un script de automatización integral que realiza la prueba E2E (End-to-End) de la API REST en un entorno Docker totalmente aislado.

### Para ejecutar el test completo de la API:
```bash
make test-api
```
*(O directamente `./scripts/test-api.sh`)*

### ¿Qué realiza automáticamente `make test-api`?
1. **Tareas Previas**: Genera el código con `sqlc`, compila el binario en `tmp/TP_WEB`, elimina contenedores o volúmenes previos de test, levanta un contenedor aislado de PostgreSQL en el puerto `5433` y carga el esquema SQL (`db/schema/schema.sql`).
2. **Ejecución del Servidor**: Inicia la aplicación en segundo plano conectada a la base de pruebas.
3. **Prueba de Endpoints (`requests.sh`)**: Ejecuta peticiones `curl` probando el ciclo de vida completo:
   * Crea un producto (`POST`).
   * Consulta el catálogo completo (`GET`).
   * Consulta el producto creado por su ID (`GET /productos/{id}`).
   * Modifica el producto (`PUT`).
   * Borra el producto (`DELETE`).
   * Recomprueba que devuelva `404 Not Found` al intentar consultar el producto eliminado.
4. **Tareas Posteriores**: Detiene el proceso del servidor Go y destruye los contenedores y volúmenes de prueba (`docker compose down -v`).

---

## 6. Pruebas Unitarias e Integración con Go (`make test`)

Para ejecutar los tests de integración escritos con el paquete `testing` de Go:
```bash
make test
```
Este comando levanta la base de datos de pruebas, ejecuta la suite de `go test ./...` para verificar la lógica de repositorios y servicios, y limpia el entorno al finalizar.


### Estructura de Archivos Clave:
* `main.go`: Punto de entrada del sistema. Conecta la base de datos, inicializa repositorios, servicios y handlers, y registra los endpoints en `http.ServeMux`.
* `handler/producto.go`: Handler HTTP para los endpoints de `/productos`.
* `service/producto.go`: Lógica de negocio y validación de productos.
* `repository/producto.go`: Capa de datos para productos utilizando `sqlc`.
* `db/schema/schema.sql`: Definición de tablas en PostgreSQL.
* `db/queries/queries.sql`: Consultas SQL utilizadas por `sqlc`.
* `requests.sh`: Script ejecutable con peticiones `curl` de demostración/prueba.
* `scripts/test-api.sh`: Script de automatización de entorno y ejecución de tests API.