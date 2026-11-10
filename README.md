# Trabajo de Cursada - TP5: Renderizado del lado del servidor con Templ

Trabajo de Programación Web, UNICEN. Esta entrega adapta el catálogo de perfumes
del TP4 al renderizado del lado del servidor (SSR): Go obtiene los datos de
PostgreSQL mediante sqlc y genera el HTML con Templ. El listado y el formulario
funcionan sin JavaScript, con recargas completas de página y los estilos CSS del TP4.

## 1. Dominio y alcance de la entrega

El sistema conserva cuatro entidades:

- **Producto:** nombre, descripción opcional, precio, stock y categoría.
- **Usuario:** email y contraseña.
- **Carrito:** usuario asociado y total.
- **ProductosCarrito:** producto, carrito y cantidad de cada ítem.

La interfaz del TP5 permite **listar y crear productos**. Incluye un estado de
catálogo vacío, contador de productos y mensajes de validación. No incluye
pantallas de carrito, edición ni eliminación de productos.

Se conserva la arquitectura `handler → service → repository → sqlc → PostgreSQL`.
Los endpoints JSON `/productos`, `/usuarios`, `/carritos` y `/productos_carrito`,
con sus rutas `/{id}`, siguen disponibles para compatibilidad con entregas
anteriores. La pantalla SSR no los consulta mediante JavaScript.

## 2. Requisitos del entorno

- Git para clonar el repositorio.
- **Go 1.26 o superior** para ejecutar también el generador sqlc fijado en el proyecto.
  La aplicación declara Go 1.25 en `go.mod`; si se parte de esa versión, debe estar
  habilitada la descarga automática del toolchain para ejecutar sqlc.
- Docker en ejecución y Docker Compose v2 con soporte de `up --wait`.
- Make, shell POSIX (`sh`) y curl.
- Conexión a Internet durante la primera ejecución para descargar módulos,
  herramientas Go y la imagen de PostgreSQL.
- Los puertos del entorno elegido disponibles (ver sección 6).

No es necesario instalar sqlc ni Templ como comandos independientes: el Makefile
los ejecuta mediante `go run`, con versiones fijadas: sqlc `v1.31.1` y Templ
`v0.3.960`. Tampoco se necesita PostgreSQL local, Node ni Air para estos pasos.

## 3. Configuración: `.env.example`, `.env` y valores por defecto

### ¿Qué archivo se debe configurar?

`.env.example` es una **plantilla versionada**, con valores de ejemplo.
`.env` contiene la **configuración local** de quien ejecuta el proyecto y está
excluido de Git mediante `.gitignore`.

Para configurar la base de desarrollo, copiar la plantilla y editar la copia:

```sh
cp .env.example .env
```

Si ya existe `.env`, editarlo directamente para conservar su configuración.
También se puede crear manualmente con las mismas variables, pero copiar la
plantilla evita omitir nombres necesarios. No hace falta editar `.env.example`
para cambiar las credenciales de una ejecución local.

La plantilla actual contiene:

```dotenv
DB_USER=postgres
DB_PASSWORD=123456
DB_NAME=postgres-perfumes
DB_PORT=5432
TEST_DB_PORT=5433
```

| Variable | Uso |
| --- | --- |
| `DB_USER` | Usuario de PostgreSQL para desarrollo. |
| `DB_PASSWORD` | Contraseña de ese usuario. |
| `DB_NAME` | Nombre de la base de desarrollo. |
| `DB_PORT` | Puerto de PostgreSQL publicado en la computadora; dentro del contenedor sigue siendo 5432. |
| `TEST_DB_PORT` | Puerto publicado para PostgreSQL de pruebas; no cambia el puerto de desarrollo. |

El usuario puede elegir las credenciales y el puerto antes de inicializar una
base de desarrollo nueva. No son credenciales de una cuenta externa: Docker
utiliza estos valores para inicializar PostgreSQL.

### ¿Cómo llegan esas variables a Docker y Go?

El Makefile lee `.env` y exporta las variables `DB_*` a los comandos que ejecuta.
Por ejemplo, en `docker-compose.yml`:

```yaml
POSTGRES_USER: ${DB_USER}
POSTGRES_PASSWORD: ${DB_PASSWORD}
POSTGRES_DB: ${DB_NAME}
```

Compose sustituye cada expresión por su valor configurado. También puede leer
`.env` al ejecutarse directamente, pero los valores de respaldo del Makefile
solo se aplican cuando se usa Make. Para seguir esta guía, utilizar los comandos
`make` indicados.

**Go no carga `.env` por sí solo.** Recibe las variables exportadas por Make.
Por eso ejecutar `go run .` directamente puede usar una configuración distinta.
Si existe `DATABASE_URL` en el entorno, la aplicación la prioriza sobre `DB_*`;
para usar la configuración de desarrollo de esta guía, no dejar una URL de otro
entorno exportada.

### Configuración y valores de respaldo

La copia de `.env.example` a `.env` es un paso de preparación del entorno de
desarrollo. Make lee el archivo si existe; no comprueba su existencia ni bloquea
la compilación cuando falta. Antes de ejecutar `make db-init` o `make run`,
completar `DB_USER`, `DB_PASSWORD` y `DB_NAME` en `.env`.

**No hay credenciales de respaldo en Make ni en Go.** Solo se conservan defaults
para host y puertos: `localhost`, PostgreSQL de desarrollo `5432`, pruebas `5433`
y HTTP `8080`.

La compilación no necesita credenciales. Al ejecutar el binario, Go exige recibir
las tres variables de conexión o una `DATABASE_URL` explícita; si faltan, se
detiene con un mensaje de configuración. Go no lee `.env` por sí solo: usar Make
para cargarlo. Los scripts de entrega y pruebas proporcionan sus propias
conexiones aisladas y pueden ejecutarse sin `.env`.

### ¿Por qué el Compose de entrega tiene valores escritos directamente?

`docker-compose.delivery.yml` define el usuario `postgres`, la contraseña
`tp5_local` y la base `perfumes_tp5` como **valores fijos**. El script de entrega
utiliza exactamente esos mismos valores para cargar el esquema y conectar Go.
Esto prepara un entorno de entrega separado del desarrollo.

Cambiar `DB_USER`, `DB_PASSWORD` o `DB_NAME` en `.env` **no cambia las credenciales
de entrega**. Cambiar solo esos valores en el Compose de entrega tampoco alcanza:
la conexión y la carga del esquema en el script deben coincidir.

En cambio, `${DELIVERY_DB_PORT:-5434}` significa: usar `DELIVERY_DB_PORT` si está
definida y no está vacía; de lo contrario, usar `5434`. Los ejemplos de esta guía
pasan los puertos de entrega y HTTP por la línea de comandos para que lleguen
tanto al script como a Compose.

## 4. Guía paso a paso para ejecutar la entrega

### Paso 1: Clonar el proyecto

```sh
git clone --branch tp5 https://github.com/santiago-iturralde/TP_WEB.git
cd TP_WEB
```

Si el repositorio ya está clonado, ubicarse en su carpeta y en la rama `tp5`.

### Paso 2: Configurar e iniciar el entorno de entrega

Si todavía no existe `.env`, copiar la plantilla:

```sh
cp .env.example .env
```

Revisar las credenciales de la copia (sección 3). Esto deja preparado el entorno
de desarrollo persistente; el modo de entrega usa su propia base aislada y
no necesita esas credenciales. Para iniciar la entrega:


```sh
make run-app
```

No es necesario crear tablas manualmente; este modo utiliza la conexión aislada
de entrega.
`scripts/run-app.sh` realiza lo siguiente:

1. Elimina los contenedores y volúmenes anteriores del proyecto Compose `tpweb-tp5`.
2. Genera sqlc y Templ y compila el ejecutable en `tmp/TP_WEB`.
3. Inicia PostgreSQL 15 y espera a que esté disponible.
4. Carga `db/schema/schema.sql`, deteniéndose si falla.
5. Inicia Go conectado a la base de entrega y comprueba su respuesta HTTP.
6. Al finalizar, fallar o recibir `Ctrl+C`, detiene la aplicación y elimina los
   contenedores, la red y el volumen del entorno de entrega mediante un `trap`.

### Paso 3: Abrir la aplicación

Abrir [http://localhost:8080/](http://localhost:8080/).
Para elegir otros puertos:

```sh
HTTP_PORT=8081 DELIVERY_DB_PORT=5435 make run-app
```

En ese caso, abrir [http://localhost:8081/](http://localhost:8081/).

### Paso 4: Comprobar el flujo en el navegador

1. Verificar el formulario y el mensaje de catálogo vacío.
2. Completar nombre, precio, stock y categoría; la descripción es opcional.
3. Enviar el formulario: el producto aparece en el catálogo después de la redirección.
4. Recargar la página: el producto sigue visible y no se repite el alta.
5. Agregar otro producto y comprobar la lista y el contador.

El servidor valida los datos además de las restricciones del navegador. Ante
valores de campos inválidos, devuelve HTML con estado 400, un mensaje y los
valores ingresados. Los errores internos usan estado 500.
La validación compartida admite hasta 255 caracteres en nombre, 100 en categoría
y 1000 en descripción, tanto en el formulario como en el alta y actualización JSON.

### Paso 5: Detener la aplicación

Presionar `Ctrl+C` en la terminal.

**Cada ejecución de entrega comienza con un catálogo vacío. Los productos se
guardan durante esa ejecución y se eliminan al cerrarla.** El volumen del entorno
de desarrollo se conserva. Para guardar productos entre ejecuciones, usar el modo
de desarrollo de la sección siguiente.

Una finalización con SIGKILL o un apagado no permite ejecutar el `trap`; el
siguiente arranque limpia los recursos anteriores. Make puede mostrar `Error 130`
al interrumpir con `Ctrl+C`: corresponde a la interrupción.

## 5. Desarrollo con datos persistentes

### Paso 1: Configurar el entorno local

Desde la carpeta del repositorio, si todavía no existe `.env`:

```sh
cp .env.example .env
```

Editar `.env` con el usuario, contraseña, nombre de base y puerto deseados.
Este es el paso de configuración para ejecutar el entorno de desarrollo. Make
no lo comprueba previamente; no hay valores de respaldo para las credenciales.

### Paso 2: Inicializar la base de datos

Ejecutar **solo cuando la base todavía no tiene las tablas**:

```sh
make db-init
```

Este comando inicia PostgreSQL de desarrollo y carga el esquema SQL.
No ejecutarlo nuevamente sobre un esquema ya creado.

### Paso 3: Iniciar el servidor

```sh
make run
```

Genera sqlc y Templ, inicia la base de desarrollo y ejecuta Go.
Abrir [http://localhost:8080/](http://localhost:8080/).
Para cambiar el puerto HTTP:

```sh
HTTP_PORT=8081 make run
```

### Paso 4: Detener el entorno

Detener Go con `Ctrl+C` y luego ejecutar:

```sh
make db-down
```

Se elimina el contenedor de desarrollo **sin borrar el volumen**. En la próxima
ejecución de `make run`, los productos continúan guardados.

**Si la base ya fue inicializada, cambiar las credenciales en `.env` no modifica
los usuarios ni las contraseñas almacenados en PostgreSQL.** Se debe mantener la
configuración original o realizar el cambio dentro de la base y actualizar `.env`
de forma coherente. Cambiar `DB_NAME` tampoco crea otra base en un volumen ya
inicializado. No borrar un volumen con datos que se quieran conservar.

## 6. Entornos, puertos y persistencia

| Entorno | Comando | Archivo Compose | Proyecto Compose | Puerto PostgreSQL por defecto | Puerto HTTP por defecto | Datos |
| --- | --- | --- | --- | --- | --- | --- |
| Desarrollo | `make run` | `docker-compose.yml` | `tp_web` | 5432 (`DB_PORT`) | 8080 (`HTTP_PORT`) | Se conservan al detener. |
| Entrega | `make run-app` | `docker-compose.delivery.yml` | `tpweb-tp5` | 5434 (`DELIVERY_DB_PORT`) | 8080 (`HTTP_PORT`) | Se eliminan al finalizar. |
| Pruebas Go | `make test` | `docker-compose.test.yml` | `tpweb-test` | 5433 (`TEST_DB_PORT`) | No inicia servidor HTTP externo. | Se eliminan al finalizar. |
| Pruebas API | `make test-api` | `docker-compose.test.yml` | `tpweb-test` | 5433 (`TEST_DB_PORT`) | 18081 (`TEST_HTTP_PORT`) | Se eliminan al finalizar. |

Las pruebas usan el usuario `postgres`, la contraseña `test_password` y la base
`perfumes_test`, independientes de las credenciales de desarrollo.
`make check-delivery` comparte el entorno de entrega.

No ejecutar dos instancias del entorno de entrega simultáneamente, ni
`make check-delivery` mientras está activo `make run-app`: comparten recursos que
se eliminan al iniciar y al terminar. Ejecutar `make test` y `make test-api` uno
después del otro porque comparten `tpweb-test`. Desarrollo y entrega usan por
defecto el mismo puerto HTTP; elegir otro si se necesitan ambos servidores.

## 7. Templ, arquitectura y flujo SSR

Templ permite escribir componentes de HTML con datos y lógica de Go en archivos
`.templ`. En esta aplicación, el layout contiene la estructura general; el
formulario y el listado son componentes separados, reunidos por `IndexPage`.

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

## 8. Pruebas automatizadas

Las pruebas preparan sus propias bases aisladas y proporcionan las credenciales
necesarias. No requieren `.env` ni utilizan los datos de la base de desarrollo.

### Flujo de entrega y SSR

```sh
make check-delivery
```

Ejecuta el ciclo de preparación y limpieza de entrega y `scripts/check-ssr.sh`.
Comprueba catálogo vacío, CSS, formulario sin JavaScript, alta HTTP, redirección
303, producto en HTML, escape de texto, rechazo de un precio inválido y
conservación de campos. También verifica que las pantallas retiradas no se sirvan.
Al terminar elimina la base utilizada.

### Pruebas Go con PostgreSQL real

```sh
make test
```

Genera y compila el proyecto, prepara la base de pruebas y ejecuta
`go test -count=1 -v ./...`. Incluye pruebas del alta SSR, persistencia consultada
desde otra conexión, CRUD y validaciones. Limpia el entorno al finalizar.
Ejecutar `go test ./...` directamente no prepara esa base ni configura
`TEST_DATABASE_URL`; usar `make test` para la suite de integración.

### Pruebas HTTP de la API conservada

```sh
make test-api
```

Prepara la base de pruebas, inicia su propio servidor y ejecuta `requests.sh`
para comprobar CRUD y validaciones de Producto, Usuario, Carrito y
ProductosCarrito. Detiene el servidor y elimina los recursos de prueba al terminar.
Para cambiar su puerto HTTP:

```sh
TEST_HTTP_PORT=18082 make test-api
```

## 9. Estructura de archivos y rutas

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
| `Makefile` | Comandos de generación, ejecución, desarrollo y pruebas. |
| `.env.example` | Plantilla de configuración local. |
| `docker-compose.yml` | Base de desarrollo con volumen persistente. |
| `docker-compose.test.yml` | Base aislada para las pruebas Go y API. |
| `scripts/check-ssr.sh` | Comprobaciones HTTP de la interfaz SSR. |
| `scripts/test.sh`, `scripts/test-api.sh`, `requests.sh` | Preparación y ejecución de las suites de pruebas. |

| Método y ruta | Comportamiento |
| --- | --- |
| `GET /` | Devuelve la página completa con formulario y catálogo. |
| `POST /productos` con formulario | Valida, guarda y redirige a `/` con 303. |
| `POST /productos` con JSON | Conserva el alta de la API con respuesta JSON. |
| `GET /productos` | Conserva el listado JSON de la API. |
| `GET /static/estilos.css` | Sirve la hoja de estilos. |

`/usuario`, `/static/index.html` y `/static/app.js` ya no se sirven como parte de
la interfaz. Los archivos `_templ.go` y `db/sqlc/` son código generado a partir
de las vistas y las consultas SQL, respectivamente.

## 10. Problemas frecuentes y desarrollo opcional

- **Falta `.env` o una credencial al ejecutar desarrollo:** copiar `.env.example`
  a `.env` si no existe y completar `DB_USER`, `DB_PASSWORD` y `DB_NAME` antes de
  ejecutar `make db-init` o `make run`.

- **Puerto ocupado:** elegir otro puerto con la variable del entorno correspondiente
  (sección 6). El script de entrega no termina procesos ajenos.
- **Error de autenticación:** comprobar que se usa el entorno correcto y que las
  credenciales coinciden con las que inicializaron su volumen. Revisar también si
  hay una `DATABASE_URL` exportada que reemplace `DB_*` en desarrollo.
- **Tablas inexistentes en desarrollo:** ejecutar `make db-init` en una base sin
  esquema. En entrega, el script carga las tablas automáticamente.
- **Tablas ya existentes al ejecutar `db-init`:** la inicialización ya se realizó;
  iniciar con `make run` sin repetir ese paso.
- **Los productos desaparecen al cerrar:** es el comportamiento del modo entrega;
  usar desarrollo si se necesita persistencia entre ejecuciones.
- **Error de versión de Go al generar sqlc:** usar Go 1.26 o superior, o permitir
  que Go descargue automáticamente el toolchain requerido.

Air es opcional. Con la base iniciada y las variables de conexión exportadas,
`air` observa `.go`, `.sql` y `.templ`, ejecuta `make build` y reinicia el servidor.
La configuración excluye `_templ.go` para evitar recompilaciones causadas por
archivos generados. El navegador se recarga manualmente. No ejecutar Air y
`make run` al mismo tiempo.
