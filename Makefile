.PHONY: run migrate test test-integration check build

run:
	cd backend && go run ./cmd/api

migrate:
	cd backend && go run ./cmd/migrate up

test:
	cd backend && go test ./...

test-integration:
	@test -n "$(TEST_DATABASE_URL)" || (echo "TEST_DATABASE_URL is required" && exit 1)
	cd backend && go test -race -count=1 ./...

check:
	cd backend && test -z "$$(gofmt -l .)" && go vet ./...

build:
	cd backend && go build -o bin/api ./cmd/api && go build -o bin/migrate ./cmd/migrate
