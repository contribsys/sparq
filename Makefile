.DEFAULT_GOAL := test

 generate:
	go generate ./...

test: generate
	go test ./...

int:
	go run test/main.go

db:
	sqlite3 sparq.db

up:
	go run ./cmd/sparq migrate

redo:
	go run ./cmd/sparq migrate redo

run:
	go run ./cmd/sparq run -l debug

build:
	go build -o sparq ./cmd/sparq

clean:
	rm redis.log faktory.rdb

.PHONY: build run test generate db up redo clean
