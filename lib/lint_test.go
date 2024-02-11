package lib

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sivukhin/gobughunt/lib/dto"
)

func TestDumb(t *testing.T) {
	t.Run("non-zero exit code", func(t *testing.T) {
		d := t.TempDir()
		lines, err := Docker.Exec(
			context.Background(),
			"docker.io/sivukhinnikita/dumb-fail:1.0.0@sha256:acc0726e21d1e9ea1c205216ad74c9d647b8f126d26af3586603462255fef969",
			d,
			"/src",
		)
		require.ErrorIs(t, err, DockerNonZeroExitCodeErr)
		t.Log(lines, err)
	})
	t.Run("long", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		d := t.TempDir()
		lines, err := Docker.Exec(
			ctx,
			"docker.io/sivukhinnikita/dumb-long:1.0.0@sha256:79844422ce2abefacdd5451a098944293864942e37e10b0e84c8b687c098780a",
			d,
			"/src",
		)
		require.NotNil(t, err)
		t.Log(lines, err)
	})
}

func TestLint(t *testing.T) {
	repo := dto.RepoInstance{
		Id:            "test-repo",
		GitUrl:        "https://github.com/gin-gonic/gin",
		GitCommitHash: "bb3519d26f52835cf00e5e430b52651a9c378c97",
	}
	linter := dto.LinterInstance{
		Id:                 "test-linter",
		DockerImage:        "docker.io/sivukhinnikita/govanish:8.0.0",
		DockerImageShaHash: "4257681aec436662049ed919c9aa2e8028e59e647efa4c996495a308c48dd77d",
	}
	highlights, err := Lint.Run(context.Background(), repo, linter)
	t.Log(highlights, err)
}

func TestExtractHighlightSnippetsForFile(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		snippets, err := ExtractHighlightSnippetsForFile([]byte(`line 1
line 2
line 3
line 4
line 5`), []dto.LintHighlight{{StartLine: 1, EndLine: 1}, {StartLine: 3, EndLine: 3}})
		require.Nil(t, err)
		require.Equal(t, []dto.LintHighlightSnippet{{
			LintHighlight: dto.LintHighlight{StartLine: 1, EndLine: 1},
			Snippet: dto.HighlightSnippet{
				StartLine: 1,
				EndLine:   2,
				Code: `line 1
line 2`,
			},
		}, {
			LintHighlight: dto.LintHighlight{StartLine: 3, EndLine: 3},
			Snippet: dto.HighlightSnippet{
				StartLine: 2,
				EndLine:   4,
				Code: `line 2
line 3
line 4`,
			},
		}}, snippets)
	})
}

func TestExtractHighlights(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		highlights, skipped := ExtractHighlights([]string{
			`2024/02/10 11:09:26 module path: /home/sivukhin/code/test-go-kek/repo`,
			`2024/02/10 11:09:26 ready to compile project at path '/home/sivukhin/code/test-go-kek/repo' for assembly inspection`,
			`2024/02/10 11:09:26 ready to parse assembly output`,
			`2024/02/10 11:09:31 ready to normalize assembly lines (size 169)`,
			`2024/02/10 11:09:32 built func registry: 2779 entries`,
			`2024/02/10 11:09:32 ready to analyze module AST`,
			`::warning file=internal/vfs/cryptfs.go,line=262,endLine=263::seems like code vanished from compiled binary`,
			`::warning file=internal/service/signals_unix.go,line=77::seems like code vanished from compiled binary`,
			`::error file=internal/service/signals_unix.go,line=77,title=Bug::seems like code vanished from compiled binary`,
		})
		require.False(t, skipped)
		require.Equal(t, []dto.LintHighlight{{
			Path:        "internal/vfs/cryptfs.go",
			StartLine:   262,
			EndLine:     263,
			Explanation: "seems like code vanished from compiled binary",
		}, {
			Path:        "internal/service/signals_unix.go",
			StartLine:   77,
			EndLine:     77,
			Explanation: "seems like code vanished from compiled binary",
		}, {
			Path:        "internal/service/signals_unix.go",
			StartLine:   77,
			EndLine:     77,
			Explanation: "Bug: seems like code vanished from compiled binary",
		}}, highlights)
		t.Log(highlights, skipped)
	})
	t.Run("skipped", func(t *testing.T) {
		highlights, skipped := ExtractHighlights([]string{
			`2024/02/10 11:09:26 module path: /home/sivukhin/code/test-go-kek/repo`,
			`2024/02/10 11:09:26 ready to compile project at path '/home/sivukhin/code/test-go-kek/repo' for assembly inspection`,
			`2024/02/10 11:09:26 ready to parse assembly output`,
			`2024/02/10 11:09:31 ready to normalize assembly lines (size 169)`,
			`2024/02/10 11:09:32 built func registry: 2779 entries`,
			`2024/02/10 11:09:32 ready to analyze module AST`,
			`::skip`,
		})
		require.True(t, skipped)
		require.Nil(t, highlights)
	})
}
