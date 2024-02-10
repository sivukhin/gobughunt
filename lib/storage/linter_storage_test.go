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

func TestLinterStorageInsert(t *testing.T) {
	storage, err := NewPgStorage(context.Background(), "postgresql://postgres:local@localhost/postgres")
	require.Nil(t, err)

	linterStorage := PgLinterStorage(storage)
	require.Nil(t, linterStorage.InitTables(context.Background()))

	err = linterStorage.AddOrUpdate(context.Background(), dto.Linter{
		Meta: dto.LinterMeta{
			LinterId:        utils.Must(guid.NewV4()).String(),
			LinterGitUrl:    "git",
			LinterGitBranch: "branch",
		},
	}, time.Now())
	require.Nil(t, err)
}

func TestLinterStorageUpsert(t *testing.T) {
	storage, err := NewPgStorage(context.Background(), "postgresql://postgres:local@localhost/postgres")
	require.Nil(t, err)

	linterStorage := PgLinterStorage(storage)
	require.Nil(t, linterStorage.InitTables(context.Background()))

	meta := dto.LinterMeta{
		LinterId:        utils.Must(guid.NewV4()).String(),
		LinterGitUrl:    "git",
		LinterGitBranch: "branch",
	}
	now := time.Now()
	err = linterStorage.AddOrUpdate(context.Background(), dto.Linter{Meta: meta}, now)
	require.Nil(t, err)
	instance := &dto.LinterInstance{
		LinterId:           meta.LinterId,
		DockerImage:        "docker-image",
		DockerImageShaHash: "docker-sha",
	}
	err = linterStorage.AddOrUpdate(context.Background(), dto.Linter{Meta: meta, Instance: instance}, now.Add(10*time.Minute))
	require.Nil(t, err)
}