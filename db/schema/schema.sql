CREATE TABLE producto (
    id_producto SERIAL PRIMARY KEY,
    nombre_prod VARCHAR(255) NOT NULL,
    descripcion_prod TEXT,
    precio_prod DECIMAL(10, 2) NOT NULL,
    stock_prod INT NOT NULL DEFAULT 0,
    categoria VARCHAR(100) NOT NULL
)


CREATE TABLE usuario (
    id_usuario SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    contrasena VARCHAR(255) NOT NULL
);


CREATE TABLE carrito (
    id_carrito SERIAL PRIMARY KEY,
    id_usuario INT REFERENCES usuario(id_usuario) ON DELETE CASCADE,
    total DECIMAL(10, 2) NOT NULL DEFAULT 0.00
);


CREATE TABLE productos_carrito (
    id_productos_carrito SERIAL PRIMARY KEY,
    id_producto INT NOT NULL REFERENCES producto(id_producto) ON DELETE CASCADE,
    id_carrito INT NOT NULL REFERENCES carrito(id_carrito) ON DELETE CASCADE,
    cantidad INT NOT NULL DEFAULT 1
);
