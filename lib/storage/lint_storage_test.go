package storage

import (
	"context"
	_ "embed"
	"testing"
	"time"

	"github.com/Microsoft/go-winio/pkg/guid"
	"github.com/stretchr/testify/require"

	"github.com/sivukhin/gobughunt/lib/dto"
	"github.com/sivukhin/gobughunt/lib/utils"
)

func TestSet(t *testing.T) {
	storage, err := NewPgStorage(context.Background(), "postgresql://postgres:local@localhost/postgres")
	require.Nil(t, err)

	lintStorage := PgLintStorage(storage)

	task := dto.LintTask{
		LintId: utils.Must(guid.NewV4()).String(),
		Linter: dto.LinterInstance{
			LinterId:           "test-linter-" + utils.Must(guid.NewV4()).String(),
			DockerImage:        "sivukhinnikita/govanish:1.0.0",
			DockerImageShaHash: utils.Must(guid.NewV4()).String(),
		},
		Repo: dto.RepoInstance{
			RepoId:        "test-repo-" + utils.Must(guid.NewV4()).String(),
			GitUrl:        "https://github.com/drakkan/sftpgo",
			GitCommitHash: utils.Must(guid.NewV4()).String(),
		},
	}
	err = lintStorage.TryAdd(context.Background(), task, time.Now())
	require.Nil(t, err)
	task, err = lintStorage.TryTake(context.Background(), time.Now(), time.Now())
	t.Log(task)
	require.Nil(t, err)
	result := dto.LintResult{
		LintStatus:        dto.Failed,
		LintStatusComment: "non zero exit code",
		LintDuration:      time.Minute,
		Highlights: []dto.LintHighlightSnippet{{
			LintHighlight: dto.LintHighlight{
				Path:        "/some/path",
				StartLine:   1,
				EndLine:     2,
				Explanation: "lines were vanished",
			},
			Snippet: "if err != nil {\n    return err\n}\n",
		}},
	}
	err = lintStorage.Set(context.Background(), task, result, time.Now())
	require.Nil(t, err)
}
