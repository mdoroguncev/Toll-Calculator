.PHONY: obu receiver calculator agg invoicer proto gate prometheus grafana

obu:
	@go build -o bin/obu obu/main.go
	@./bin/obu

receiver:
	@go build -o bin/receiver ./data_receiver
	@./bin/receiver

calculator:
	@go build -o bin/calculator ./distance_calculator
	@./bin/calculator

agg:
	@go build -o bin/agg ./aggregator
	@./bin/agg

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative types/ptypes.proto

gate:
	@go build -o bin/gate gateway/main.go
	@./bin/gate

prometheus:
	@../prometheus/prometheus --config.file=$(CURDIR)/.config/prometheus.yml --web.listen-address=:9091

grafana:
	@GF_PATHS_PROVISIONING=$(CURDIR)/.config/grafana/provisioning GF_PATHS_DATA=$(CURDIR)/.config/grafana/data ../grafana/bin/grafana server --homepath $(abspath ../grafana)
