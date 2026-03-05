.PHONY: up down migrate-up migrate-down build run

up:
	docker-compose up -d

down:
	docker-compose down

migrate-up:
	migrate -path internal/db/migrations -database "postgresql://gophermart:gophermart@localhost:5432/gophermart?sslmode=disable" up

migrate-down:
	migrate -path internal/db/migrations -database "postgresql://gophermart:gophermart@localhost:5432/gophermart?sslmode=disable" down 1

build:
	go build -o bin/gophermart ./cmd/gophermart/...

run: build
	go run ./cmd/gophermart/...