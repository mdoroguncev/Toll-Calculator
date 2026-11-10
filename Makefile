.PHONY: up down obu receiver calculator agg invoicer buf gate migrate migrate-down

up:
	@docker compose up -d

down:
	@docker compose down

obu:
	@go build -o bin/obu ./cmd/obu/main.go
	@./bin/obu

receiver:
	@go build -o bin/receiver ./cmd/receiver
	@./bin/receiver

calculator:
	@go build -o bin/calculator ./cmd/calculator
	@./bin/calculator

agg:
	@go build -o bin/agg ./cmd/aggregator
	@./bin/agg

buf:
	buf lint
	buf generate

gate:
	@go build -o bin/gate ./cmd/gateway
	@./bin/gate

DB_DSN ?= postgres://toll:toll@localhost:5433/toll?sslmode=disable

migrate:
	goose -dir migrations postgres "$(DB_DSN)" up

migrate-down:
	goose -dir migrations postgres "$(DB_DSN)" down