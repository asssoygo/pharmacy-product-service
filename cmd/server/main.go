package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"

	productpb "github.com/asssoygo/pharmacy-proto/gen/go/product"
	"github.com/asssoygo/pharmacy-product-service/config"
	grpchandler "github.com/asssoygo/pharmacy-product-service/internal/handler/grpc"
	pgrepo "github.com/asssoygo/pharmacy-product-service/internal/repository/postgres"
	rediscache "github.com/asssoygo/pharmacy-product-service/internal/repository/redis"
	"github.com/asssoygo/pharmacy-product-service/internal/usecase"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		log.Fatalf("postgres open: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("postgres ping: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("migrate driver: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		log.Fatalf("migrate init: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("migrate up: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	productPG := pgrepo.NewProductRepository(db)
	categoryPG := pgrepo.NewCategoryRepository(db)
	cachedProduct := rediscache.NewCachedProductRepository(productPG, rdb)
	uc := usecase.NewProductUseCase(cachedProduct, categoryPG)
	handler := grpchandler.NewProductHandler(uc)

	srv := grpc.NewServer()
	productpb.RegisterProductServiceServer(srv, handler)
	reflection.Register(srv)

	addr := fmt.Sprintf(":%s", cfg.GRPCPort)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	log.Printf("gRPC server listening on %s", addr)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
