# toll-calculator

## Infrastructure

Kafka (KRaft mode, no Zookeeper), Kafka UI, Prometheus and Grafana run in Docker:
```
make up     # docker compose up -d
make down   # docker compose down
```

| Service    | Address                | Notes                                         |
|------------|------------------------|-----------------------------------------------|
| Kafka      | `localhost:9092`       | used by the Go services                       |
| Kafka UI   | http://localhost:8090  | topics, partitions, messages                  |
| Prometheus | http://localhost:9091  | scrapes the aggregator on `localhost:4000`    |
| Grafana    | http://localhost:3030  | login `admin` / `admin`                       |

The Go services run on the host:
```
cp .env.example .env
make agg
make calculator
make receiver
make obu
make gate
```

Prometheus scrapes the aggregator through `host.docker.internal:4000` (see `.config/prometheus.yml`),
so start the aggregator to see its target as UP in http://localhost:9091/targets.

Grafana is provisioned from `.config/grafana/provisioning`: the Prometheus datasource and the
"Aggregator" dashboard (Dashboards → Toll Calculator). After logging in Grafana asks to change the
password — you can press **Skip**.

## Installing protobuf compiler (protoc compiler)

For linux users or (WSL2)
```
sudo apt install -y protobuf-compiler
```

For Mac users you can use Brew for this
```
brew install protobuf
```

## Installing GRPC and Protobuffer plugins for Golang.
1. Protobuffers
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
```

2. GRPC
```
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
```

3. NOTE that you need to set the /go/bin directory in your path
Just like this or whatever your go directory lives
```
PATH="${PATH}:${HOME}/go/bin"
```

4. install the package dependencies
4.1 protobuffer package
```
go get google.golang.org/protobuf
```
4.2 grpc package
```
go get google.golang.org/grpc
```
