package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sivukhin/gobughunt/lib/logging"
	"github.com/sivukhin/gobughunt/storage/db"
)

func NewPgStorage(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("ping failed: %w", err)
	}
	config := pool.Config().ConnConfig.Config
	logging.Logger.Infof("connection established: host=%v, db=%v", config.Host, config.Database)
	return pool, nil
}

func NewPgQueries(ctx context.Context, connectionString string) (*db.Queries, error) {
	pool, err := NewPgStorage(ctx, connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create pg storage for queries: %w", err)
	}
	return db.New(pool), nil
}

func ViolatesUniqueConstraint(err error) bool {
	var pgError *pgconn.PgError
	if errors.As(err, &pgError) {
		return pgError.Code == "23505" /* unique_violation */
	}
	return false
}

func TryGetText(text pgtype.Text) *string {
	if !text.Valid {
		return nil
	}
	return &text.String
}

func TryGetDurationSec(duration pgtype.Interval) *float64 {
	if !duration.Valid {
		return nil
	}
	value := time.Duration(duration.Microseconds) * time.Microsecond
	seconds := value.Seconds()
	return &seconds
}
