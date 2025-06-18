package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/Xacor/go-test-service/internal/api"
	"github.com/Xacor/go-test-service/internal/db"
	"github.com/Xacor/go-test-service/internal/model"
	"github.com/Xacor/go-test-service/pkg/migrator"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		zap.L().Info("No .env file found, using environment variables directly")
	}

	pgDSN := getEnv("POSTGRES_DSN", "")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
	redisPass := getEnv("REDIS_PASS", "")
	natsURL := getEnv("NATS_URL", nats.DefaultURL)
	migrationFile := getEnv("MIGRATION_FILE", "migrations.sql")

	pg, err := pgxpool.New(context.Background(), pgDSN)
	if err != nil {
		zap.L().Fatal("Postgres connection failed:", zap.Error(err))
	}

	err = migrator.RunMigrations(pg, migrationFile)
	if err != nil {
		zap.L().Fatal("Run migration:", zap.Error(err))
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPass,
		DB:       0,
	})

	if _, err := redisClient.Ping(context.TODO()).Result(); err != nil {
		zap.L().Fatal("Redis ping failed:", zap.Error(err))
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		zap.L().Fatal("NATS connection failed:", zap.Error(err))
	}
	defer nc.Close()
	js, _ := jetstream.New(nc)

	mdl := model.New(redisClient, js, db.NewDB(pg))
	srv := api.NewServer(mdl)
	if err := srv.Listen(); err != nil {
		zap.L().Fatal("Listen http")
	}
	defer srv.Shutdown(context.Background())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
