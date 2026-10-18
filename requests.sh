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
echo "=== TODAS LAS PRUEBAS DE LA API FINALIZARON CON EXITO ==="
