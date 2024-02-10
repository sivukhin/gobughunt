package lib

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sivukhin/gobughunt/lib/dto"
)

func TestLint(t *testing.T) {
	repo := dto.RepoInstance{
		RepoId:        "test-repo",
		GitUrl:        "https://github.com/drakkan/sftpgo",
		GitCommitHash: "c8da72a7f7ea10a3ca853f66f0ad80855893b775",
	}
	linter := dto.LinterInstance{
		LinterId:           "test-linter",
		DockerImage:        "sivukhinnikita/govanish:3.0.0",
		DockerImageShaHash: "e0a20eadc9d60c1a67f9b3cb7aa1431b2fb8b93c1cf6ae002e297f2465c16dd8",
	}
	highlights, err := NaiveLinting.Run(context.Background(), repo, linter)
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
