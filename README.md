# Trabajo de Cursada: Catálogo de Productos (E-commerce)

## Descripción del Proyecto
Este proyecto es una aplicación web para la gestión de un catálogo de productos de e-commerce [1]. Se enfoca en un panel de administración que permite realizar operaciones CRUD sobre artículos de inventario.

**Entidad Principal:** Producto  
**Atributos:** ID, Nombre, Descripción, Precio, Stock y Categoría.

## Requisitos del Entorno
Para el correcto funcionamiento de esta aplicación, se requiere:
* **Sistema Operativo:** Linux (Recomendado: Mint, Debian, Arch, etc.) [2].
* **Lenguaje:** Go (Golang) [1, 3].
* **Herramientas de red:** curl (para pruebas de cabeceras) [4].

## Estructura de Archivos
* `main.go`: Servidor web principal desarrollado en Go [5].
* `/static`: Directorio que contiene los archivos estáticos [6].
  * `index.html`: Página de presentación del proyecto [5].

## Instrucciones de Ejecución
Siga estos pasos para iniciar el servidor en su entorno local:

1. Abra una terminal en la carpeta raíz del proyecto [2].
2. Ejecute el siguiente comando para iniciar el servidor:
   ```bash
   go run main.go
