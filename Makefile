.DEFAULT_GOAL := test

 generate:
	go generate ./...

test: generate
	go test github.com/contribsys/sparq/finger

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
	rm -f redis.log faktory.rdb sparq.db

.PHONY: build run test generate db up redo clean
