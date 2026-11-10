# toll-calculator

A toll calculation system built from Go microservices. Simulated on-board units (OBU) send GPS
points, the distance driven by each vehicle is calculated and summed up, and an invoice can be
requested through an HTTP gateway.

```
 OBU ──websocket──▶ receiver ──Kafka──▶ calculator ──gRPC──▶ aggregator ◀──gRPC── gateway ◀──HTTP── client
                                                                 │
                                                          Prometheus / Grafana
```

| Service      | Command           | What it does                                                         |
|--------------|-------------------|----------------------------------------------------------------------|
| `obu`        | `make obu`        | simulates 20 vehicles, sends their coordinates over websocket        |
| `receiver`   | `make receiver`   | accepts websocket connections, publishes points to Kafka             |
| `calculator` | `make calculator` | consumes points, calculates distance per vehicle (haversine, km)     |
| `aggregator` | `make agg`        | sums distances per vehicle, calculates invoices (HTTP + gRPC)        |
| `gateway`    | `make gate`       | public HTTP API for invoices                                         |

All services stop gracefully on `Ctrl+C` / `SIGTERM`.

## Requirements

- Go 1.26+
- Docker with the compose plugin
- For code generation only: [buf](https://buf.build) and the Go protobuf plugins
  ```
  go install github.com/bufbuild/buf/cmd/buf@v1.75.0
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.2
  ```
  Make sure `$(go env GOPATH)/bin` is in your `PATH`.

The generated code in `gen/` is committed, so building and running the project does not need `buf`.

## Quick start

1. Start the infrastructure:
   ```
   make up
   ```
2. Create the local config (the defaults work out of the box):
   ```
   cp .env.example .env
   ```
3. Start the services, each in its own terminal, in this order:
   ```
   make agg
   make calculator
   make receiver
   make obu
   make gate
   ```
4. Take an `obuID` from the aggregator log and request its invoice:
   ```
   curl localhost:6000/invoice/<obuID>
   ```

Stop the infrastructure with `make down`.

## Infrastructure

Kafka (KRaft mode, no Zookeeper), Kafka UI, Prometheus and Grafana run in Docker (`docker-compose.yml`).

| Service    | Address                | Notes                                      |
|------------|------------------------|--------------------------------------------|
| Kafka      | `localhost:9092`       | used by receiver and calculator            |
| Kafka UI   | http://localhost:8090  | topics, partitions, messages               |
| Prometheus | http://localhost:9091  | scrapes the aggregator on `localhost:4000` |
| Grafana    | http://localhost:3030  | login `admin` / `admin`                    |

Prometheus scrapes the aggregator through `host.docker.internal:4000` (see `.config/prometheus.yml`),
so start the aggregator to see its target as UP in http://localhost:9091/targets.

Grafana is provisioned from `.config/grafana/provisioning`: the Prometheus datasource and the
"Aggregator" dashboard (Dashboards → Toll Calculator). After logging in Grafana asks to change the
password — you can press **Skip**.

## Configuration

Services read their settings from environment variables. A `.env` file in the working directory is
loaded if it exists; variables already set in the environment take precedence over it.
See `.env.example` for all variables.

| Variable            | Default                   | Used by              |
|---------------------|---------------------------|----------------------|
| `KAFKA_BROKERS`     | `localhost:9092`          | receiver, calculator |
| `KAFKA_TOPIC`       | `obuData`                 | receiver, calculator |
| `OBU_WS_URL`        | `ws://127.0.0.1:30000/ws` | obu                  |
| `RECEIVER_ADDR`     | `:30000`                  | receiver             |
| `CALC_KAFKA_GROUP`  | `distance-calculator`     | calculator           |
| `AGG_HTTP_ENDPOINT` | `:4000`                   | aggregator           |
| `AGG_GRPC_ENDPOINT` | `:3001`                   | aggregator           |
| `AGG_STORE_TYPE`    | `memory`                  | aggregator           |
| `AGG_GRPC_URL`      | `127.0.0.1:3001`          | calculator, gateway  |
| `GATEWAY_ADDR`      | `:6000`                   | gateway              |

The gateway also accepts `-listenAddr`, which overrides `GATEWAY_ADDR`.

## API

### Gateway (HTTP, `:6000`)

| Method | Path               | Response                                                   |
|--------|--------------------|------------------------------------------------------------|
| `GET`  | `/invoice/{obuID}` | `200` invoice, `400` invalid id, `404` no data for the OBU |

```
$ curl localhost:6000/invoice/7
{"obuID":7,"totalDistance":10,"totalAmount":31.5}
```

### Aggregator (HTTP, `:4000`)

| Method | Path                | Description                         |
|--------|---------------------|-------------------------------------|
| `POST` | `/aggregate`        | body `{"obuID":7,"value":10,"unix":1}` |
| `GET`  | `/invoice?obu={id}` | invoice for the OBU, `404` if none  |
| `GET`  | `/metrics`          | Prometheus metrics                  |

### Aggregator (gRPC, `:3001`)

Service `aggregator.v1.AggregatorService`, defined in `api/proto/aggregator/v1/aggregator.proto`:
`Aggregate` and `GetInvoice` (`NotFound` if there is no data for the OBU).
Server reflection is enabled, so it can be called with [grpcurl](https://github.com/fullstorydev/grpcurl):

```
grpcurl -plaintext localhost:3001 list
grpcurl -plaintext -d '{"obu_id": 7}' localhost:3001 aggregator.v1.AggregatorService/GetInvoice
```

## Development

Run the tests:
```
go test -race ./...
```

After changing a `.proto` file, lint it and regenerate the code into `gen/`:
```
make buf
```

### Project layout

```
cmd/<service>/main.go   entry points: config, wiring, start and graceful shutdown
internal/<service>/     business logic of each service
internal/aggregator/client/  HTTP and gRPC clients of the aggregator (used by calculator and gateway)
internal/config/        environment and .env loading
internal/httpx/         shared HTTP helpers: JSON responses, error handling, request logging
internal/types/         data types shared between services
api/proto/              protobuf definitions
gen/                    generated protobuf / gRPC code (do not edit)
go-kit-example/         separate go-kit learning sandbox, not part of the pipeline
```
