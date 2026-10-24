#!/bin/sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== 1. POST /productos (Crear producto) ==="
CREATE_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/productos" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre_prod": "Chanel No. 5",
    "descripcion_prod": "Perfume clasico elegante",
    "precio_prod": 15000.50,
    "stock_prod": 20,
    "categoria": "Fragancias"
  }')

STATUS=$(echo "$CREATE_RESP" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$CREATE_RESP" | grep -v "HTTP_STATUS:")

if [ "$STATUS" -ne 201 ]; then
  echo "Error en POST /productos. Status esperado: 201, obtenido: $STATUS"
  echo "Respuesta: $BODY"
  exit 1
fi

ID=$(echo "$BODY" | grep -o '"id_producto":[0-9]*' | head -n 1 | cut -d: -f2)
echo "Producto creado exitosamente con ID: $ID"
echo "$BODY"
echo ""

echo "=== 2. GET /productos (Listar productos) ==="
LIST_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/productos")
if [ "$LIST_STATUS" -ne 200 ]; then
  echo "Error en GET /productos. Status esperado: 200, obtenido: $LIST_STATUS"
  exit 1
fi
curl -s "$BASE_URL/productos"
echo ""
echo ""

echo "=== 3. GET /productos/$ID (Obtener producto por ID) ==="
GET_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/productos/$ID")
if [ "$GET_STATUS" -ne 200 ]; then
  echo "Error en GET /productos/$ID. Status esperado: 200, obtenido: $GET_STATUS"
  exit 1
fi
curl -s "$BASE_URL/productos/$ID"
echo ""
echo ""

echo "=== 4. PUT /productos/$ID (Actualizar producto) ==="
UPDATE_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X PUT "$BASE_URL/productos/$ID" \
  -H "Content-Type: application/json" \
  -d '{
    "nombre_prod": "Chanel No. 5 Redesign",
    "descripcion_prod": "Edicion especial 2026",
    "precio_prod": 18500.00,
    "stock_prod": 15,
    "categoria": "Fragancias de Lujo"
  }')

UPDATE_STATUS=$(echo "$UPDATE_RESP" | grep "HTTP_STATUS:" | cut -d: -f2)
UPDATE_BODY=$(echo "$UPDATE_RESP" | grep -v "HTTP_STATUS:")

if [ "$UPDATE_STATUS" -ne 200 ]; then
  echo "Error en PUT /productos/$ID. Status esperado: 200, obtenido: $UPDATE_STATUS"
  echo "Respuesta: $UPDATE_BODY"
  exit 1
fi
echo "Producto actualizado exitosamente:"
echo "$UPDATE_BODY"
echo ""

echo "=== 5. DELETE /productos/$ID (Eliminar producto) ==="
DEL_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/productos/$ID")
if [ "$DEL_STATUS" -ne 204 ]; then
  echo "Error en DELETE /productos/$ID. Status esperado: 204, obtenido: $DEL_STATUS"
  exit 1
fi
echo "Producto eliminado exitosamente (HTTP 204)."
echo ""

echo "=== 6. GET /productos/$ID (Verificar eliminacion -> 404) ==="
NOT_FOUND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/productos/$ID")
if [ "$NOT_FOUND_STATUS" -ne 404 ]; then
  echo "Error en comprobacion 404. Status esperado: 404, obtenido: $NOT_FOUND_STATUS"
  exit 1
fi
echo "Verificacion correcta: el producto borrado retorno HTTP 404."
echo ""

echo "=== TEST: Crear producto con valores invalidos ==="

probar_producto_invalido() {

    DESCRIPCION=$1
    DATOS_JSON=$2
    ERROR_ESPERADO=$3

    echo "=== Probando: $DESCRIPCION ==="

    RESP=$(curl -sS -w "\nHTTP_STATUS:%{http_code}" \
        -X POST "$BASE_URL/productos" \
        -H "Content-Type: application/json" \
        -d "$DATOS_JSON")

    STATUS=$(printf '%s\n' "$RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)

    BODY=$(printf '%s\n' "$RESP" | sed '/^HTTP_STATUS:/d')

    if [ "$STATUS" != "400" ]; then
        echo "ERROR: Esperabamos HTTP 400, recibimos $STATUS"
        exit 1
    fi

    case "$BODY" in
        *"$ERROR_ESPERADO"*)
            echo "OK: Error detectado correctamente"
            ;;
        *)
            echo "ERROR: El mensaje no coincide"
            echo "Esperado: $ERROR_ESPERADO"
            echo "Recibido: $BODY"
            exit 1
            ;;
    esac
}

probar_producto_invalido "precio negativo" \
    '{"nombre_prod":"Invictus","precio_prod":-100,"stock_prod":10,"categoria":"Perfumes"}' \
    'el precio debe ser un numero valido mayor a cero'

# Nombre vacio
probar_producto_invalido "nombre vacio" \
    '{"nombre_prod":"","precio_prod":100,"stock_prod":10,"categoria":"Perfumes"}' \
    "el nombre del producto no puede estar vacio"

# Stock negativo
probar_producto_invalido "stock negativo" \
    '{"nombre_prod":"Invictus","precio_prod":100,"stock_prod":-5,"categoria":"Perfumes"}' \
    "el stock no puede ser negativo"

# Categoria vacia
probar_producto_invalido "categoria vacia" \
    '{"nombre_prod":"Invictus","precio_prod":100,"stock_prod":10,"categoria":""}' \
    "la categoria no puede estar vacia"

# Tipo de precio incorrecto
probar_producto_invalido "precio como texto" \
    '{"nombre_prod":"Invictus","precio_prod":"hola","stock_prod":10,"categoria":"Perfumes"}' \
    "JSON invalido"

echo ""
echo "=== 1. POST /carritos (Crear carrito) ==="
CREATE_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/carritos" \
  -H "Content-Type: application/json" \
  -d '{"total": 0}')

STATUS=$(echo "$CREATE_RESP" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$CREATE_RESP" | grep -v "HTTP_STATUS:")

if [ "$STATUS" -ne 201 ]; then
  echo "Error en POST /carritos. Status esperado: 201, obtenido: $STATUS"
  echo "Respuesta: $BODY"
  exit 1
fi

ID=$(echo "$BODY" | grep -o '"id_carrito":[0-9]*' | head -n 1 | cut -d: -f2)
echo "Carrito creado exitosamente con ID: $ID"
echo "$BODY"
echo ""

echo "=== 2. GET /carritos (Listar carritos) ==="
LIST_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/carritos")
if [ "$LIST_STATUS" -ne 200 ]; then
  echo "Error en GET /carritos. Status esperado: 200, obtenido: $LIST_STATUS"
  exit 1
fi
curl -s "$BASE_URL/carritos"
echo ""
echo ""

echo "=== 3. GET /carritos/$ID (Obtener carrito por ID) ==="
GET_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/carritos/$ID")
if [ "$GET_STATUS" -ne 200 ]; then
  echo "Error en GET /carritos/$ID. Status esperado: 200, obtenido: $GET_STATUS"
  exit 1
fi
curl -s "$BASE_URL/carritos/$ID"
echo ""
echo ""

echo "=== 4. PUT /carritos/$ID (Actualizar carrito) ==="
UPDATE_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X PUT "$BASE_URL/carritos/$ID" \
  -H "Content-Type: application/json" \
  -d '{"total": 18500.00}')

UPDATE_STATUS=$(echo "$UPDATE_RESP" | grep "HTTP_STATUS:" | cut -d: -f2)
UPDATE_BODY=$(echo "$UPDATE_RESP" | grep -v "HTTP_STATUS:")

if [ "$UPDATE_STATUS" -ne 200 ]; then
  echo "Error en PUT /carritos/$ID. Status esperado: 200, obtenido: $UPDATE_STATUS"
  echo "Respuesta: $UPDATE_BODY"
  exit 1
fi
echo "Carrito actualizado exitosamente:"
echo "$UPDATE_BODY"
echo ""

echo "=== 5. DELETE /carritos/$ID (Eliminar carrito) ==="
DEL_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/carritos/$ID")
if [ "$DEL_STATUS" -ne 204 ]; then
  echo "Error en DELETE /carritos/$ID. Status esperado: 204, obtenido: $DEL_STATUS"
  exit 1
fi
echo "Carrito eliminado exitosamente (HTTP 204)."
echo ""

echo "=== 6. GET /carritos/$ID (Verificar eliminacion -> 404) ==="
NOT_FOUND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/carritos/$ID")
if [ "$NOT_FOUND_STATUS" -ne 404 ]; then
  echo "Error en comprobacion 404. Status esperado: 404, obtenido: $NOT_FOUND_STATUS"
  exit 1
fi
echo "Verificacion correcta: el carrito borrado retorno HTTP 404."
echo ""
echo "=== Preparar producto y carrito para los items ==="
PRODUCTO_RESP=$(curl -fsS -X POST "$BASE_URL/productos" \
  -H "Content-Type: application/json" \
  -d '{"nombre_prod":"Perfume para carrito","precio_prod":100,"stock_prod":10,"categoria":"Fragancias"}')
PRODUCTO_ID=$(echo "$PRODUCTO_RESP" | grep -o '"id_producto":[0-9]*' | head -n 1 | cut -d: -f2)
CARRITO_RESP=$(curl -fsS -X POST "$BASE_URL/carritos" \
  -H "Content-Type: application/json" -d '{"total":0}')
CARRITO_ID=$(echo "$CARRITO_RESP" | grep -o '"id_carrito":[0-9]*' | head -n 1 | cut -d: -f2)

echo "=== 1. POST /productos_carrito (Crear item) ==="
CREATE_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/productos_carrito" \
  -H "Content-Type: application/json" \
  -d "{\"id_producto\":$PRODUCTO_ID,\"id_carrito\":$CARRITO_ID,\"cantidad\":2}")

STATUS=$(echo "$CREATE_RESP" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$CREATE_RESP" | grep -v "HTTP_STATUS:")

if [ "$STATUS" -ne 201 ]; then
  echo "Error en POST /productos_carrito. Status esperado: 201, obtenido: $STATUS"
  echo "Respuesta: $BODY"
  exit 1
fi

ID=$(echo "$BODY" | grep -o '"id_productos_carrito":[0-9]*' | head -n 1 | cut -d: -f2)
echo "Item creado exitosamente con ID: $ID"
echo "$BODY"
echo ""

echo "=== 2. GET /productos_carrito (Listar items) ==="
LIST_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/productos_carrito")
if [ "$LIST_STATUS" -ne 200 ]; then
  echo "Error en GET /productos_carrito. Status esperado: 200, obtenido: $LIST_STATUS"
  exit 1
fi
curl -s "$BASE_URL/productos_carrito"
echo ""
echo ""

echo "=== 3. GET /productos_carrito/$ID (Obtener item por ID) ==="
GET_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/productos_carrito/$ID")
if [ "$GET_STATUS" -ne 200 ]; then
  echo "Error en GET /productos_carrito/$ID. Status esperado: 200, obtenido: $GET_STATUS"
  exit 1
fi
curl -s "$BASE_URL/productos_carrito/$ID"
echo ""
echo ""

echo "=== 4. PUT /productos_carrito/$ID (Actualizar item) ==="
UPDATE_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X PUT "$BASE_URL/productos_carrito/$ID" \
  -H "Content-Type: application/json" \
  -d '{"cantidad": 3}')

UPDATE_STATUS=$(echo "$UPDATE_RESP" | grep "HTTP_STATUS:" | cut -d: -f2)
UPDATE_BODY=$(echo "$UPDATE_RESP" | grep -v "HTTP_STATUS:")

if [ "$UPDATE_STATUS" -ne 200 ]; then
  echo "Error en PUT /productos_carrito/$ID. Status esperado: 200, obtenido: $UPDATE_STATUS"
  echo "Respuesta: $UPDATE_BODY"
  exit 1
fi
echo "Item actualizado exitosamente:"
echo "$UPDATE_BODY"
echo ""

echo "=== 5. DELETE /productos_carrito/$ID (Eliminar item) ==="
DEL_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/productos_carrito/$ID")
if [ "$DEL_STATUS" -ne 204 ]; then
  echo "Error en DELETE /productos_carrito/$ID. Status esperado: 204, obtenido: $DEL_STATUS"
  exit 1
fi
echo "Item eliminado exitosamente (HTTP 204)."
echo ""

echo "=== 6. GET /productos_carrito/$ID (Verificar eliminacion -> 404) ==="
NOT_FOUND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/productos_carrito/$ID")
if [ "$NOT_FOUND_STATUS" -ne 404 ]; then
  echo "Error en comprobacion 404. Status esperado: 404, obtenido: $NOT_FOUND_STATUS"
  exit 1
fi
echo "Verificacion correcta: el item borrado retorno HTTP 404."
echo ""
curl -fsS -X DELETE "$BASE_URL/carritos/$CARRITO_ID"
curl -fsS -X DELETE "$BASE_URL/productos/$PRODUCTO_ID"

USUARIO_EMAIL="crud-$(date +%s%N)@example.com"
echo "=== 1. POST /usuarios (Crear usuario) ==="
CREATE_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "$BASE_URL/usuarios" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$USUARIO_EMAIL\",\"contrasena\":\"secreto123\"}")

STATUS=$(echo "$CREATE_RESP" | grep "HTTP_STATUS:" | cut -d: -f2)
BODY=$(echo "$CREATE_RESP" | grep -v "HTTP_STATUS:")

if [ "$STATUS" -ne 201 ]; then
  echo "Error en POST /usuarios. Status esperado: 201, obtenido: $STATUS"
  echo "Respuesta: $BODY"
  exit 1
fi

ID=$(echo "$BODY" | grep -o '"id_usuario":[0-9]*' | head -n 1 | cut -d: -f2)
echo "Usuario creado exitosamente con ID: $ID"
echo "$BODY"
echo ""

echo "=== 2. GET /usuarios (Listar usuarios) ==="
LIST_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/usuarios")
if [ "$LIST_STATUS" -ne 200 ]; then
  echo "Error en GET /usuarios. Status esperado: 200, obtenido: $LIST_STATUS"
  exit 1
fi
curl -s "$BASE_URL/usuarios"
echo ""
echo ""

echo "=== 3. GET /usuarios/$ID (Obtener usuario por ID) ==="
GET_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/usuarios/$ID")
if [ "$GET_STATUS" -ne 200 ]; then
  echo "Error en GET /usuarios/$ID. Status esperado: 200, obtenido: $GET_STATUS"
  exit 1
fi
curl -s "$BASE_URL/usuarios/$ID"
echo ""
echo ""

echo "=== 4. PUT /usuarios/$ID (Actualizar usuario) ==="
UPDATE_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X PUT "$BASE_URL/usuarios/$ID" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"actualizado-$USUARIO_EMAIL\",\"contrasena\":\"nuevaclave123\"}")

UPDATE_STATUS=$(echo "$UPDATE_RESP" | grep "HTTP_STATUS:" | cut -d: -f2)
UPDATE_BODY=$(echo "$UPDATE_RESP" | grep -v "HTTP_STATUS:")

if [ "$UPDATE_STATUS" -ne 200 ]; then
  echo "Error en PUT /usuarios/$ID. Status esperado: 200, obtenido: $UPDATE_STATUS"
  echo "Respuesta: $UPDATE_BODY"
  exit 1
fi
echo "Usuario actualizado exitosamente:"
echo "$UPDATE_BODY"
echo ""

echo "=== 5. DELETE /usuarios/$ID (Eliminar usuario) ==="
DEL_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -X DELETE "$BASE_URL/usuarios/$ID")
if [ "$DEL_STATUS" -ne 204 ]; then
  echo "Error en DELETE /usuarios/$ID. Status esperado: 204, obtenido: $DEL_STATUS"
  exit 1
fi
echo "Usuario eliminado exitosamente (HTTP 204)."
echo ""

echo "=== 6. GET /usuarios/$ID (Verificar eliminacion -> 404) ==="
NOT_FOUND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/usuarios/$ID")
if [ "$NOT_FOUND_STATUS" -ne 404 ]; then
  echo "Error en comprobacion 404. Status esperado: 404, obtenido: $NOT_FOUND_STATUS"
  exit 1
fi
echo "Verificacion correcta: el usuario borrado retorno HTTP 404."
echo ""
echo "=== TODAS LAS PRUEBAS DE LA API FINALIZARON CON EXITO ==="
