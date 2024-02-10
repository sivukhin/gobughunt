package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sivukhin/gobughunt/lib/logging"
)

type PgStorage struct{ *pgxpool.Pool }

func NewPgStorage(ctx context.Context, connectionString string) (PgStorage, error) {
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return PgStorage{}, fmt.Errorf("failed to connect: %w", err)
	}
	config := pool.Config().ConnConfig.Config
	logging.Logger.Infof("connection established: host=%v, db=%v", config.Host, config.Database)
	return PgStorage{pool}, nil
}
