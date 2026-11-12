FROM golang:1.26-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG SERVICE
RUN CGO_ENABLED=1 go build -o /out/app ./cmd/${SERVICE}

FROM gcr.io/distroless/base-debian12
COPY --from=build /out/app /app
ENTRYPOINT ["/app"]