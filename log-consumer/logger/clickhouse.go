package logger

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/Xacor/go-test-service/pkg/log"
	"github.com/pkg/errors"
)

var chConn clickhouse.Conn

func InitClickHouse(addr string) error {
	var err error
	chConn, err = clickhouse.Open(&clickhouse.Options{
		Addr: []string{addr},
		Auth: clickhouse.Auth{
			Database: "logs",
			Username: "default",
			Password: "password",
		},
		Debug: false,
	})
	if err != nil {
		return err
	}
	err = chConn.Ping(context.Background())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	createDDL := `
	CREATE TABLE IF NOT EXISTS logs.goods (
		id          Int32,
		project_id  Int32,
		name        String,
		description String,
		priority    Int32,
		removed     UInt8,
		event_time  DateTime


		INDEX idx_id id TYPE minmax GRANULARITY 1,
		INDEX idx_project_id project_id TYPE minmax GRANULARITY 1,
    	INDEX idx_name name TYPE tokenbf_v1(512, 3, 0) GRANULARITY 1
	) ENGINE = MergeTree()
	ORDER BY (event_time)
	`

	if err := chConn.Exec(ctx, createDDL); err != nil {
		return errors.Wrap(err, "failed to create ClickHouse table")
	}

	return nil
}

func FlushToClickHouse(entries []log.LogEntry) error {
	if len(entries) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	batch, err := chConn.PrepareBatch(ctx, "INSERT INTO logs.goods")
	if err != nil {
		return err
	}

	for _, e := range entries {
		err = batch.Append(
			e.ID,
			e.ProjectID,
			e.Name,
			e.Description,
			e.Priority,
			boolToUint8(e.Removed),
			e.EventTime,
		)
		if err != nil {
			return err
		}
	}

	return batch.Send()
}

func boolToUint8(b bool) uint8 {
	if b {
		return 1
	}
	return 0
}

func ParseLog(data []byte) (log.LogEntry, error) {
	var logEntry log.LogEntry
	err := json.Unmarshal(data, &logEntry)
	return logEntry, err
}
