.PHONY: up down obu receiver calculator agg invoicer proto buf gate

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

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative internal/types/ptypes.proto

buf:
	buf lint
	buf generate

gate:
	@go build -o bin/gate ./cmd/gateway
	@./bin/gate
