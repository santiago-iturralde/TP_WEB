# TP5 — Renderizado del lado del servidor con Templ

Trabajo de cursada de Programación Web, UNICEN.
El catálogo de perfumes se genera en Go con Templ y datos de PostgreSQL obtenidos
mediante SQLc. El listado y el formulario funcionan sin JavaScript, con recargas
completas de página. Se mantienen los estilos CSS del TP4.

## 1. Requisitos

- Git para clonar el repositorio.
- Go 1.25 o superior (con descarga automática del toolchain habilitada cuando una dependencia la requiera).
- Docker en ejecución y Docker Compose v2 con soporte de `up --wait`.
- Make, shell POSIX (`sh`) y curl.
- Internet durante la primera ejecución para descargar módulos Go e imagen PostgreSQL.
- Puertos 8080 (aplicación) y 5434 (PostgreSQL) disponibles.

No hace falta instalar SQLc, Templ, Air, Node ni PostgreSQL localmente.
Los generadores se ejecutan con `go run` y versiones fijadas en el Makefile:
SQLc `v1.31.1` y Templ `v0.3.960`.

## 2. Clonar y ejecutar la aplicación

```sh
git clone --branch tp5 https://github.com/santiago-iturralde/TP_WEB.git
cd TP_WEB
make run-app
```

Abrir **http://localhost:8080/**. No se necesita copiar `.env`, crear tablas
manualmente ni ejecutar comandos previos.

`make run-app` ejecuta `scripts/run-app.sh`, que:

1. Borra contenedores y volúmenes anteriores **del proyecto Compose `tpweb-tp5`**.
2. Genera código SQLc y Templ y compila el ejecutable en `tmp/TP_WEB`.
3. Levanta PostgreSQL 15 y espera a que su healthcheck confirme disponibilidad.
4. Carga `db/schema/schema.sql`, deteniéndose si el esquema falla.
5. Arranca el servidor Go, conectado explícitamente a esa base, y comprueba su respuesta.
6. Al terminar, fallar o recibir `Ctrl+C`, detiene el servidor y elimina los
   contenedores, la red y el volumen utilizados mediante un `trap`.

**Cada ejecución comienza con un catálogo vacío. Los perfumes agregados se guardan
mientras dura esa ejecución y se eliminan al cerrarla**. El volumen de desarrollo `tp_web` no se toca. Una finalización forzada
con SIGKILL o un apagado no permite ejecutar traps; el siguiente arranque limpia
los recursos anteriores antes de comenzar. Al interrumpir con `Ctrl+C`, Make puede
mostrar `Error 130`: es el código de interrupción, no un fallo de la limpieza.

Para usar otros puertos:

```sh
HTTP_PORT=8081 DELIVERY_DB_PORT=5435 make run-app
```

No ejecutar dos instancias de `make run-app` simultáneas: comparten el proyecto Compose
`tpweb-tp5`. Si un puerto está ocupado, cerrar el proceso correspondiente o elegir
otro; el script no termina procesos ajenos.

## 3. Qué probar en el navegador

1. Abrir `/`: aparece el formulario y el mensaje de catálogo vacío.
2. Completar nombre, descripción opcional, precio, stock y categoría.
3. Enviar: el servidor guarda el producto y vuelve a `/`, donde aparece su tarjeta.
4. Recargar: el producto sigue guardado y no se repite el alta.
5. Agregar otro producto y comprobar el contador y la lista.
6. Detener con `Ctrl+C`: se elimina la base de esta ejecución.

El navegador valida campos requeridos y límites básicos. El servidor también
valida, sin depender del navegador. Si rechaza el formulario, devuelve HTML con
estado 400, un mensaje y los valores ingresados. Los errores internos usan 500.

## 4. Arquitectura y flujo SSR

```text
GET /
  -> ProductoHandler.Index
  -> ProductoService.Listar
  -> ProductoRepository -> SQLc -> PostgreSQL
  -> views.IndexPage(productos, datosFormulario, mensaje)
  -> Render(contexto, respuesta)
  -> HTML completo al navegador
```

```text
POST /productos (application/x-www-form-urlencoded)
  -> ParseForm y conversión de los campos numéricos
  -> validación y ProductoService.Crear
  -> repositorio -> SQLc -> INSERT en PostgreSQL
  -> 303 See Other, Location: /
  -> el navegador hace GET / y recibe el catálogo actualizado
```

Este último recorrido es **Post/Redirect/Get (PRG)**. El HTML lo genera el
servidor; el navegador aplica CSS y presenta la página. No se utiliza `fetch`,
JavaScript ni una API JSON para construir la interfaz.

Los `.templ` son fuentes editables. `templ generate` produce los `_templ.go`,
que Go compila con la aplicación; no se editan manualmente ni se generan por
cada petición. En cada petición se ejecutan las funciones Go ya compiladas.

## 5. Archivos y responsabilidades

| Archivo o carpeta | Responsabilidad |
| --- | --- |
| `main.go` | Conexión PostgreSQL, composición de capas y registro de rutas. |
| `views/layout.templ` | Documento HTML, encabezado, CSS, contenido hijo y pie. |
| `views/index.templ` | Composición de la página del catálogo. |
| `views/entity_list.templ` | Tarjetas, contador y estado vacío a partir de `[]db.Producto`. |
| `views/entity_form.templ` | Formulario clásico y mensajes de validación. |
| `handler/producto.go` | GET principal y renderizado; despacho del POST según Content-Type. |
| `handler/producto_formulario.go` | Lectura de campos, conversiones y PRG. |
| `service/` | Validaciones y reglas de negocio. |
| `repository/` | Acceso a datos mediante consultas SQLc. |
| `db/schema/`, `db/queries/`, `db/sqlc/` | Esquema, consultas SQL y código generado. |
| `static/estilos.css` | Diseño del catálogo; único recurso estático de la interfaz. |
| `scripts/run-app.sh` | Preparación, ejecución y limpieza de la aplicación. |
| `docker-compose.delivery.yml` | Configuración de PostgreSQL utilizada por `make run-app`. |

**Alcance de la interfaz:** listado y alta de productos. No incluye una pantalla
de carrito, edición o borrado desde la interfaz. Se retiraron las páginas estáticas
anteriores y `app.js`; la interfaz funciona sin JavaScript.

El dominio original (Producto, Usuario, Carrito y ProductosCarrito), sus consultas,
servicios y pruebas de regresión se conserva. Los endpoints JSON `/productos`,
`/usuarios`, `/carritos` y `/productos_carrito` y sus rutas `/{id}` continúan como
API de compatibilidad, pero la pantalla SSR no los consulta por JavaScript.
El POST `/productos` reconoce el formulario HTML; los clientes JSON conservan
su respuesta JSON. `/usuario` y `/static/index.html` ya no son pantallas disponibles.

## 6. Pruebas automatizadas

```sh
make check-delivery
```

Ejecuta el **mismo ciclo que `make run-app`** y `scripts/check-ssr.sh`: catálogo vacío,
CSS disponible, formulario, alta HTTP, redirección 303, producto en HTML, escape
de texto, rechazo de un precio inválido y conservación de los campos. Verifica
que las páginas retiradas no se sirvan. Al terminar elimina la base utilizada.
No ejecutar mientras `make run-app` esté activo, porque usan el mismo entorno.

```sh
make test
make test-api
```

- `make test`: genera, compila y ejecuta las pruebas Go con PostgreSQL real,
  incluyendo el alta SSR y comprobación de persistencia desde otra conexión.
- `make test-api`: prueba por HTTP el CRUD y validaciones de las cuatro entidades.
  Usa el puerto HTTP 18081 (configurable mediante `TEST_HTTP_PORT`).
- Ambas utilizan el proyecto `tpweb-test`, la base `perfumes_test` y el puerto
  PostgreSQL 5433 (`TEST_DB_PORT`). Crean el esquema y eliminan sus volúmenes
  al terminar. Ejecutarlas una después de otra, no simultáneamente.

## 7. Desarrollo opcional con datos persistentes

Para trabajar conservando los datos entre ejecuciones:

```sh
cp .env.example .env  # opcional: ajustar credenciales y puerto de desarrollo
make db-init         # solo si la base de desarrollo todavía no tiene tablas
make run
```

`make run` genera SQLc y Templ, levanta la base de desarrollo y ejecuta Go.
`make db-down` elimina el contenedor de desarrollo sin borrar su volumen.
No usar `db-init` sobre un esquema ya creado.

Air es opcional. Con la base iniciada y las variables de conexión del entorno
configuradas, `air` observa `.go`, `.sql` y `.templ`, ejecuta `make build` y reinicia
el servidor. Excluye `_templ.go` para evitar recompilaciones por archivos generados.
El navegador se recarga manualmente. No ejecutar Air y `make run` al mismo tiempo.
