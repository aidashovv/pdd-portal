-include .env
export

export PROJECT_ROOT=$(shell pwd)

check-env:
	@test -f .env || (echo "required file is missing: .env. Create it with: cp .env.example .env" && exit 1)

env-up: check-env
	@docker compose up -d postgres redis

env-down: check-env
	@docker compose stop postgres redis

env-cleanup: check-env
	@docker compose down -v
	@rm -rf .out/pgdata

migrate-create: check-env
	@if [ -z "$(seq)" ]; then \
		echo "required parameter is missing: make migrate-create seq=your_variant"; \
		exit 1; \
	fi

	@docker compose run --rm migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-up: check-env
	@make migrate-action action=up

migrate-down: check-env
	@make migrate-action action=down

migrate-action: check-env
	@if [ -z "$(action)" ]; then \
		echo "required parameter is missing: make migrate-action action=your_variant"; \
		exit 1; \
	fi

	@docker compose run --rm migrate \
		-path /migrations \
		-database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSL_MODE}" \
		"$(action)"

up: check-env
	@docker compose up -d --build

down: check-env
	@docker compose down

run: check-env
	@go run ./cmd/pdd-service

fmt:
	@gofmt -w $$(find cmd internal -name '*.go' -type f)

test:
	@go test ./...

lint:
	@go vet ./...
