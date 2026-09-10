include .env
export DB_SERVICE

APP_NAME := TP_WEB
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable


.PHONY: all run generate build test clean


all: build


run:	
	@air

generate:
	@sqlc generate

build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

test:
	@echo "Limpiando entorno..."
	docker compose  down -v

	@echo "Levantando PostgreSQL..."
	docker compose up -d

	@echo "Esperando PostgreSQL..."
	@until docker compose exec -T $(DB_SERVICE) \
	psql -U $(DB_USER) -d $(DB_NAME) -c "SELECT 1" > /dev/null 2>&1; do \
		sleep 1; \
	done

	@echo "Cargando schema..."
	@docker compose exec -T $(DB_SERVICE) \
		psql -v ON_ERROR_STOP=1 -U $(DB_USER) -d $(DB_NAME) < db/schema/schema.sql

	@echo "Ejecutando pruebas..."
	go test -v ./...

	@echo "Dejando entorno limpio..."
	docker compose  down -v

clean:
	@rm -rf tmp
	
