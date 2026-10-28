const formulario = document.querySelector("#formulario-producto");
const listaProductos = document.querySelector("#lista-productos");
const mensajeFormulario = document.querySelector("#mensaje-formulario");
const mensajeCatalogo = document.querySelector("#mensaje-catalogo");
const cantidadProductos = document.querySelector("#cantidad-productos");
const botonEnviar = formulario.querySelector('button[type="submit"]');

const formatoPrecio = new Intl.NumberFormat("es-AR", {
  style: "currency",
  currency: "ARS",
  maximumFractionDigits: 2,
});

function mostrarMensaje(elemento, texto, tipo = "") {
  elemento.textContent = texto;
  elemento.className = `mensaje ${tipo}`.trim();
}

function actualizarCantidad() {
  const total = listaProductos.querySelectorAll(".tarjeta-producto").length;
  cantidadProductos.textContent = `${total} ${total === 1 ? "producto" : "productos"}`;
}

function mostrarCatalogoVacio() {
  const estadoVacio = document.createElement("div");
  estadoVacio.className = "estado-vacio";

  const titulo = document.createElement("strong");
  titulo.textContent = "Todavía no hay perfumes";

  const texto = document.createElement("span");
  texto.textContent = "Usá el formulario para agregar el primero al catálogo.";

  estadoVacio.append(titulo, texto);
  listaProductos.replaceChildren(estadoVacio);
  actualizarCantidad();
}

function obtenerDescripcion(producto) {
  const descripcion = producto.descripcion_prod;
  if (typeof descripcion === "string") return descripcion;
  if (descripcion && descripcion.Valid) return descripcion.String;
  return "";
}

function crearTarjeta(producto) {
  const tarjeta = document.createElement("article");
  tarjeta.className = "tarjeta-producto";
  tarjeta.dataset.id = producto.id_producto;

  const parteSuperior = document.createElement("div");
  parteSuperior.className = "tarjeta-superior";

  const categoria = document.createElement("span");
  categoria.className = "etiqueta-categoria";
  categoria.textContent = producto.categoria;
  parteSuperior.append(categoria);

  const nombre = document.createElement("h3");
  nombre.textContent = producto.nombre_prod;

  const descripcion = document.createElement("p");
  descripcion.className = "descripcion-producto";
  descripcion.textContent = obtenerDescripcion(producto) || "Una fragancia especial para sumar a tu colección.";

  const pie = document.createElement("div");
  pie.className = "tarjeta-pie";

  const datos = document.createElement("div");
  const precio = document.createElement("p");
  precio.className = "precio-producto";
  precio.textContent = formatoPrecio.format(producto.precio_prod);

  const stock = document.createElement("p");
  stock.className = "stock-producto";
  stock.textContent = `${producto.stock_prod} unidades disponibles`;
  datos.append(precio, stock);

  const botonBorrar = document.createElement("button");
  botonBorrar.className = "boton-borrar";
  botonBorrar.type = "button";
  botonBorrar.textContent = "Eliminar";
  botonBorrar.setAttribute("aria-label", `Eliminar ${producto.nombre_prod}`);
  botonBorrar.addEventListener("click", () => eliminarProducto(producto.id_producto, tarjeta, botonBorrar));

  pie.append(datos, botonBorrar);
  tarjeta.append(parteSuperior, nombre, descripcion, pie);
  return tarjeta;
}

function agregarTarjeta(producto) {
  listaProductos.querySelector(".estado-vacio")?.remove();
  listaProductos.append(crearTarjeta(producto));
  actualizarCantidad();
}

async function cargarProductos() {
  mostrarMensaje(mensajeCatalogo, "Cargando productos...");
  listaProductos.replaceChildren();

  try {
    const respuesta = await fetch("/productos");
    if (!respuesta.ok) {
      throw new Error(`No se pudo cargar el catálogo (HTTP ${respuesta.status}).`);
    }

    const productos = await respuesta.json();
    mostrarMensaje(mensajeCatalogo, "");

    if (productos.length === 0) {
      mostrarCatalogoVacio();
      return;
    }

    productos.forEach(agregarTarjeta);
  } catch (error) {
    cantidadProductos.textContent = "Catálogo no disponible";
    mostrarMensaje(mensajeCatalogo, `${error.message} Verificá que el servidor esté funcionando.`, "error");
  }
}

formulario.addEventListener("submit", async (evento) => {
  evento.preventDefault();
  mostrarMensaje(mensajeFormulario, "");
  botonEnviar.disabled = true;
  botonEnviar.textContent = "Guardando...";

  const datos = new FormData(formulario);
  const productoNuevo = {
    nombre_prod: datos.get("nombre_prod").trim(),
    descripcion_prod: datos.get("descripcion_prod").trim(),
    precio_prod: Number(datos.get("precio_prod")),
    stock_prod: Number(datos.get("stock_prod")),
    categoria: datos.get("categoria").trim(),
  };

  try {
    const respuesta = await fetch("/productos", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(productoNuevo),
    });

    if (!respuesta.ok) {
      const detalle = (await respuesta.text()).trim();
      throw new Error(detalle || `No se pudo crear el producto (HTTP ${respuesta.status}).`);
    }

    const productoCreado = await respuesta.json();
    agregarTarjeta(productoCreado);
    formulario.reset();
    mostrarMensaje(mensajeFormulario, "El perfume se agregó al catálogo.", "exito");
  } catch (error) {
    mostrarMensaje(mensajeFormulario, error.message, "error");
  } finally {
    botonEnviar.disabled = false;
    botonEnviar.innerHTML = 'Agregar al catálogo <span aria-hidden="true">→</span>';
  }
});

async function eliminarProducto(id, tarjeta, boton) {
  boton.disabled = true;
  boton.textContent = "Eliminando...";

  try {
    const respuesta = await fetch(`/productos/${encodeURIComponent(id)}`, { method: "DELETE" });
    if (!respuesta.ok) {
      const detalle = (await respuesta.text()).trim();
      throw new Error(detalle || `No se pudo eliminar el producto (HTTP ${respuesta.status}).`);
    }

    tarjeta.remove();
    if (!listaProductos.querySelector(".tarjeta-producto")) {
      mostrarCatalogoVacio();
    } else {
      actualizarCantidad();
    }
    mostrarMensaje(mensajeCatalogo, "El producto se eliminó del catálogo.", "exito");
  } catch (error) {
    mostrarMensaje(mensajeCatalogo, error.message, "error");
    boton.disabled = false;
    boton.textContent = "Eliminar";
  }
}

cargarProductos();
