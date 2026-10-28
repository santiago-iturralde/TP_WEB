#!/bin/sh
set -eu

BASE_URL="${BASE_URL:-http://localhost:8080}"

verificar_json() {
  CUERPO_JSON=$1
  FRAGMENTO_JSON=$2
  DESCRIPCION_JSON=$3

  case "$CUERPO_JSON" in
    *"$FRAGMENTO_JSON"*) ;;
    *)
      echo "Error: $DESCRIPCION_JSON"
      echo "Se esperaba encontrar: $FRAGMENTO_JSON"
      echo "Respuesta: $CUERPO_JSON"
      exit 1
      ;;
  esac
}

verificar_json_sin_contrasena() {
  CUERPO_JSON=$1
  case "$CUERPO_JSON" in
    *'"contrasena"'*)
      echo "Error: la respuesta JSON expone la contrasena"
      echo "Respuesta: $CUERPO_JSON"
      exit 1
      ;;
  esac
}

probar_dato_invalido() {
  DESCRIPCION=$1
  RUTA=$2
  METODO=$3
  DATOS_JSON=$4
  ERROR_ESPERADO=$5

  echo "=== Probando dato invalido: $DESCRIPCION ==="
  RESP=$(curl -sS -w "\nHTTP_STATUS:%{http_code}" -X "$METODO" "$BASE_URL$RUTA" \
    -H "Content-Type: application/json" -d "$DATOS_JSON")
  STATUS=$(printf '%s\n' "$RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
  BODY=$(printf '%s\n' "$RESP" | sed '/^HTTP_STATUS:/d')

  if [ "$STATUS" != "400" ]; then
    echo "Error: esperabamos HTTP 400, recibimos $STATUS"
    echo "Respuesta: $BODY"
    exit 1
  fi
  case "$BODY" in
    *"$ERROR_ESPERADO"*) echo "OK: HTTP 400 y mensaje de error esperado" ;;
    *)
      echo "Error: el mensaje no coincide"
      echo "Esperado: $ERROR_ESPERADO"
      echo "Recibido: $BODY"
      exit 1
      ;;
  esac
}

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
verificar_json "$BODY" '"nombre_prod":"Chanel No. 5"' "el nombre del producto creado no coincide"
verificar_json "$BODY" '"precio_prod":15000.5' "el precio del producto creado no coincide"
verificar_json "$BODY" '"stock_prod":20' "el stock del producto creado no coincide"
verificar_json "$BODY" '"categoria":"Fragancias"' "la categoria del producto creado no coincide"

ID=$(echo "$BODY" | grep -o '"id_producto":[0-9]*' | head -n 1 | cut -d: -f2)
echo "Producto creado exitosamente con ID: $ID"
echo "$BODY"
echo ""

echo "=== 2. GET /productos (Listar productos) ==="
LIST_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/productos")
LIST_STATUS=$(printf '%s\n' "$LIST_RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
LIST_BODY=$(printf '%s\n' "$LIST_RESP" | sed '/^HTTP_STATUS:/d')
if [ "$LIST_STATUS" -ne 200 ]; then
  echo "Error en GET /productos. Status esperado: 200, obtenido: $LIST_STATUS"
  exit 1
fi
verificar_json "$LIST_BODY" "\"id_producto\":$ID" "el listado no incluye el producto creado"
verificar_json "$LIST_BODY" '"nombre_prod":"Chanel No. 5"' "el listado no devuelve los datos del producto"
echo "$LIST_BODY"
echo ""
echo ""

echo "=== 3. GET /productos/$ID (Obtener producto por ID) ==="
GET_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/productos/$ID")
GET_STATUS=$(printf '%s\n' "$GET_RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
GET_BODY=$(printf '%s\n' "$GET_RESP" | sed '/^HTTP_STATUS:/d')
if [ "$GET_STATUS" -ne 200 ]; then
  echo "Error en GET /productos/$ID. Status esperado: 200, obtenido: $GET_STATUS"
  exit 1
fi
verificar_json "$GET_BODY" "\"id_producto\":$ID" "el ID del producto obtenido no coincide"
verificar_json "$GET_BODY" '"nombre_prod":"Chanel No. 5"' "el nombre del producto obtenido no coincide"
verificar_json "$GET_BODY" '"descripcion_prod":{"String":"Perfume clasico elegante","Valid":true}' "la descripcion del producto obtenido no coincide"
echo "$GET_BODY"
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
verificar_json "$UPDATE_BODY" '"nombre_prod":"Chanel No. 5 Redesign"' "el nombre actualizado del producto no coincide"
verificar_json "$UPDATE_BODY" '"descripcion_prod":{"String":"Edicion especial 2026","Valid":true}' "la descripcion actualizada del producto no coincide"
verificar_json "$UPDATE_BODY" '"precio_prod":18500' "el precio actualizado del producto no coincide"
verificar_json "$UPDATE_BODY" '"stock_prod":15' "el stock actualizado del producto no coincide"
verificar_json "$UPDATE_BODY" '"categoria":"Fragancias de Lujo"' "la categoria actualizada del producto no coincide"
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

probar_dato_invalido "JSON invalido en producto" "/productos" "POST" "{" "JSON invalido"
probar_dato_invalido "precio negativo en producto" "/productos" "POST" \
  '{"nombre_prod":"Prueba","precio_prod":-1,"stock_prod":1,"categoria":"Perfumes"}' \
  'el precio debe ser un numero valido mayor a cero'

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

verificar_json "$BODY" '"total":0' "el total del carrito creado no coincide"

ID=$(echo "$BODY" | grep -o '"id_carrito":[0-9]*' | head -n 1 | cut -d: -f2)
echo "Carrito creado exitosamente con ID: $ID"
echo "$BODY"
echo ""

echo "=== 2. GET /carritos (Listar carritos) ==="
LIST_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/carritos")
LIST_STATUS=$(printf '%s\n' "$LIST_RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
LIST_BODY=$(printf '%s\n' "$LIST_RESP" | sed '/^HTTP_STATUS:/d')
if [ "$LIST_STATUS" -ne 200 ]; then
  echo "Error en GET /carritos. Status esperado: 200, obtenido: $LIST_STATUS"
  exit 1
fi
verificar_json "$LIST_BODY" "\"id_carrito\":$ID" "el listado no incluye el carrito creado"
verificar_json "$LIST_BODY" '"total":0' "el listado no devuelve el total esperado"
echo "$LIST_BODY"
echo ""
echo ""

echo "=== 3. GET /carritos/$ID (Obtener carrito por ID) ==="
GET_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/carritos/$ID")
GET_STATUS=$(printf '%s\n' "$GET_RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
GET_BODY=$(printf '%s\n' "$GET_RESP" | sed '/^HTTP_STATUS:/d')
if [ "$GET_STATUS" -ne 200 ]; then
  echo "Error en GET /carritos/$ID. Status esperado: 200, obtenido: $GET_STATUS"
  exit 1
fi
verificar_json "$GET_BODY" "\"id_carrito\":$ID" "el ID del carrito obtenido no coincide"
verificar_json "$GET_BODY" '"total":0' "el total del carrito obtenido no coincide"
echo "$GET_BODY"
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
verificar_json "$UPDATE_BODY" "\"id_carrito\":$ID" "el ID del carrito actualizado no coincide"
verificar_json "$UPDATE_BODY" '"total":18500' "el total actualizado del carrito no coincide"
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
probar_dato_invalido "total negativo en carrito" "/carritos" "POST" '{"total":-1}' 'el total debe ser un numero valido entre 0 y 99999999.99'
probar_dato_invalido "total sobre el limite en carrito" "/carritos" "POST" '{"total":100000000}' 'el total debe ser un numero valido entre 0 y 99999999.99'
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

verificar_json "$BODY" "\"id_producto\":$PRODUCTO_ID" "el item no conserva el ID del producto enviado"
verificar_json "$BODY" "\"id_carrito\":$CARRITO_ID" "el item no conserva el ID del carrito enviado"
verificar_json "$BODY" '"cantidad":2' "la cantidad del item creado no coincide"

ID=$(echo "$BODY" | grep -o '"id_productos_carrito":[0-9]*' | head -n 1 | cut -d: -f2)
echo "Item creado exitosamente con ID: $ID"
echo "$BODY"
echo ""

echo "=== 2. GET /productos_carrito (Listar items) ==="
LIST_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/productos_carrito")
LIST_STATUS=$(printf '%s\n' "$LIST_RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
LIST_BODY=$(printf '%s\n' "$LIST_RESP" | sed '/^HTTP_STATUS:/d')
if [ "$LIST_STATUS" -ne 200 ]; then
  echo "Error en GET /productos_carrito. Status esperado: 200, obtenido: $LIST_STATUS"
  exit 1
fi
verificar_json "$LIST_BODY" "\"id_productos_carrito\":$ID" "el listado no incluye el item creado"
verificar_json "$LIST_BODY" "\"id_producto\":$PRODUCTO_ID" "el listado no devuelve el producto asociado"
verificar_json "$LIST_BODY" '"cantidad":2' "el listado no devuelve la cantidad esperada"
echo "$LIST_BODY"
echo ""
echo ""

echo "=== 3. GET /productos_carrito/$ID (Obtener item por ID) ==="
GET_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/productos_carrito/$ID")
GET_STATUS=$(printf '%s\n' "$GET_RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
GET_BODY=$(printf '%s\n' "$GET_RESP" | sed '/^HTTP_STATUS:/d')
if [ "$GET_STATUS" -ne 200 ]; then
  echo "Error en GET /productos_carrito/$ID. Status esperado: 200, obtenido: $GET_STATUS"
  exit 1
fi
verificar_json "$GET_BODY" "\"id_productos_carrito\":$ID" "el ID del item obtenido no coincide"
verificar_json "$GET_BODY" "\"id_carrito\":$CARRITO_ID" "el carrito del item obtenido no coincide"
verificar_json "$GET_BODY" '"cantidad":2' "la cantidad del item obtenido no coincide"
echo "$GET_BODY"
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
verificar_json "$UPDATE_BODY" "\"id_productos_carrito\":$ID" "el ID del item actualizado no coincide"
verificar_json "$UPDATE_BODY" '"cantidad":3' "la cantidad actualizada del item no coincide"
echo "Item actualizado exitosamente:"
echo "$UPDATE_BODY"
echo ""

probar_dato_invalido "cantidad cero al actualizar item" "/productos_carrito/$ID" "PUT" \
  '{"cantidad":0}' 'la cantidad debe ser mayor a cero'

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
probar_dato_invalido "cantidad cero al crear item" "/productos_carrito" "POST" \
  "{\"id_producto\":$PRODUCTO_ID,\"id_carrito\":$CARRITO_ID,\"cantidad\":0}" \
  'la cantidad debe ser mayor a cero'
probar_dato_invalido "producto inexistente al crear item" "/productos_carrito" "POST" \
  "{\"id_producto\":2147483647,\"id_carrito\":$CARRITO_ID,\"cantidad\":1}" \
  'el producto o el carrito no existe'
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

verificar_json "$BODY" "\"email\":\"$USUARIO_EMAIL\"" "el email del usuario creado no coincide"
verificar_json_sin_contrasena "$BODY"

ID=$(echo "$BODY" | grep -o '"id_usuario":[0-9]*' | head -n 1 | cut -d: -f2)
echo "Usuario creado exitosamente con ID: $ID"
echo "$BODY"
echo ""

echo "=== 2. GET /usuarios (Listar usuarios) ==="
LIST_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/usuarios")
LIST_STATUS=$(printf '%s\n' "$LIST_RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
LIST_BODY=$(printf '%s\n' "$LIST_RESP" | sed '/^HTTP_STATUS:/d')
if [ "$LIST_STATUS" -ne 200 ]; then
  echo "Error en GET /usuarios. Status esperado: 200, obtenido: $LIST_STATUS"
  exit 1
fi
verificar_json "$LIST_BODY" "\"id_usuario\":$ID" "el listado no incluye el usuario creado"
verificar_json "$LIST_BODY" "\"email\":\"$USUARIO_EMAIL\"" "el listado no devuelve el email esperado"
verificar_json_sin_contrasena "$LIST_BODY"
echo "$LIST_BODY"
echo ""
echo ""

echo "=== 3. GET /usuarios/$ID (Obtener usuario por ID) ==="
GET_RESP=$(curl -s -w "\nHTTP_STATUS:%{http_code}" "$BASE_URL/usuarios/$ID")
GET_STATUS=$(printf '%s\n' "$GET_RESP" | grep '^HTTP_STATUS:' | cut -d: -f2)
GET_BODY=$(printf '%s\n' "$GET_RESP" | sed '/^HTTP_STATUS:/d')
if [ "$GET_STATUS" -ne 200 ]; then
  echo "Error en GET /usuarios/$ID. Status esperado: 200, obtenido: $GET_STATUS"
  exit 1
fi
verificar_json "$GET_BODY" "\"id_usuario\":$ID" "el ID del usuario obtenido no coincide"
verificar_json "$GET_BODY" "\"email\":\"$USUARIO_EMAIL\"" "el email del usuario obtenido no coincide"
verificar_json_sin_contrasena "$GET_BODY"
echo "$GET_BODY"
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
verificar_json "$UPDATE_BODY" "\"id_usuario\":$ID" "el ID del usuario actualizado no coincide"
verificar_json "$UPDATE_BODY" "\"email\":\"actualizado-$USUARIO_EMAIL\"" "el email actualizado no coincide"
verificar_json_sin_contrasena "$UPDATE_BODY"
echo "Usuario actualizado exitosamente:"
echo "$UPDATE_BODY"
echo ""

probar_dato_invalido "email invalido al actualizar usuario" "/usuarios/$ID" "PUT" \
  '{"email":"email-invalido","contrasena":"secreto123"}' 'el email no es valido'
probar_dato_invalido "contrasena corta al actualizar usuario" "/usuarios/$ID" "PUT" \
  "{\"email\":\"actualizado-$USUARIO_EMAIL\",\"contrasena\":\"corta\"}" \
  'la contrasena debe tener al menos 8 caracteres'

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
probar_dato_invalido "email invalido al crear usuario" "/usuarios" "POST" \
  '{"email":"email-invalido","contrasena":"secreto123"}' 'el email no es valido'
probar_dato_invalido "contrasena corta al crear usuario" "/usuarios" "POST" \
  '{"email":"invalido-corto@example.com","contrasena":"corta"}' 'la contrasena debe tener al menos 8 caracteres'
echo ""
echo "=== TODAS LAS PRUEBAS DE LA API FINALIZARON CON EXITO ==="
