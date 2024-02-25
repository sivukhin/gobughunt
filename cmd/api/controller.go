package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"slices"

	"github.com/sivukhin/gobughunt/storage"
	"github.com/sivukhin/gobughunt/storage/db"
)

type ApiController struct {
	Storage         *db.Queries
	ModeratorLogins []string
}

func RenderTemplate(t *template.Template, data any) (string, error) {
	buffer := bytes.NewBuffer(nil)
	err := t.Execute(buffer, data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (c ApiController) LintHighlightModerate(ctx context.Context, lintId string, path string, startLine, endLine int, status string) error {
	user, _ := ctx.Value("user").(string)
	if !slices.Contains(c.ModeratorLogins, user) {
		return fmt.Errorf("access denied")
	}
	arg := db.ModerateBugHuntHighlightParams{
		LintID:           lintId,
		Path:             path,
		StartLine:        int32(startLine),
		EndLine:          int32(endLine),
		ModerationStatus: db.HighlightStatus(status),
	}
	return c.Storage.ModerateBugHuntHighlight(ctx, arg)
}

func (c ApiController) LintHighlights(ctx context.Context, lintId, repoId, linterId string) (LintHighlightsDto, error) {
	user, _ := ctx.Value("user").(string)
	if !slices.Contains(c.ModeratorLogins, user) {
		return LintHighlightsDto{}, fmt.Errorf("access denied")
	}
	args := db.ListBugHuntHighlightsParams{LintID: lintId, RepoID: repoId, LinterID: linterId}
	highlights, err := c.Storage.ListBugHuntHighlights(ctx, args)
	if err != nil {
		return LintHighlightsDto{}, err
	}
	dtoHighlights := make([]LintHighlightDto, 0, len(highlights))
	for _, highlight := range highlights {
		dtoHighlights = append(dtoHighlights, LintHighlightDto{
			LintId: highlight.LintID,
			Linter: LinterDto{
				Id:                 highlight.LinterID,
				GitUrl:             highlight.LinterGitUrl,
				GitBranch:          highlight.LinterGitBranch,
				DockerImage:        &highlight.LinterDockerImage,
				DockerImageShaHash: &highlight.LinterDockerShaHash,
			},
			Repo: RepoDto{
				Id:            highlight.RepoID,
				GitUrl:        highlight.RepoGitUrl,
				GitBranch:     highlight.RepoGitBranch,
				GitCommitHash: &highlight.RepoGitCommitHash,
			},
			Status:      string(highlight.ModerationStatus),
			Path:        highlight.Path,
			StartLine:   int(highlight.StartLine),
			EndLine:     int(highlight.EndLine),
			Explanation: highlight.Explanation,
			Snippet: HighlightSnippetDto{
				StartLine: int(highlight.SnippetStartLine),
				EndLine:   int(highlight.SnippetEndLine),
				Code:      highlight.SnippetCode,
			},
		})
	}
	return LintHighlightsDto{Login: user, LintId: lintId, RepoId: repoId, LinterId: linterId, Highlights: dtoHighlights}, nil
}

type LintTasksDto struct {
	Login string
	Tasks []LintTaskDto
}

func (c ApiController) LintTasks(ctx context.Context, skip, take int) (LintTasksDto, error) {
	user, _ := ctx.Value("user").(string)
	if !slices.Contains(c.ModeratorLogins, user) {
		return LintTasksDto{}, fmt.Errorf("access denied")
	}
	args := db.ListBugHuntLintTasksParams{Offset: int32(skip), Limit: int32(take)}
	tasks, err := c.Storage.ListBugHuntLintTasks(ctx, args)
	if err != nil {
		return LintTasksDto{}, err
	}
	dtoTasks := make([]LintTaskDto, 0, len(tasks))
	for _, task := range tasks {
		dtoTasks = append(dtoTasks, LintTaskDto{
			Id:              task.LintID,
			Status:          string(task.LintStatus),
			StatusComment:   storage.TryGetText(task.LintStatusComment),
			LintDurationSec: storage.TryGetDurationSec(task.LintDuration),
			Linter: LinterDto{
				Id:                 task.LinterID,
				GitUrl:             task.LinterGitUrl,
				GitBranch:          task.LinterGitBranch,
				DockerImage:        &task.LinterDockerImage,
				DockerImageShaHash: &task.LinterDockerShaHash,
			},
			Repo: RepoDto{
				Id:            task.RepoID,
				GitUrl:        task.RepoGitUrl,
				GitBranch:     task.RepoGitBranch,
				GitCommitHash: &task.RepoGitCommitHash,
			},
		})
	}
	return LintTasksDto{Login: user, Tasks: dtoTasks}, nil
}

func (c ApiController) Dashboard(ctx context.Context) (DashboardDto, error) {
	linters, err := c.Storage.ListBugHuntLinters(ctx)
	if err != nil {
		return DashboardDto{}, err
	}
	dtoLinters := make([]LinterDto, 0, len(linters))
	for _, linter := range linters {
		dtoLinters = append(dtoLinters, LinterDto{
			Id:                 linter.LinterID,
			GitUrl:             linter.LinterGitUrl,
			GitBranch:          linter.LinterGitBranch,
			DockerImage:        storage.TryGetText(linter.LinterLastDockerImage),
			DockerImageShaHash: storage.TryGetText(linter.LinterLastDockerShaHash),
			StatDto: &StatDto{
				TotalHighlight:    int(linter.TotalHighlight),
				PendingHighlight:  int(linter.PendingHighlight),
				RejectedHighlight: int(linter.RejectedHighlight),
				AcceptedHighlight: int(linter.AcceptedHighlight),
			},
		})
	}
	repos, err := c.Storage.ListBugHuntRepos(ctx)
	if err != nil {
		return DashboardDto{}, err
	}
	dtoRepos := make([]RepoDto, 0, len(repos))
	for _, repo := range repos {
		dtoRepos = append(dtoRepos, RepoDto{
			Id:            repo.RepoID,
			GitUrl:        repo.RepoGitUrl,
			GitBranch:     repo.RepoGitBranch,
			GitCommitHash: storage.TryGetText(repo.RepoLastGitCommitHash),
			StatDto: &StatDto{
				TotalHighlight:    int(repo.TotalHighlight),
				PendingHighlight:  int(repo.PendingHighlight),
				RejectedHighlight: int(repo.RejectedHighlight),
				AcceptedHighlight: int(repo.AcceptedHighlight),
			},
		})
	}
	user, _ := ctx.Value("user").(string)
	return DashboardDto{
		Login:   user,
		Linters: dtoLinters,
		Repos:   dtoRepos,
	}, nil
}
