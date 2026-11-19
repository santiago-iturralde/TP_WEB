#!/bin/sh
# Prueba HTTP de la interfaz SSR y de sus fragmentos HTML, sobre una base vacía.
set -eu
BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
CHECK_DIR=$(mktemp -d)
trap 'rm -rf "$CHECK_DIR"' EXIT
curl -fsS "$BASE_URL/" > "$CHECK_DIR/page"
grep -q 'Todavía no hay perfumes' "$CHECK_DIR/page"
grep -q 'action="/productos"' "$CHECK_DIR/page"
grep -q 'htmx.org@2.0.11/dist/htmx.min.js' "$CHECK_DIR/page"
curl -fsS "$BASE_URL/static/estilos.css" >/dev/null
code=$(curl -sS -o "$CHECK_DIR/post" -w '%{http_code}' -H 'HX-Request: true' \
    --data-urlencode 'nombre_prod=Perfume <SSR>' --data-urlencode 'descripcion_prod=Prueba sin JS' \
    --data-urlencode 'precio_prod=150.50' --data-urlencode 'stock_prod=3' --data-urlencode 'categoria=Prueba' "$BASE_URL/productos")
[ "$code" = 200 ]
grep -q 'id="lista-productos"' "$CHECK_DIR/post"
grep -q 'Perfume &lt;SSR&gt;' "$CHECK_DIR/post"
if grep -qi '<!doctype html>' "$CHECK_DIR/post"; then echo "El alta HTMX devolvió la página completa" >&2; exit 1; fi
curl -fsS "$BASE_URL/" > "$CHECK_DIR/page"
grep -q 'Perfume &lt;SSR&gt;' "$CHECK_DIR/page"
grep -q '1 producto' "$CHECK_DIR/page"
grep -q 'hx-delete="/productos/' "$CHECK_DIR/page"
grep -q 'hx-target="closest article"' "$CHECK_DIR/page"
grep -q 'hx-swap="outerHTML"' "$CHECK_DIR/page"
code=$(curl -sS -o "$CHECK_DIR/invalid" -w '%{http_code}' --data-urlencode 'nombre_prod=No guardar' \
    --data-urlencode 'precio_prod=NaN' --data-urlencode 'stock_prod=1' --data-urlencode 'categoria=Prueba' "$BASE_URL/productos")
[ "$code" = 400 ]
grep -q 'value="No guardar"' "$CHECK_DIR/invalid"
curl -fsS "$BASE_URL/" > "$CHECK_DIR/page"
grep -q '1 producto' "$CHECK_DIR/page"
PRODUCTO_ID=$(curl -fsS "$BASE_URL/productos" | grep -o '"id_producto":[0-9]*' | head -n 1 | cut -d: -f2)
code=$(curl -sS -o "$CHECK_DIR/delete" -w '%{http_code}' -X DELETE -H 'HX-Request: true' "$BASE_URL/productos/$PRODUCTO_ID")
[ "$code" = 200 ]
[ ! -s "$CHECK_DIR/delete" ]
curl -fsS "$BASE_URL/" > "$CHECK_DIR/page"
grep -q 'Todavía no hay perfumes' "$CHECK_DIR/page"
grep -q '0 productos' "$CHECK_DIR/page"
for path in /static/app.js /static/index.html /usuario; do
    code=$(curl -sS -o /dev/null -w '%{http_code}' "$BASE_URL$path")
    [ "$code" = 404 ]
done
echo "HTMX/SSR OK: catálogo, alta y borrado con fragmentos, escape y validación."
