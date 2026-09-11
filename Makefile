-include .env

DB_USER ?= postgres
DB_PASSWORD ?= 123456
DB_NAME ?= perfumes_db
DB_PORT ?= 5432
DB_HOST ?= localhost
TEST_DB_PORT ?= 5433
export DB_USER DB_PASSWORD DB_NAME DB_PORT DB_HOST TEST_DB_PORT

APP_NAME := TP_WEB
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable
SQLC_VERSION := v1.31.1


.PHONY: all run generate build test clean db-up db-down db-init


all: build


run: db-up
	@go run .

db-up:
	@docker compose -p tp_web -f docker-compose.yml up -d --wait --wait-timeout 60

db-down:
	@docker compose -p tp_web -f docker-compose.yml down

db-init: db-up
	@docker compose -p tp_web -f docker-compose.yml exec -T database psql -v ON_ERROR_STOP=1 -U $(DB_USER) -d $(DB_NAME) < db/schema/schema.sql

generate:
	@go run github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION) generate

build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

test: build
	@sh scripts/test.sh

clean:
	@rm -rf tmp
	
