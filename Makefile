.PHONY: build run test test-unit test-integration test-benchmark docker-up docker-down migrate-up migrate-down goose-up goose-down goose-status db-create db-drop db-reset

build:
	go build -o bin/api ./api
	go build -o bin/image-processor ./image-processor

run:
	go run ./api

# Test targets
test:
	go test ./... -v

test-unit:
	go test ./api/tests/unit/... -v

test-integration:
	go test ./api/tests/integration/... -v

test-benchmark:
	go test ./api/tests/benchmark/... -bench=. -v

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

# Goose migrations
goose-up:
	goose -dir ./migrations/sql postgres "user=postgres password=postgres dbname=product_management host=localhost port=5431 sslmode=disable" up

goose-down:
	goose -dir ./migrations/sql postgres "user=postgres password=postgres dbname=product_management host=localhost port=5431 sslmode=disable" down

goose-status:
	goose -dir ./migrations/sql postgres "user=postgres password=postgres dbname=product_management host=localhost port=5431 sslmode=disable" status

# Database management
db-create:
	docker exec -it product-management-system-postgres-1 psql -U postgres -c "CREATE DATABASE product_management;"

db-drop:
	docker exec -it product-management-system-postgres-1 psql -U postgres -c "DROP DATABASE IF EXISTS product_management;"

db-reset: db-drop db-create goose-up

.DEFAULT_GOAL := run
