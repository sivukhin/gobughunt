package lib

import (
	"context"
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sivukhin/gobughunt/lib/dto"
)

func TestCloneRepository(t *testing.T) {
	t.Run("fetch branch", func(t *testing.T) {
		d := t.TempDir()
		err := NaiveGitApi.Clone(context.Background(), "https://github.com/sivukhin/govanish", dto.GitRef{Branch: "master"}, d)
		require.Nil(t, err)
		{
			stat, err := os.Stat(path.Join(d, "go.mod"))
			require.Nil(t, err)
			require.Equal(t, "go.mod", stat.Name())
		}
		{
			_, err := os.Stat(path.Join(d, "absent"))
			var expected *fs.PathError
			require.ErrorAs(t, err, &expected)
		}
	})
	t.Run("fetch commit hash", func(t *testing.T) {
		d := t.TempDir()
		err := NaiveGitApi.Clone(context.Background(), "https://github.com/sivukhin/govanish", dto.GitRef{CommitHash: "a4d8a6fc86afcc1aa8c4c919d3073ee996eb139d"}, d)
		require.Nil(t, err)
		{
			stat, err := os.Stat(path.Join(d, "go.mod"))
			require.Nil(t, err)
			require.Equal(t, "go.mod", stat.Name())
		}
		{
			_, err := os.Stat(path.Join(d, "absent"))
			var expected *fs.PathError
			require.ErrorAs(t, err, &expected)
		}
	})
}
