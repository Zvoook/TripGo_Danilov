-include .env
export

.PHONY: run build test generate migrate migrate-down env-check

run:
	go run ./cmd/trip-service

build:
	go build -o bin/trip-service ./cmd/trip-service

test:
	go test -race ./...

generate:
	@mkdir -p internal/generated
	go tool oapi-codegen -generate types,chi-server -package api \
	  -include-operation-ids createTrip,getTrip,finishTrip,health,ready \
	  -o internal/generated/api.gen.go \
	  contracts/openapi/trip-service.openapi.yaml

migrate:
	go tool goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	go tool goose -dir migrations postgres "$(DATABASE_URL)" down
env-check:
	bash scripts/check-environment.sh
