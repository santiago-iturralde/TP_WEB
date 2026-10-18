# Trabajo de Cursada - Parte 3: Exponiendo la Lógica como API REST

Este repositorio contiene la implementación de la **Parte 3 (TP3)** de la materia. Consiste en la construcción de una **API REST HTTP** con operaciones CRUD para Producto, Usuario, Carrito y Productos_carrito en un sistema de e-commerce de perfumes, manteniendo la arquitectura en capas y la persistencia en PostgreSQL mediante `sqlc`.

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

* **Carrito**:
  * `id_carrito`: Identificador autogenerado.
  * `id_usuario`: Referencia opcional a un usuario existente.
  * `total`: Importe entre 0 y 99999999.99; por defecto 0.

* **ProductosCarrito**: Ítems seleccionados con `id_productos_carrito`, `id_producto`, `id_carrito` y `cantidad`.

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

### Carrito

Carrito sigue la misma arquitectura `handler → service → repository → sqlc` que Producto.

| Método | Endpoint | Descripción | Código de éxito |
| :--- | :--- | :--- | :--- |
| POST | `/carritos` | Crea un carrito. | 201 |
| GET | `/carritos` | Lista los carritos (array JSON, `[]` si no hay). | 200 |
| GET | `/carritos/{id}` | Obtiene un carrito por ID. | 200 |
| PUT | `/carritos/{id}` | Reemplaza usuario y total. | 200 |
| DELETE | `/carritos/{id}` | Elimina un carrito. | 204 |

POST y PUT reciben JSON, por ejemplo `{"id_usuario": 1, "total": 1500.50}`.
El usuario debe existir; también se permite omitir `id_usuario` o enviar `null`,
como admite el esquema actual. Si se omite `total`, se usa 0. En PUT, omitir
el usuario lo deja en NULL y omitir el total lo deja en 0.
El total se devuelve como número JSON, siguiendo el mapeo del precio de Producto.
El usuario opcional conserva el formato generado por sqlc (`{"Int32":1,"Valid":true}`).
Se devuelve 400 ante JSON o IDs inválidos, total negativo o fuera de rango,
o usuario inexistente; 404 si no existe el carrito; 405 para métodos no permitidos;
y 500 ante errores internos.

El total del carrito se recibe en el JSON. El CRUD de ítems no recalcula ese total
ni modifica el stock de Producto.

### Productos_carrito

La entidad usa las mismas capas y herramientas que Producto y Carrito.

| Método | Endpoint | Descripción | Código de éxito |
| :--- | :--- | :--- | :--- |
| POST | `/productos_carrito` | Agrega un ítem. | 201 |
| GET | `/productos_carrito` | Lista todos los ítems (`[]` si no hay). | 200 |
| GET | `/productos_carrito/{id}` | Obtiene un ítem por su ID. | 200 |
| PUT | `/productos_carrito/{id}` | Actualiza la cantidad del ítem. | 200 |
| DELETE | `/productos_carrito/{id}` | Elimina un ítem. | 204 |

POST recibe `{"id_producto": 1, "id_carrito": 1, "cantidad": 2}`.
Ambas referencias deben existir y los tres campos deben ser enteros positivos.
PUT recibe `{"cantidad": 3}` y mantiene el producto y el carrito del ítem,
respetando la consulta SQL existente. El ID de la URL identifica el registro
`id_productos_carrito`, no el producto ni el carrito.

Se devuelve 400 ante datos inválidos o referencias inexistentes, 404 si el ítem
no existe, 405 para métodos no permitidos y 500 ante errores internos.
`requests.sh` crea un producto y un carrito, prueba el CRUD de ítems y elimina
los datos utilizados al finalizar correctamente.

### Usuario

Usuario expone el CRUD con las mismas capas que las demás entidades.

| Método | Endpoint | Descripción | Código de éxito |
| :--- | :--- | :--- | :--- |
| POST | `/usuarios` | Crea un usuario. | 201 |
| GET | `/usuarios` | Lista los usuarios (`[]` si no hay). | 200 |
| GET | `/usuarios/{id}` | Obtiene un usuario por ID. | 200 |
| PUT | `/usuarios/{id}` | Actualiza email y contraseña. | 200 |
| DELETE | `/usuarios/{id}` | Elimina un usuario. | 204 |

POST y PUT reciben `{"email":"persona@example.com","contrasena":"secreto123"}`.
El email se normaliza a minúsculas y se quitan los espacios de los extremos.
La contraseña debe tener al menos 8 caracteres. PUT requiere ambos campos.
Las respuestas JSON incluyen únicamente `id_usuario` y `email`.
Se devuelve 400 ante datos o IDs inválidos, 404 si no existe el usuario,
409 si el email ya está registrado y 500 ante errores internos.
Al eliminar un usuario se aplica la cascada definida en el esquema para sus carritos.

El formulario y el alta originales en `/usuario` continúan disponibles.

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
Antes de ejecutar las pruebas, detené con `Ctrl+C` el servidor que hayas iniciado con `make run` y asegurate de que el puerto `8080` esté libre. Las pruebas inician su propio servidor en ese puerto.

```bash
make test-api
```

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
   * Repite el ciclo CRUD completo para Carrito, Productos_carrito y Usuario.
4. **Tareas Posteriores**: Detiene el proceso del servidor Go y destruye los contenedores y volúmenes de prueba (`docker compose down -v`).

---

## 6. Pruebas Unitarias e Integración con Go (`make test`)

Para ejecutar los tests de integración escritos con el paquete `testing` de Go:
```bash
make test
```
Los tests de Usuario, Carrito y Productos_carrito verifican el CRUD HTTP, la persistencia
desde otra conexión, las validaciones, las referencias y los códigos de error.

Este comando levanta la base de datos de pruebas, ejecuta la suite de `go test ./...` para verificar la lógica de repositorios y servicios, y limpia el entorno al finalizar.


### Estructura de Archivos Clave:
* `main.go`: Punto de entrada del sistema. Conecta la base de datos, inicializa repositorios, servicios y handlers, y registra los endpoints en `http.ServeMux`.
* `handler/productos_carrito.go`, `service/productos_carrito.go`, `repository/productos_carrito.go`: CRUD de ítems del carrito.
* `handler/carrito.go`, `service/carrito.go`, `repository/carrito.go`: CRUD de Carrito.
* `handler/producto.go`: Handler HTTP para los endpoints de `/productos`.
* `service/producto.go`: Lógica de negocio y validación de productos.
* `repository/producto.go`: Capa de datos para productos utilizando `sqlc`.
* `db/schema/schema.sql`: Definición de tablas en PostgreSQL.
* `db/queries/queries.sql`: Consultas SQL utilizadas por `sqlc`.
* `requests.sh`: Script ejecutable con peticiones `curl` de demostración/prueba.
* `scripts/test-api.sh`: Script de automatización de entorno y ejecución de tests API.
