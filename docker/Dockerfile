FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /product-service ./cmd/server

FROM scratch

COPY --from=builder /product-service /product-service
COPY --from=builder /app/migrations /migrations

EXPOSE 50051

ENTRYPOINT ["/product-service"]
