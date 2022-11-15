.DEFAULT_GOAL := test

 generate:
	go generate ./...

test: generate
	go test ./...

int:
	go run test/main.go

db:
	sqlite3 sparq.db

run:
	go run cmd/sparq/main.go -l debug

build:
	go build -o sparq cmd/sparq/main.go

.PHONY: build run test generate