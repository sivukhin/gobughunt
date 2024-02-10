package storage

import (
	"context"
	"testing"
	"time"

	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/stretchr/testify/require"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/utils"
)

func TestRepoStorageInsert(t *testing.T) {
	storage, err := NewPgStorage(context.Background(), "postgresql://postgres:local@localhost/postgres")
	require.Nil(t, err)

	repoStorage := PgRepoStorage(storage)

	err = repoStorage.AddOrUpdate(context.Background(), dto.Repo{
		Meta: dto.RepoMeta{
			Id:        utils.Must(guid.NewV4()).String(),
			GitUrl:    "git",
			GitBranch: "branch",
		},
	}, time.Now())
	require.Nil(t, err)
}
