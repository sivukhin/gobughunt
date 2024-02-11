package storage

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/sivukhin/gobughunt/lib/dto"
)

type StatDto struct {
	TotalHighlightDedup    int `json:"totalHighlightDedup"`
	PendingHighlightDedup  int `json:"pendingHighlightDedup"`
	RejectedHighlightDedup int `json:"rejectedHighlightDedup"`
	AcceptedHighlightDedup int `json:"acceptedHighlightDedup"`
}

type LinterDto struct {
	Id                 string  `json:"id"`
	GitUrl             string  `json:"gitUrl"`
	GitBranch          string  `json:"gitBranch"`
	DockerImage        *string `json:"dockerImage"`
	DockerImageShaHash *string `json:"dockerImageShaHash"`
	*StatDto
}

type RepoDto struct {
	Id            string  `json:"id"`
	GitUrl        string  `json:"gitUrl"`
	GitBranch     string  `json:"gitBranch"`
	GitCommitHash *string `json:"gitCommitHash"`
	*StatDto
}

type LintTaskDto struct {
	Id              string    `json:"id"`
	Status          string    `json:"status"`
	StatusComment   *string   `json:"statusComment"`
	LintDurationSec *float64  `json:"lintDurationSec"`
	Linter          LinterDto `json:"linter"`
	Repo            RepoDto   `json:"repo"`
}

type HighlightSnippetDto struct {
	StartLine int    `json:"startLine"`
	EndLine   int    `json:"endLine"`
	Code      string `json:"code"`
}

type LintHighlightDto struct {
	Linter      LinterDto           `json:"linter"`
	Repo        RepoDto             `json:"repo"`
	Status      string              `json:"status"`
	Path        string              `json:"path"`
	StartLine   int                 `json:"startLine"`
	EndLine     int                 `json:"endLine"`
	Explanation string              `json:"explanation"`
	Snippet     HighlightSnippetDto `json:"snippet"`
}

type LintHighlightsFilter struct {
	LintId   string
	LinterId string
	RepoId   string
}

type BugHuntStorage interface {
	Linters(ctx context.Context) ([]LinterDto, error)
	Repos(ctx context.Context) ([]RepoDto, error)
	LintTasks(ctx context.Context, skip int, take int) ([]LintTaskDto, error)
	LintHighlights(ctx context.Context, filter LintHighlightsFilter) ([]LintHighlightDto, error)
	ModerateHighlight(ctx context.Context, lintId string, highlight dto.LintHighlight, status string) error
}

type PgBugHuntStorage PgStorage

var _ BugHuntStorage = PgBugHuntStorage{}

//go:embed queries/bug_hunt_lint_highlights.sql
var bugHuntLintHighlightsSql string

//go:embed queries/bug_hunt_lint_tasks.sql
var bugHuntLintTasksSql string

//go:embed queries/bug_hunt_linters.sql
var bugHuntLintersSql string

//go:embed queries/bug_hunt_repos.sql
var bugHuntReposSql string

//go:embed queries/bug_hunt_lint_highlight_moderate.sql
var bugHuntLintHighlightModerateSql string

func (b PgBugHuntStorage) Linters(ctx context.Context) ([]LinterDto, error) {
	rows, err := b.Query(ctx, bugHuntLintersSql)
	if err != nil {
		return nil, err
	}
	linters := make([]LinterDto, 0)
	for rows.Next() {
		var (
			linterId                string
			linterGitUrl            string
			linterGitBranch         string
			linterLastDockerImage   *string
			linterLastDockerShaHash *string
			totalHighlightDedup     int
			pendingHighlightDedup   int
			rejectedHighlightDedup  int
			acceptedHighlightDedup  int
		)
		err = rows.Scan(
			&linterId,
			&linterGitUrl,
			&linterGitBranch,
			&linterLastDockerImage,
			&linterLastDockerShaHash,
			&totalHighlightDedup,
			&pendingHighlightDedup,
			&rejectedHighlightDedup,
			&acceptedHighlightDedup,
		)
		if err != nil {
			return nil, err
		}
		linters = append(linters, LinterDto{
			Id:                 linterId,
			GitUrl:             linterGitUrl,
			GitBranch:          linterGitBranch,
			DockerImage:        linterLastDockerImage,
			DockerImageShaHash: linterLastDockerShaHash,
			StatDto: &StatDto{
				TotalHighlightDedup:    totalHighlightDedup,
				PendingHighlightDedup:  pendingHighlightDedup,
				RejectedHighlightDedup: rejectedHighlightDedup,
				AcceptedHighlightDedup: acceptedHighlightDedup,
			},
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return linters, nil
}

func (b PgBugHuntStorage) Repos(ctx context.Context) ([]RepoDto, error) {
	rows, err := b.Query(ctx, bugHuntReposSql)
	if err != nil {
		return nil, err
	}
	repos := make([]RepoDto, 0)
	for rows.Next() {
		var (
			repoId                 string
			repoGitUrl             string
			repoGitBranch          string
			repoLastGitCommitHash  *string
			totalHighlightDedup    int
			pendingHighlightDedup  int
			rejectedHighlightDedup int
			acceptedHighlightDedup int
		)
		err = rows.Scan(
			&repoId,
			&repoGitUrl,
			&repoGitBranch,
			&repoLastGitCommitHash,
			&totalHighlightDedup,
			&pendingHighlightDedup,
			&rejectedHighlightDedup,
			&acceptedHighlightDedup,
		)
		if err != nil {
			return nil, err
		}
		repos = append(repos, RepoDto{
			Id:            repoId,
			GitUrl:        repoGitUrl,
			GitBranch:     repoGitBranch,
			GitCommitHash: repoLastGitCommitHash,
			StatDto: &StatDto{
				TotalHighlightDedup:    totalHighlightDedup,
				PendingHighlightDedup:  pendingHighlightDedup,
				RejectedHighlightDedup: rejectedHighlightDedup,
				AcceptedHighlightDedup: acceptedHighlightDedup,
			},
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return repos, nil
}

func (b PgBugHuntStorage) LintTasks(ctx context.Context, skip int, take int) ([]LintTaskDto, error) {
	rows, err := b.Query(ctx, bugHuntLintTasksSql, skip, take)
	if err != nil {
		return nil, err
	}
	lintTasks := make([]LintTaskDto, 0)
	for rows.Next() {
		var (
			repoId              string
			repoGitUrl          string
			repoGitBranch       string
			repoGitCommitHash   string
			linterId            string
			linterGitUrl        string
			linterGitBranch     string
			linterDockerImage   string
			linterDockerShaHash string
			lintId              string
			lintStatus          string
			lintStatusComment   *string
			lintDuration        *time.Duration
		)
		err = rows.Scan(
			&repoId,
			&repoGitUrl,
			&repoGitBranch,
			&repoGitCommitHash,
			&linterId,
			&linterGitUrl,
			&linterGitBranch,
			&linterDockerImage,
			&linterDockerShaHash,
			&lintId,
			&lintStatus,
			&lintStatusComment,
			&lintDuration,
		)
		if err != nil {
			return nil, err
		}
		var durationSec *float64
		if lintDuration != nil {
			duration := lintDuration.Seconds()
			durationSec = &duration
		}
		lintTasks = append(lintTasks, LintTaskDto{
			Id:              lintId,
			Status:          lintStatus,
			StatusComment:   lintStatusComment,
			LintDurationSec: durationSec,
			Linter: LinterDto{
				Id:                 linterId,
				GitUrl:             linterGitUrl,
				GitBranch:          linterGitBranch,
				DockerImage:        &linterDockerImage,
				DockerImageShaHash: &linterDockerShaHash,
			},
			Repo: RepoDto{
				Id:            repoId,
				GitUrl:        repoGitUrl,
				GitBranch:     repoGitBranch,
				GitCommitHash: &repoGitCommitHash,
			},
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return lintTasks, nil
}

func (b PgBugHuntStorage) LintHighlights(ctx context.Context, filter LintHighlightsFilter) ([]LintHighlightDto, error) {
	rows, err := b.Query(ctx, bugHuntLintHighlightsSql, filter.LintId, filter.LinterId, filter.RepoId)
	if err != nil {
		return nil, err
	}
	highlights := make([]LintHighlightDto, 0)
	for rows.Next() {
		var (
			repoId              string
			repoGitUrl          string
			repoGitBranch       string
			repoGitCommitHash   string
			linterId            string
			linterGitUrl        string
			linterGitBranch     string
			linterDockerImage   string
			linterDockerShaHash string
			lintStatus          string
			lintStatusComment   *string
			lintDuration        *time.Duration
			id                  string
			path                string
			startLine           int
			endLine             int
			explanation         string
			snippetStartLine    int
			snippetEndLine      int
			snippetCode         string
			moderationStatus    string
			moderationComment   *string
			moderatedAt         *time.Time
		)
		err = rows.Scan(
			&repoId,
			&repoGitUrl,
			&repoGitBranch,
			&repoGitCommitHash,
			&linterId,
			&linterGitUrl,
			&linterGitBranch,
			&linterDockerImage,
			&linterDockerShaHash,
			&lintStatus,
			&lintStatusComment,
			&lintDuration,
			&id,
			&path,
			&startLine,
			&endLine,
			&explanation,
			&snippetStartLine,
			&snippetEndLine,
			&snippetCode,
			&moderationStatus,
			&moderationComment,
			&moderatedAt,
		)
		if err != nil {
			return nil, err
		}
		highlights = append(highlights, LintHighlightDto{
			Linter: LinterDto{
				Id:                 linterId,
				GitUrl:             linterGitUrl,
				GitBranch:          linterGitBranch,
				DockerImage:        &linterDockerImage,
				DockerImageShaHash: &linterDockerShaHash,
			},
			Repo: RepoDto{
				Id:            repoId,
				GitUrl:        repoGitUrl,
				GitBranch:     repoGitBranch,
				GitCommitHash: &repoGitCommitHash,
			},
			Status:      moderationStatus,
			Path:        path,
			StartLine:   startLine,
			EndLine:     endLine,
			Explanation: explanation,
			Snippet: HighlightSnippetDto{
				StartLine: snippetStartLine,
				EndLine:   snippetEndLine,
				Code:      snippetCode,
			},
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return highlights, nil
}

func (b PgBugHuntStorage) ModerateHighlight(ctx context.Context, lintId string, highlight dto.LintHighlight, status string) error {
	_, err := b.Exec(
		ctx,
		bugHuntLintHighlightModerateSql,
		lintId,
		highlight.Path,
		highlight.StartLine,
		highlight.EndLine,
		status,
	)
	if err != nil {
		return fmt.Errorf("failed to moderate lint highlight: %w", err)
	}
	return nil
}
