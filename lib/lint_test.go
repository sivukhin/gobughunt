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
		t.Log(lines, err)
		require.ErrorIs(t, err, DockerNonZeroExitCodeErr)
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
		t.Log(lines, err)
		require.NotNil(t, err)
	})
	t.Run("mem", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		d := t.TempDir()
		lines, err := NaiveDockerApi{MemoryBytes: 128 * 1024 * 1024}.Exec(
			ctx,
			"docker.io/sivukhinnikita/dumb-mem:1.0.0@sha256:1405e034c51723503eff603a3e0134be2b1471b216161679011e2fc9e6030131",
			d,
			"/src",
		)
		t.Log(lines, err)
		require.ErrorContains(t, err, "non zero exit code: 137")
	})
	t.Run("fork", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		d := t.TempDir()
		lines, err := NaiveDockerApi{MemoryBytes: 128 * 1024 * 1024, CpuMilli: 100, PidLimit: 1024}.Exec(
			ctx,
			"docker.io/sivukhinnikita/dumb-fork:1.0.0@sha256:4313537ddc991431929700790b060a3daa639c37144d745fb364a4655eabc989",
			d,
			"/src",
		)
		t.Log(lines, err)
		require.ErrorContains(t, err, "non zero exit code: 2")
	})
}

func TestLint(t *testing.T) {
	repo := dto.RepoInstance{
		Id:            "test-repo",
		GitUrl:        "https://github.com/gocolly/colly",
		GitCommitHash: "3c987f1982edbb5ba8876eef56dd35e1ff05932a",
	}
	linter := dto.LinterInstance{
		Id:                 "test-linter",
		DockerImage:        "docker.io/sivukhinnikita/nilaway:4.0.0",
		DockerImageShaHash: "8639b178ae41861765a8149f298353bcac36de4b9e1de3bfe175d92953591fd0",
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
