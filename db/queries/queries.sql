
-- CONSULTAS PARA LA ENTIDAD: PRODUCTO--------------------------------------------------------
-- name: CreateProducto :one
INSERT INTO producto (nombre_prod, descripcion_prod, precio_prod, stock_prod, categoria)
VALUES ($1, $2, $3, $4, $5)
RETURNING id_producto, nombre_prod, descripcion_prod, precio_prod, stock_prod, categoria;

-- name: GetProducto :one
SELECT id_producto, nombre_prod, descripcion_prod, precio_prod, stock_prod, categoria
FROM producto
WHERE id_producto = $1;

-- name: ListProductos :many
SELECT id_producto, nombre_prod, descripcion_prod, precio_prod, stock_prod, categoria
FROM producto
ORDER BY id_producto;

-- name: UpdateProducto :exec
UPDATE producto
SET nombre_prod = $2,
    descripcion_prod = $3,
    precio_prod = $4,
    stock_prod = $5,
    categoria = $6
WHERE id_producto = $1;

-- name: DeleteProducto :exec
DELETE FROM producto
WHERE id_producto = $1;




-- CONSULTAS PARA LA ENTIDAD: USUARIO-----------------------------------------------------
-- name: CreateUsuario :one
INSERT INTO usuario (email, contrasena)
VALUES ($1, $2)
RETURNING id_usuario, email, contrasena;

-- name: GetUsuario :one
SELECT id_usuario, email, contrasena
FROM usuario
WHERE id_usuario = $1;

-- name: GetUsuarioByEmail :one
SELECT id_usuario, email, contrasena
FROM usuario
WHERE email = $1;

-- name: DeleteUsuario :exec
DELETE FROM usuario
WHERE id_usuario = $1;



-- CONSULTAS PARA LA ENTIDAD: CARRITO-----------------------------------------------------------------
-- name: CreateCarrito :one
INSERT INTO carrito (id_usuario, total)
VALUES ($1, $2)
RETURNING id_carrito, id_usuario, total;

-- name: GetCarrito :one
SELECT id_carrito, id_usuario, total
FROM carrito
WHERE id_carrito = $1;

-- name: ListCarritos :many
SELECT id_carrito, id_usuario, total
FROM carrito
ORDER BY id_carrito;

-- name: UpdateCarrito :exec
UPDATE carrito
SET id_usuario = $2,
    total = $3
WHERE id_carrito = $1;

-- name: DeleteCarrito :exec
DELETE FROM carrito
WHERE id_carrito = $1;



-- CONSULTAS PARA LA ENTIDAD: PRODUCTOS_CARRITO----------------------------------------------------------------------

