SHELL := /bin/sh

.PHONY: up down psql redis-ping migrate-up migrate-down migrate-version migrate-up-test migrate-down-test migrate-version-test migrate-up-docker migrate-down-docker migrate-version-docker migrate-up-test-docker migrate-down-test-docker migrate-version-test-docker migrate-force-docker be-run be-run-no-pdf be-run-test fe-run test-all test-no-pdf cover-html cover-func cover-gate fmt imports vet lint vuln ci build run test fmt-go vet-go tidy clean test-cover check-coverage

DATABASE_URL ?= postgres://botla:botla@localhost:5432/botla_dev?sslmode=disable
MIGRATIONS_DIR ?= db/migrations
TEST_DATABASE_URL ?= postgres://botla:botla@localhost:5432/botla_test?sslmode=disable&options=-c%20search_path%3Dtest

DOCKER_DATABASE_URL ?= postgres://botla:botla@botla-postgres:5432/botla_dev?sslmode=disable
DOCKER_TEST_DATABASE_URL ?= $(DOCKER_DATABASE_URL)&options=-c%20search_path%3Dtest

up:
	 docker compose -f docker-compose.dev.yml up -d

down:
	 docker compose -f docker-compose.dev.yml down -v

psql:
	 docker exec -it botla-postgres psql -U botla -d botla_dev

create-test-schema:
	 docker exec -it botla-postgres psql -U botla -d botla_dev -c "CREATE SCHEMA IF NOT EXISTS test"

redis-ping:
	 redis-cli -h 127.0.0.1 -p 6379 ping

migrate-up:
	 migrate -path=$(MIGRATIONS_DIR) -database="$(DATABASE_URL)" up

migrate-down:
	 migrate -path=$(MIGRATIONS_DIR) -database="$(DATABASE_URL)" down

migrate-version:
	 -migrate -path=$(MIGRATIONS_DIR) -database="$(DATABASE_URL)" version

migrate-up-test:
	 migrate -path=$(MIGRATIONS_DIR) -database="$(TEST_DATABASE_URL)" up

migrate-down-test:
	 migrate -path=$(MIGRATIONS_DIR) -database="$(TEST_DATABASE_URL)" down

migrate-version-test:
	-migrate -path=$(MIGRATIONS_DIR) -database="$(TEST_DATABASE_URL)" version

migrate-force-test:
	migrate -path=$(MIGRATIONS_DIR) -database="$(TEST_DATABASE_URL)" force $(v)

migrate-up-docker:
	 docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database=$(DOCKER_DATABASE_URL) up

migrate-down-docker:
	 docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database=$(DOCKER_DATABASE_URL) down

migrate-version-docker:
	 -docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database=$(DOCKER_DATABASE_URL) version

migrate-up-test-docker:
	 docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database="$(DOCKER_TEST_DATABASE_URL)" up

migrate-down-test-docker:
	 docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database="$(DOCKER_TEST_DATABASE_URL)" down

migrate-version-test-docker:
	 -docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database="$(DOCKER_TEST_DATABASE_URL)" version

migrate-force-docker:
	 docker run --rm --network=botla-co_default -v $(PWD)/$(MIGRATIONS_DIR):/migrations migrate/migrate -path=/migrations -database="$(DOCKER_DATABASE_URL)" force $(v)



be-run:
	 CGO_ENABLED=1 go run -tags fitz cmd/server/main.go

be-run-no-pdf:
	 go run cmd/server/main.go

be-run-test:
	 DB_SCHEMA=test go run cmd/server/main.go

fe-run:
	 cd frontend && npm run dev

widget-deploy:
	cd widget && npm install && npm run build && npx wrangler pages deploy dist --project-name botla-widget

test-all:
	 CGO_ENABLED=1 go test -p 1 -tags fitz -covermode=atomic -coverpkg=./... ./... -coverprofile=coverage.out

test-no-pdf:
	 go test -covermode=atomic -coverpkg=./... ./... -coverprofile=coverage.out

cover-html:
	 @[ -f coverage.out ] || (echo "coverage.out not found. Run 'make test-all' first." && exit 1)
	 go tool cover -html=coverage.out -o coverage.html

cover-func:
	 @[ -f coverage.out ] || (echo "coverage.out not found. Run 'make test-all' first." && exit 1)
	 go tool cover -func=coverage.out

cover-gate:
	 @[ -f coverage.out ] || (echo "coverage.out not found. Run 'make test-all' first." && exit 1)
	 @pct=$$(go tool cover -func=coverage.out | tail -n1 | awk '{print $$3}' | tr -d '%'); pct_int=$${pct%.*}; \
	 if [ "$$pct_int" -lt "90" ]; then echo "Coverage gate failed: $$pct%"; exit 1; else echo "Coverage >= 90%: $$pct%"; fi

fmt:
	 gofmt -s -w .

imports:
	 goimports -w .

vet:
	go vet ./...

shadow:
	go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest
	$$(go env GOPATH)/bin/shadow ./...

lint:
	 golangci-lint run ./...

vuln:
	 govulncheck ./...

ci:
	 $(MAKE) vet
	 $(MAKE) lint
	 $(MAKE) test-no-pdf

build:
	 go build ./...

run:
	 go run ./cmd/server/main.go

test:
	 go test ./...

fmt-go:
	 go fmt ./...

vet-go:
	 go vet ./...

tidy:
	 go mod tidy

clean:
	 go clean
	 rm -f coverage.out

test-cover:
	 go test ./... -coverprofile=coverage.out
	 go tool cover -func=coverage.out

check-coverage:
	 bash scripts/check_coverage.sh
