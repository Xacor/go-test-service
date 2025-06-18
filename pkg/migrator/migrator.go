package migrator

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

func RunMigrations(db *pgxpool.Pool, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return errors.Wrap(err, "cannot read migrations file")
	}

	queries := strings.Split(string(data), ";")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}

		_, err := db.Exec(ctx, q)
		if err != nil {
			return errors.Wrap(err, "execute migration")
		}
	}

	return nil
}
