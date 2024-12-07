.PHONY: build run test docker-up docker-down

build:
	go build -o bin/api ./api
	go build -o bin/image-processor ./image-processor

run:
	go run ./api

test:
	go test ./...

docker-up:
	docker-compose up --build -d

docker-down:
	docker-compose down

migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5431/product_management?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5431/product_management?sslmode=disable" down