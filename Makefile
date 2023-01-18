.DEFAULT_GOAL := test

 generate:
	go generate ./...

test: generate
	go test github.com/contribsys/sparq/wellknown \
		github.com/contribsys/sparq/clientapi \
		github.com/contribsys/sparq/faktory \
		github.com/contribsys/sparq/model \
		github.com/contribsys/sparq/oauth2 \
		github.com/contribsys/sparq/web \
		github.com/contribsys/sparq/web/adminui \
		github.com/contribsys/sparq/web/faktoryui \
		github.com/contribsys/sparq/web/public \
		github.com/contribsys/sparq/util

int:
	go run test/main.go

db:
	sqlite3 sparq.localhost.dev.db

pdb:
	sqlite3 sparq.social.contribsys.com.db

pclean:
	rm sparq.social.contribsys.com.db ~/Library/Application\ Support/tut/accounts.toml

up:
	go run ./cmd/sparq migrate

redo:
	go run ./cmd/sparq migrate redo

prod: generate
	go run ./cmd/sparq run -l debug -h social.contribsys.com

lint:
	# brew install golangci/tap/golangci-lint
	golangci-lint run

run: build
	open http://localhost:9494
	./sparq run -l debug

rund: generate
	go run ./cmd/sparq run -l debug

build: generate
	go build -o sparq ./cmd/sparq

clean:
	rm -f redis.log faktory.rdb sparq.localhost.dev.db

tunnel:
	open http://localhost:9494
	ssh -R 9494:localhost:9494 mike@social.contribsys.com

.PHONY: build run test generate db up redo clean
