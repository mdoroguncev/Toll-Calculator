# toll-calculator

Start Kafka and Zookeeper:
```
docker-compose up
```

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


## Installing Prometheus
Download the Prometheus binary next to the project (into `../prometheus`)
```
cd ..
V=3.15.0
curl -LO https://github.com/prometheus/prometheus/releases/download/v$V/prometheus-$V.linux-amd64.tar.gz
mkdir -p prometheus
tar -xzf prometheus-$V.linux-amd64.tar.gz -C prometheus --strip-components=1
rm prometheus-$V.linux-amd64.tar.gz
cd Toll-Calculator
```

The config lives in `.config/prometheus.yml` and scrapes the aggregator metrics from `localhost:4000/metrics`.
Check the config
```
../prometheus/promtool check config ./.config/prometheus.yml
```

Run Prometheus (from the project root, after `make agg`)
```
../prometheus/prometheus --config.file=./.config/prometheus.yml --web.listen-address=:9091
```

Or simply
```
make prometheus
```

The UI is available at http://localhost:9091. Example query:
```
go_goroutines{job="aggregator"}
```

Installing prometheus golang client
```
go get github.com/prometheus/client_golang/prometheus
```

## Installing Grafana
Download the Grafana binary next to the project (into `../grafana`)
```
cd ..
V=13.2.4
curl -LO https://dl.grafana.com/oss/release/grafana-$V.linux-amd64.tar.gz
mkdir -p grafana
tar -xzf grafana-$V.linux-amd64.tar.gz -C grafana --strip-components=1
rm grafana-$V.linux-amd64.tar.gz
cd Toll-Calculator
```

The Prometheus datasource (`http://localhost:9091`) is provisioned from `.config/grafana/provisioning`.

Run Grafana (from the project root, after Prometheus is running)
```
GF_PATHS_PROVISIONING=$PWD/.config/grafana/provisioning ../grafana/bin/grafana server --homepath "$(realpath ../grafana)"
```

Or simply
```
make grafana
```

The UI is available at http://localhost:3000.
This is a login page, not a registration: username `admin`, password `admin`. After logging in Grafana asks to change the password — you can press **Skip**.
The "Aggregator" dashboard (`.config/grafana/provisioning/dashboards/aggregator.json`) is under Dashboards → Toll Calculator.
