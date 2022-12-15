.DEFAULT_GOAL := test

 generate:
	go generate ./...

test: generate
	go test github.com/contribsys/sparq/wellknown \
					github.com/contribsys/sparq/faktoryui \
					github.com/contribsys/sparq/mastapi \
					github.com/contribsys/sparq/public \
					github.com/contribsys/sparq/util

int:
	go run test/main.go

db:
	sqlite3 sparq.localhost.dev.db

pdb:
	sqlite3 sparq.social.contribsys.com.db

up:
	go run ./cmd/sparq migrate

redo:
	go run ./cmd/sparq migrate redo

prod: generate
	go run ./cmd/sparq run -l debug -h social.contribsys.com

lint:
	# brew install golangci/tap/golangci-lint
	golangci-lint run

run: generate
	go run ./cmd/sparq run -l debug

build: generate
	go build -o sparq ./cmd/sparq

clean:
	rm -f redis.log faktory.rdb sparq.localhost.dev.db

tunnel:
	open http://localhost:9494
	ssh -R 9494:localhost:9494 mike@social.contribsys.com

.PHONY: build run test generate db up redo clean
