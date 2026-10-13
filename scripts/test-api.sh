#!/bin/sh
set -eu
cd "$(dirname "$0")/.."

APP_PID=""

compose() {
    docker compose -p tpweb-test -f docker-compose.test.yml "$@"
}

cleanup() {
    status=$?
    trap - EXIT HUP INT TERM
    if [ -n "$APP_PID" ] && kill -0 "$APP_PID" 2>/dev/null; then
        echo "Deteniendo la aplicacion web (PID: $APP_PID)..."
        kill "$APP_PID" 2>/dev/null || true
    fi
    echo "Limpiando la base de tests API..."
    compose down -v || { [ "$status" -ne 0 ] || status=1; }
    exit "$status"
}

trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP

echo "Preparando entorno de pruebas para la API..."
compose down -v
compose up -d --wait --wait-timeout 60
compose exec -T database psql -v ON_ERROR_STOP=1 -U postgres -d perfumes_test < db/schema/schema.sql

export DB_HOST=localhost DB_PORT="${TEST_DB_PORT:-5433}"
export DB_USER=postgres DB_PASSWORD=test_password DB_NAME=perfumes_test
export DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"

echo "Iniciando servidor de la aplicacion en segundo plano..."
./tmp/TP_WEB &
APP_PID=$!

echo "Esperando a que el servidor este listo..."
i=0
while [ $i -lt 30 ]; do
    if curl -s http://localhost:8080/usuario >/dev/null 2>&1; then
        echo "Servidor en ejecucion."
        break
    fi
    sleep 1
    i=$((i + 1))
done

if [ $i -eq 30 ]; then
    echo "Error: El servidor no respondio a tiempo."
    exit 1
fi

echo "Ejecutando pruebas de peticiones HTTP a la API REST..."
sh requests.sh