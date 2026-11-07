#!/bin/sh
# Prueba HTTP sin navegador ni JavaScript, sobre una base de entrega vacía.
set -eu
BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
CHECK_DIR=$(mktemp -d)
trap 'rm -rf "$CHECK_DIR"' EXIT
curl -fsS "$BASE_URL/" > "$CHECK_DIR/page"
grep -q 'Todavía no hay perfumes' "$CHECK_DIR/page"
grep -q 'action="/productos"' "$CHECK_DIR/page"
if grep -qi '<script' "$CHECK_DIR/page"; then echo "La página contiene JavaScript" >&2; exit 1; fi
curl -fsS "$BASE_URL/static/estilos.css" >/dev/null
code=$(curl -sS -o "$CHECK_DIR/post" -D "$CHECK_DIR/headers" -w '%{http_code}' \
    --data-urlencode 'nombre_prod=Perfume <SSR>' --data-urlencode 'descripcion_prod=Prueba sin JS' \
    --data-urlencode 'precio_prod=150.50' --data-urlencode 'stock_prod=3' --data-urlencode 'categoria=Prueba' "$BASE_URL/productos")
[ "$code" = 303 ]
grep -qi '^Location: /' "$CHECK_DIR/headers"
curl -fsS "$BASE_URL/" > "$CHECK_DIR/page"
grep -q 'Perfume &lt;SSR&gt;' "$CHECK_DIR/page"
grep -q '1 producto' "$CHECK_DIR/page"
code=$(curl -sS -o "$CHECK_DIR/invalid" -w '%{http_code}' --data-urlencode 'nombre_prod=No guardar' \
    --data-urlencode 'precio_prod=NaN' --data-urlencode 'stock_prod=1' --data-urlencode 'categoria=Prueba' "$BASE_URL/productos")
[ "$code" = 400 ]
grep -q 'value="No guardar"' "$CHECK_DIR/invalid"
curl -fsS "$BASE_URL/" > "$CHECK_DIR/page"
grep -q '1 producto' "$CHECK_DIR/page"
for path in /static/app.js /static/index.html /usuario; do
    code=$(curl -sS -o /dev/null -w '%{http_code}' "$BASE_URL$path")
    [ "$code" = 404 ]
done
echo "SSR OK: catálogo vacío, CSS, alta, 303, HTML escapado y validación sin JavaScript."
