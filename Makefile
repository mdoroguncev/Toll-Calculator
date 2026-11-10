.PHONY: up down obu receiver calculator agg invoicer buf gate

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
