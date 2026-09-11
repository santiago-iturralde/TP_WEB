# Trabajo de Cursada: Catálogo de Productos (E-commerce)

## Descripción del Proyecto
El objetivo del proyecto es una aplicación web para la gestión de un catálogo de productos de e-commerce, con un panel de administración para realizar operaciones CRUD sobre artículos de inventario. Actualmente, la aplicación implementa el formulario de alta de usuarios con persistencia en PostgreSQL; el esquema también define productos y carritos.

**Entidad** Producto  
**Atributos:** idProducto, Nombre, Descripción, Precio, Stock, Categoría

**Entidad** Carrito  
**Atributos:** idCarrito, idUsuario, total

**Entidad** ProductosCarrito  
**Atributos:** idProductosCarrito, idProducto, idCarrito, cantidad

**Entidad** Usuario   
**Atributos:** idUsuario, email, contraseña

## Requisitos del entorno

- Go 1.25 o superior, Make y Docker con Docker Compose compatible con `up --wait`.
- Acceso a Internet para descargar módulos de Go, sqlc y la imagen de PostgreSQL.

## Desarrollo

Desde la raíz del repositorio, opcionalmente copiá `.env.example` a `.env`
para personalizar las credenciales y los puertos. Make usa los valores del ejemplo
si no existe `.env`.

```bash
cp .env.example .env
```

```bash
make db-init  # Solo la primera vez: levanta PostgreSQL y crea las tablas
make db-up # Las demas veces que se quiera levantar la base de datos de DESARROLLO
make db-down # Baja el contenedor de DESARROLLO
make run     # Levanta la base y ejecuta la aplicación
```

Abrí http://localhost:8080/usuario. La ruta raíz `/` redirige automáticamente a `/usuario`.
Detené la aplicación con Ctrl+C y la base con `make db-down`.
Este último comando conserva los datos. `make db-init` se ejecuta una sola vez
por volumen nuevo: el esquema no es una migración y falla si las tablas ya existen.

La base de desarrollo usa el proyecto Compose `tp_web`, el volumen
`tp_web_db_perfumes_data` y el puerto 5432 por defecto. Se conserva el nombre
habitual del proyecto de esta carpeta para reutilizar su volumen existente.
Si antes usabas otro nombre de proyecto Compose, sus datos permanecen en su
volumen anterior; estos comandos no los migran automáticamente.

## Compilación y tests

Después de clonar, compilá y ejecutá los tests desde la raíz:

```bash
make build && make test
```

`make build` genera el código con sqlc y compila la aplicación en `tmp/TP_WEB`
sin levantar contenedores. `&&` ejecuta los tests solo si la compilación termina
correctamente.

`make test` también depende de `build`, por lo que vuelve a generar y compilar;
puede ejecutarse por separado. Luego elimina el entorno
de tests anterior, levanta PostgreSQL, espera hasta 60 segundos, carga el esquema
y ejecuta las pruebas con el paquete `testing` de Go. Al finalizar, también ante
un error o una interrupción normal, elimina los contenedores y volúmenes de tests.

Las pruebas usan el proyecto Compose `tpweb-test`, el volumen
`tpweb-test_test_data`, la base `perfumes_test` y el puerto 5433. Se puede cambiar
con `make test TEST_DB_PORT=5434`. Este entorno es independiente de desarrollo:
la limpieza de tests no elimina los datos de desarrollo. No ejecutes dos instancias
de `make test` simultáneamente, porque comparten el entorno de pruebas.

Go se ejecuta en la máquina local y PostgreSQL en Docker. El script proporciona
`DATABASE_URL` y las variables `DB_*` apuntando a la base de tests.
Los tests usan el repositorio real y verifican la persistencia en PostgreSQL desde
una conexión independiente: alta de usuario, rechazo de datos inválidos sin insertar
filas y rechazo de emails duplicados. También comprueban el formulario y que la
respuesta no exponga la contraseña. El script configura `TEST_DATABASE_URL`;
sin esa variable o sin conexión a `perfumes_test`, los tests fallan explícitamente.

## Organización del código

El flujo de alta es `UsuarioHandler → UsuarioService → UsuarioRepository`.
El handler interpreta las peticiones y genera las respuestas HTTP; el servicio
normaliza el email, valida los datos y coordina el alta; el repositorio ejecuta
las consultas en PostgreSQL.

## Documentación de persistencia

El [documento del proyecto](Mi%20primera%20aplicacion%20web-1.pdf) describe
las entidades, relaciones y consultas. El esquema está en `db/schema/schema.sql`,
las consultas en `db/queries/queries.sql` y el código generado en `db/sqlc/`.
