FROM golang:1.26-bookworm@sha256:dc40a13b61304647ed7e5d581d246b86413dfc29d63c6d018310feb93cd98fc0 AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG SERVICE
RUN CGO_ENABLED=1 go build -o /out/app ./cmd/${SERVICE}

FROM gcr.io/distroless/base-debian12@sha256:fabbf1c0c357a3d42550111351daed089b20a2c954df13ee2fcff60602515e84
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]
