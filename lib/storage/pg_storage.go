package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/sivukhin/gobughunt/lib/logging"
)

type PgStorage struct{ *pgx.Conn }

func NewPgStorage(ctx context.Context, connectionString string) (PgStorage, error) {
	connection, err := pgx.Connect(ctx, connectionString)
	if err != nil {
		return PgStorage{}, fmt.Errorf("failed to connect: %w", err)
	}
	config := connection.Config()
	logging.Logger.Infof("connection established: host=%v, db=%v", config.Host, config.Database)
	return PgStorage{connection}, nil
}
