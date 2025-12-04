SHELL := /bin/sh

.PHONY: up down psql redis-ping migrate-up migrate-down migrate-version migrate-up-docker migrate-down-docker migrate-version-docker migrate-force-docker sqlc-generate be-run be-run-no-pdf fe-run test-all test-no-pdf cover-html cover-func cover-gate fmt imports vet lint vuln ci

# Local DB URL for tooling on host
DATABASE_URL ?= postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable
MIGRATIONS_DIR ?= db/migrations

# In-docker DB URL (service name on compose network)
DOCKER_DATABASE_URL ?= postgres://botla:botla@botla-postgres:5432/botla_dev?sslmode=disable

up:
	docker compose -f docker-compose.dev.yml up -d

down:
	docker compose -f docker-compose.dev.yml down -v

psql:
	docker exec -it botla-postgres psql -U botla -d botla_dev

redis-ping:
	redis-cli -h 127.0.0.1 -p 6379 ping

# Migrations using local migrate binary
migrate-up:
	migrate -path=$(MIGRATIONS_DIR) -database=$(DATABASE_URL) up

migrate-down:
	migrate -path=$(MIGRATIONS_DIR) -database=$(DATABASE_URL) down

migrate-version:
	-migrate -path=$(MIGRATIONS_DIR) -database=$(DATABASE_URL) version

# Migrations using docker image (works without local install)
migrate-up-docker:
	docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database=$(DOCKER_DATABASE_URL) up

migrate-down-docker:
	docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database=$(DOCKER_DATABASE_URL) down

migrate-version-docker:
	-docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database=$(DOCKER_DATABASE_URL) version

# Force migration version (usage: make migrate-force-docker v=4)
migrate-force-docker:
	docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database=$(DOCKER_DATABASE_URL) force $(v)

sqlc-generate:
	sqlc generate

be-run:
	CGO_ENABLED=1 go run -tags fitz cmd/server/main.go

be-run-no-pdf:
	go run cmd/server/main.go

fe-run:
	cd frontend && npm run dev

# Test and coverage targets
test-all:
	CGO_ENABLED=1 go test -tags fitz -race -covermode=atomic -coverpkg=./... ./... -coverprofile=coverage.out

test-no-pdf:
	go test -race -covermode=atomic -coverpkg=./... ./... -coverprofile=coverage.out

cover-html:
	@[ -f coverage.out ] || (echo "coverage.out not found. Run 'make test-all' first." && exit 1)
	go tool cover -html=coverage.out -o coverage.html

cover-func:
	@[ -f coverage.out ] || (echo "coverage.out not found. Run 'make test-all' first." && exit 1)
	go tool cover -func=coverage.out

cover-gate:
	@[ -f coverage.out ] || (echo "coverage.out not found. Run 'make test-all' first." && exit 1)
	@pct=$$(go tool cover -func=coverage.out | tail -n1 | awk '{print $$3}'); \
	 if [ "$$pct" != "100.0%" ]; then echo "Coverage gate failed: $$pct"; exit 1; else echo "Coverage 100%"; fi

fmt:
	gofmt -s -w .

imports:
	goimports -w .

vet:
	go vet ./...

lint:
	golangci-lint run ./...

vuln:
	govulncheck ./...

ci:
	$(MAKE) vet
	$(MAKE) lint
	$(MAKE) test-no-pdf
