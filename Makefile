APP_NAME := TP_WEB
DB_URL := postgres://${DB_USER}:${DB_PASSWORD}@localhost:${DB_PORT}/${DB_NAME}?sslmode=disable


.PHONY: all run generate build test clean test


all: build


run:	
	@air

generate:
	@sqlc generate

build: generate
	@mkdir -p tmp
	@go build -o tmp/$(APP_NAME) .

test:
	@go test ./

clean:
	@rm -rf tmp
	