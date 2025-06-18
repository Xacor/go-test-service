package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Xacor/go-test-service/log-consumer/logger"
	"github.com/Xacor/go-test-service/pkg/log"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	natsStream      = "LOGS"
	natsSubject     = "logs.goods"
	durableConsumer = "log_consumer"
)

var (
	buffer     []log.LogEntry
	bufferLock sync.Mutex
)

func main() {
	if err := godotenv.Load(); err != nil {
		zap.S().Info("No .env file found, using environment variables directly")
	}

	natsURL := getEnv("NATS_URL", nats.DefaultURL)
	clickhouseAddr := getEnv("CLICKHOUSE_ADDR", "")

	if err := logger.InitClickHouse(clickhouseAddr); err != nil {
		zap.S().Fatal("ClickHouse init failed: %v", zap.Error(err))
	}

	nc, err := nats.Connect(natsURL)
	if err != nil {
		zap.S().Fatal("NATS connection failed:", zap.Error(err))
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		zap.S().Fatal("JetStream init failed:", zap.Error(err))
	}

	_, err = js.AddStream(&nats.StreamConfig{
		Name:     natsStream,
		Subjects: []string{natsSubject},
	})
	if err != nil {
		zap.S().Fatal("Add stream:", zap.Error(err))
	}

	sub, err := js.PullSubscribe(natsSubject, durableConsumer)
	if err != nil {
		zap.S().Fatal("subscription failed:", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		zap.S().Info("Shutting down...")
		cancel()
	}()

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				flush()
				return
			case <-ticker.C:
				flush()
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgs, err := sub.Fetch(10, nats.MaxWait(2*time.Second))
			if err != nil && err != nats.ErrTimeout {
				zap.S().Error("Fetch error:", zap.Error(err))
				continue
			}
			for _, msg := range msgs {
				logEntry, err := logger.ParseLog(msg.Data)
				if err != nil {
					zap.S().Error("Invalid log entry: %v", zap.Error(err))
					msg.Nak()
					continue
				}
				addToBuffer(logEntry)
				msg.Ack()
			}
		}
	}
}

func addToBuffer(entry log.LogEntry) {
	bufferLock.Lock()
	defer bufferLock.Unlock()
	buffer = append(buffer, entry)
}

func flush() {
	bufferLock.Lock()
	defer bufferLock.Unlock()

	if len(buffer) == 0 {
		return
	}

	err := logger.FlushToClickHouse(buffer)
	if err != nil {
		zap.S().Error("Flush error: %v", zap.Error(err))
		return
	}

	zap.S().Info("Flushed log(s) to ClickHouse", zap.Int("Count", len(buffer)))
	buffer = nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
