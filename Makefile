.DEFAULT_GOAL := test

test:
	go test ./...

int:
	go run test/main.go

db:
	sqlite3 sparq.db

run:
	go run cmd/sparq/main.go

build:
	go build -o sparq cmd/sparq/main.go