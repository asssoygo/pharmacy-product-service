# pharmacy-product-service

gRPC microservice for product and category management in the Pharmacy Management System.

## Architecture

Clean Architecture with four layers:

```
cmd/server/          — entrypoint: wires dependencies, runs gRPC server
config/              — env-based configuration
internal/
  domain/            — entities, repository interfaces, domain errors
  usecase/           — business logic (no framework dependencies)
  repository/
    postgres/        — PostgreSQL implementation of repository interfaces
    redis/           — Redis cache-aside wrapper
  handler/grpc/      — gRPC handler: proto ↔ domain conversion
migrations/          — golang-migrate SQL files
docker/              — Dockerfile (multi-stage, final FROM scratch)
```

## Proto dependency

Imports generated gRPC code from [pharmacy-proto](https://github.com/asssoygo/pharmacy-proto).

## Configuration

| Env var       | Default        | Description          |
|---------------|---------------|----------------------|
| `GRPC_PORT`   | `50051`        | gRPC listen port     |
| `DB_HOST`     | `localhost`    | PostgreSQL host      |
| `DB_PORT`     | `5432`         | PostgreSQL port      |
| `DB_USER`     | `postgres`     | PostgreSQL user      |
| `DB_PASSWORD` | `postgres`     | PostgreSQL password  |
| `DB_NAME`     | `pharmacy`     | PostgreSQL database  |
| `REDIS_ADDR`  | `localhost:6379` | Redis address      |

Copy `.env.example` to `.env` and adjust values before running locally.

## Running locally

```bash
# start dependencies
docker compose up -d postgres redis

# run the service
go run ./cmd/server
```

## Docker

```bash
docker build -f docker/Dockerfile -t pharmacy-product-service .
docker run -e DB_HOST=postgres -e REDIS_ADDR=redis:6379 pharmacy-product-service
```

## Tests

```bash
go test ./...
```

## gRPC endpoints

All 12 RPCs from `ProductService` (defined in pharmacy-proto):

- `CreateProduct` / `GetProduct` / `GetProducts` / `UpdateProduct` / `DeleteProduct`
- `SearchProducts` / `GetLowStockProducts` / `UpdateStock`
- `GetExpiredProducts` / `GetProductsByCategory`
- `CreateCategory` / `GetCategories`
