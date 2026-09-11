#!/bin/sh
set -eu
cd "$(dirname "$0")/.."

compose() {
    docker compose -p tpweb-test -f docker-compose.test.yml "$@"
}

cleanup() {
    status=$?
    trap - EXIT HUP INT TERM
    echo "Limpiando la base de tests..."
    compose down -v || { [ "$status" -ne 0 ] || status=1; }
    exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP

compose down -v
compose up -d --wait --wait-timeout 60
compose exec -T database psql -v ON_ERROR_STOP=1 -U postgres -d perfumes_test < db/schema/schema.sql

export DB_HOST=localhost DB_PORT="${TEST_DB_PORT:-5433}"
export DB_USER=postgres DB_PASSWORD=test_password DB_NAME=perfumes_test
export DATABASE_URL="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
export TEST_DATABASE_URL="$DATABASE_URL"
go test -count=1 -v ./...
