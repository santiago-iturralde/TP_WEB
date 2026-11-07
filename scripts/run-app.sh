#!/bin/sh
# Ciclo de entrega: base limpia -> generación/compilación -> esquema -> app -> limpieza.
set -eu
cd "$(dirname "$0")/.."
APP_PID=""
case "${1:-}" in ""|--check) ;; *) echo "Uso: sh scripts/run-app.sh [--check]" >&2; exit 2 ;; esac
for tool in go make docker curl; do
    command -v "$tool" >/dev/null || { echo "Falta instalar $tool" >&2; exit 1; }
done
docker compose version >/dev/null
compose() { docker compose -p tpweb-tp5 -f docker-compose.delivery.yml "$@"; }
cleanup() {
    status=$?
    trap - EXIT HUP INT TERM
    if [ -n "$APP_PID" ]; then
        kill "$APP_PID" 2>/dev/null || true
        wait "$APP_PID" 2>/dev/null || true
    fi
    echo "Eliminando contenedores y volumen de la entrega TP5..."
    compose down -v --remove-orphans || { [ "$status" -ne 0 ] || status=1; }
    exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP

compose down -v --remove-orphans
make build
compose up -d --wait --wait-timeout 60
compose exec -T database psql -v ON_ERROR_STOP=1 -U postgres -d perfumes_tp5 < db/schema/schema.sql

export DATABASE_URL="postgres://postgres:tp5_local@127.0.0.1:${DELIVERY_DB_PORT:-5434}/perfumes_tp5?sslmode=disable"
export HTTP_PORT="${HTTP_PORT:-8080}"
export BASE_URL="http://127.0.0.1:$HTTP_PORT"
./tmp/TP_WEB &
APP_PID=$!
i=0
while :; do
    kill -0 "$APP_PID" 2>/dev/null || { wait "$APP_PID"; exit 1; }
    if curl -fsS "$BASE_URL/" >/dev/null 2>&1; then
        # Comprobar que no respondió un servidor anterior mientras el nuestro fallaba.
        sleep 1
        kill -0 "$APP_PID" 2>/dev/null || { wait "$APP_PID"; exit 1; }
        break
    fi
    i=$((i + 1))
    [ "$i" -lt 30 ] || { echo "El servidor no respondió a tiempo" >&2; exit 1; }
    sleep 1
done
if [ "${1:-}" = --check ]; then
    sh scripts/check-ssr.sh
else
    echo "TP5 listo en $BASE_URL/ — Ctrl+C detiene la aplicación y elimina los datos de esta ejecución."
    wait "$APP_PID"
fi
