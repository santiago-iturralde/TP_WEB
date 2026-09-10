APP_NAME := TP_WEB
DB_URL := postgres://postgres:postgres@localhost:5432/tp_web?sslmode=disable


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
	@go test ./

clean:
	@rm -rf tmp
	