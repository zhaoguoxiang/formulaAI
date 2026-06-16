.PHONY: run test migrate-up migrate-down

DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= formula
DB_PASSWORD ?= test
DB_NAME ?= formula_ai

DB_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_PATH ?= migrations

run:
	go run ./cmd/server/

test:
	go test ./... -v

migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down
