SHELL := /bin/sh

.PHONY: up down psql redis-ping migrate-up migrate-down migrate-version migrate-up-docker migrate-down-docker migrate-version-docker sqlc-generate

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

sqlc-generate:
	sqlc generate

be-run:
	go run cmd/server/main.go

fe-run:
	cd frontend && npm run dev
