# Trabajo de Cursada: Catálogo de Productos (E-commerce)

## Descripción del Proyecto
Este proyecto es una aplicación web para la gestión de un catálogo de productos de e-commerce. Se enfoca en un panel de administración que permite realizar operaciones CRUD sobre artículos de inventario.

**Entidad** Producto  
**Atributos:** idProducto, Nombre, Descripción, Precio, Stock, Categoría

**Entidad** Carrito
**Atributos:** idCarrito, idProductosCarrito, total 

**Entidad** ProductosCarrito
**Atributos:** idProductosCarrito, idProducto, idCarrito, cantidad

**Entidad** Usuario 
**Atributos:** idUsuario, email, contraseña

## Requisitos del Entorno
Para el correcto funcionamiento de esta aplicación, se requiere:
* **Sistema Operativo:** Linux (Recomendado: Mint, Debian, Arch, etc.)
* **Lenguaje:** Go (Golang)
* **Herramientas de red:** curl (para pruebas de cabeceras)

## Estructura de Archivos
* `main.go`: Servidor web principal desarrollado en Go 
* `/static`: Directorio que contiene los archivos estáticos 
  * `index.html`: Página de presentación del proyecto 

## Instrucciones de Ejecución
Siga estos pasos para descargar e iniciar el servidor en su entorno local:

1. Clone el repositorio en su computadora ejecutando en la terminal:
   ```bash
   git clone https://github.com/santiago-iturralde/TP_WEB.git
  
2. Ingrese a la carpeta del proyecto:
   ```bash
   cd TP_WEB

3. Ejecute el siguiente comando para iniciar el servidor:
   ```bash
   go run main.go

4. Una vez iniciado, abra su navegador web y acceda a:
    http://localhost:8080

5. Para detener el servidor, regrese a la terminal y presione Ctrl + C