FROM golang:1.25.0-bookworm AS builder

WORKDIR /pdd-service

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /build/pdd-service ./cmd/pdd-service

FROM alpine:3.22

WORKDIR /pdd-service

COPY --from=builder /build/pdd-service /pdd-service/pdd-service

CMD ["/pdd-service/pdd-service"]
