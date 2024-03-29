// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: bug_hunt_queries.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const listBugHuntHighlights = `-- name: ListBugHuntHighlights :many
WITH highlights AS (SELECT h.repo_id,
                           repos.repo_git_url,
                           repos.repo_git_branch,
                           h.repo_git_commit_hash,

                           h.linter_id,
                           linters.linter_git_url,
                           linters.linter_git_branch,
                           h.linter_docker_image,
                           h.linter_docker_sha_hash,

                           h.lint_status,
                           h.lint_status_comment,
                           h.lint_duration,

                           h.lint_id,
                           h.path,
                           h.start_line,
                           h.end_line,
                           h.explanation,
                           h.snippet_start_line,
                           h.snippet_end_line,
                           h.snippet_code,
                           h.moderation_status,
                           h.moderation_comment,
                           h.moderated_at
                    FROM (SELECT lint_tasks.repo_id,
                                 lint_tasks.repo_git_commit_hash,
                                 lint_tasks.linter_id,
                                 lint_tasks.linter_docker_image,
                                 lint_tasks.linter_docker_sha_hash,
                                 lint_tasks.lint_status,
                                 lint_tasks.lint_status_comment,
                                 lint_tasks.lint_duration,

                                 lint_highlights.lint_id,
                                 lint_highlights.path,
                                 lint_highlights.start_line,
                                 lint_highlights.end_line,
                                 lint_highlights.explanation,
                                 lint_highlights.snippet_start_line,
                                 lint_highlights.snippet_end_line,
                                 lint_highlights.snippet_code,
                                 lint_highlights.moderation_status,
                                 lint_highlights.moderation_comment,
                                 lint_highlights.moderated_at
                          FROM lint_highlights as lint_highlights
                                   JOIN lint_tasks as lint_tasks ON lint_highlights.lint_id = lint_tasks.lint_id) as h
                             JOIN linters as linters ON h.linter_id = linters.linter_id
                             JOIN repos as repos ON h.repo_id = repos.repo_id
                    WHERE ($1 = '' OR h.lint_id = $1)
                      AND ($2 = '' OR h.linter_id = $2)
                      AND ($3 = '' OR h.repo_id = $3))
SELECT repo_id, repo_git_url, repo_git_branch, repo_git_commit_hash, linter_id, linter_git_url, linter_git_branch, linter_docker_image, linter_docker_sha_hash, lint_status, lint_status_comment, lint_duration, lint_id, path, start_line, end_line, explanation, snippet_start_line, snippet_end_line, snippet_code, moderation_status, moderation_comment, moderated_at
FROM highlights as t
WHERE moderation_status = (SELECT MAX(moderation_status)
                           FROM highlights as h
                           WHERE t.repo_id = h.repo_id
                             AND t.path = h.path
                             AND t.start_line = h.start_line
                             AND t.end_line = h.end_line)
ORDER BY (t.moderation_status, t.repo_id, t.path, t.start_line)
`

type ListBugHuntHighlightsParams struct {
	LintID   interface{}
	LinterID interface{}
	RepoID   interface{}
}

type ListBugHuntHighlightsRow struct {
	RepoID              string
	RepoGitUrl          string
	RepoGitBranch       string
	RepoGitCommitHash   string
	LinterID            string
	LinterGitUrl        string
	LinterGitBranch     string
	LinterDockerImage   string
	LinterDockerShaHash string
	LintStatus          LintStatus
	LintStatusComment   pgtype.Text
	LintDuration        pgtype.Interval
	LintID              string
	Path                string
	StartLine           int32
	EndLine             int32
	Explanation         string
	SnippetStartLine    int32
	SnippetEndLine      int32
	SnippetCode         string
	ModerationStatus    HighlightStatus
	ModerationComment   pgtype.Text
	ModeratedAt         pgtype.Timestamp
}

func (q *Queries) ListBugHuntHighlights(ctx context.Context, arg ListBugHuntHighlightsParams) ([]ListBugHuntHighlightsRow, error) {
	rows, err := q.db.Query(ctx, listBugHuntHighlights, arg.LintID, arg.LinterID, arg.RepoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListBugHuntHighlightsRow
	for rows.Next() {
		var i ListBugHuntHighlightsRow
		if err := rows.Scan(
			&i.RepoID,
			&i.RepoGitUrl,
			&i.RepoGitBranch,
			&i.RepoGitCommitHash,
			&i.LinterID,
			&i.LinterGitUrl,
			&i.LinterGitBranch,
			&i.LinterDockerImage,
			&i.LinterDockerShaHash,
			&i.LintStatus,
			&i.LintStatusComment,
			&i.LintDuration,
			&i.LintID,
			&i.Path,
			&i.StartLine,
			&i.EndLine,
			&i.Explanation,
			&i.SnippetStartLine,
			&i.SnippetEndLine,
			&i.SnippetCode,
			&i.ModerationStatus,
			&i.ModerationComment,
			&i.ModeratedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBugHuntLintTasks = `-- name: ListBugHuntLintTasks :many
SELECT repos.repo_id,
       repos.repo_git_url,
       repos.repo_git_branch,
       lint_tasks.repo_git_commit_hash,
       linters.linter_id,
       linters.linter_git_url,
       linters.linter_git_branch,
       lint_tasks.linter_docker_image,
       lint_tasks.linter_docker_sha_hash,
       lint_tasks.lint_id,
       lint_tasks.lint_status,
       lint_tasks.lint_status_comment,
       lint_tasks.lint_duration
FROM lint_tasks as lint_tasks
         JOIN linters as linters ON linters.linter_id = lint_tasks.linter_id
         JOIN repos as repos ON repos.repo_id = lint_tasks.repo_id
ORDER BY lint_tasks.lint_status,
         lint_tasks.created_at DESC
LIMIT $2 OFFSET $1
`

type ListBugHuntLintTasksParams struct {
	Offset int32
	Limit  int32
}

type ListBugHuntLintTasksRow struct {
	RepoID              string
	RepoGitUrl          string
	RepoGitBranch       string
	RepoGitCommitHash   string
	LinterID            string
	LinterGitUrl        string
	LinterGitBranch     string
	LinterDockerImage   string
	LinterDockerShaHash string
	LintID              string
	LintStatus          LintStatus
	LintStatusComment   pgtype.Text
	LintDuration        pgtype.Interval
}

func (q *Queries) ListBugHuntLintTasks(ctx context.Context, arg ListBugHuntLintTasksParams) ([]ListBugHuntLintTasksRow, error) {
	rows, err := q.db.Query(ctx, listBugHuntLintTasks, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListBugHuntLintTasksRow
	for rows.Next() {
		var i ListBugHuntLintTasksRow
		if err := rows.Scan(
			&i.RepoID,
			&i.RepoGitUrl,
			&i.RepoGitBranch,
			&i.RepoGitCommitHash,
			&i.LinterID,
			&i.LinterGitUrl,
			&i.LinterGitBranch,
			&i.LinterDockerImage,
			&i.LinterDockerShaHash,
			&i.LintID,
			&i.LintStatus,
			&i.LintStatusComment,
			&i.LintDuration,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBugHuntLinters = `-- name: ListBugHuntLinters :many
WITH highlights AS (SELECT h.linter_id,
                           h.repo_id,
                           h.path,
                           h.start_line,
                           h.end_line,
                           max(h.moderation_status) as moderation_status
                    FROM (SELECT t.lint_id,
                                 t.linter_id,
                                 t.linter_docker_sha_hash,
                                 t.repo_id,
                                 t.repo_git_commit_hash,
                                 h.path,
                                 h.start_line,
                                 h.end_line,
                                 h.moderation_status
                          FROM lint_highlights as h
                                   JOIN lint_tasks as t ON h.lint_id = t.lint_id) h
                    GROUP BY h.linter_id,
                             h.repo_id,
                             h.path,
                             h.start_line,
                             h.end_line),
     linter_stats_total AS (SELECT linter_id,
                                   COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                            FROM highlights
                            GROUP BY linter_id),
     linter_stats_pending AS (SELECT linter_id,
                                     COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                              FROM highlights
                              WHERE moderation_status = 'pending'
                              GROUP BY linter_id),
     linter_stats_accepted AS (SELECT linter_id,
                                      COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                               FROM highlights
                               WHERE moderation_status = 'accepted'
                               GROUP BY linter_id),
     linter_stats_rejected AS (SELECT linter_id,
                                      COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                               FROM highlights
                               WHERE moderation_status = 'rejected'
                               GROUP BY linter_id)
SELECT linters.linter_id,
       linters.linter_git_url,
       linters.linter_git_branch,
       linters.linter_last_docker_image,
       linters.linter_last_docker_sha_hash,
       COALESCE(total.cnt, 0)    as total_highlight,
       COALESCE(pending.cnt, 0)  as pending_highlight,
       COALESCE(rejected.cnt, 0) as rejected_highlight,
       COALESCE(accepted.cnt, 0) as accepted_highlight
FROM linters as linters
         LEFT JOIN linter_stats_total as total ON linters.linter_id = total.linter_id
         LEFT JOIN linter_stats_pending as pending ON linters.linter_id = pending.linter_id
         LEFT JOIN linter_stats_rejected as rejected ON linters.linter_id = rejected.linter_id
         LEFT JOIN linter_stats_accepted as accepted ON linters.linter_id = accepted.linter_id
ORDER BY accepted_highlight DESC, pending_highlight DESC, rejected_highlight, updated_at DESC
`

type ListBugHuntLintersRow struct {
	LinterID                string
	LinterGitUrl            string
	LinterGitBranch         string
	LinterLastDockerImage   pgtype.Text
	LinterLastDockerShaHash pgtype.Text
	TotalHighlight          int64
	PendingHighlight        int64
	RejectedHighlight       int64
	AcceptedHighlight       int64
}

func (q *Queries) ListBugHuntLinters(ctx context.Context) ([]ListBugHuntLintersRow, error) {
	rows, err := q.db.Query(ctx, listBugHuntLinters)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListBugHuntLintersRow
	for rows.Next() {
		var i ListBugHuntLintersRow
		if err := rows.Scan(
			&i.LinterID,
			&i.LinterGitUrl,
			&i.LinterGitBranch,
			&i.LinterLastDockerImage,
			&i.LinterLastDockerShaHash,
			&i.TotalHighlight,
			&i.PendingHighlight,
			&i.RejectedHighlight,
			&i.AcceptedHighlight,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listBugHuntRepos = `-- name: ListBugHuntRepos :many
WITH 
    alive_highlights AS (SELECT t.lint_id,
                                t.linter_id,
                                t.linter_docker_sha_hash,
                                t.repo_id,
                                t.repo_git_commit_hash,
                                h.path,
                                h.start_line,
                                h.end_line,
                                h.moderation_status
                          FROM lint_highlights as h
                          JOIN lint_tasks as t ON h.lint_id = t.lint_id
                          JOIN linters as l ON t.linter_id = l.linter_id
                          JOIN repos as r ON t.repo_id = r.repo_id),
    highlights AS (SELECT h.linter_id,
                           h.repo_id,
                           h.path,
                           h.start_line,
                           h.end_line,
                           max(h.moderation_status) as moderation_status
                    FROM alive_highlights h
                    GROUP BY h.linter_id,
                             h.repo_id,
                             h.path,
                             h.start_line,
                             h.end_line),
     repo_stats_total AS (SELECT repo_id,
                                 COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                          FROM highlights
                          GROUP BY repo_id),
     repo_stats_pending AS (SELECT repo_id,
                                   COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                            FROM highlights
                            WHERE moderation_status = 'pending'
                            GROUP BY repo_id),
     repo_stats_accepted AS (SELECT repo_id,
                                    COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                             FROM highlights
                             WHERE moderation_status = 'accepted'
                             GROUP BY repo_id),
     repo_stats_rejected AS (SELECT repo_id,
                                    COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                             FROM highlights
                             WHERE moderation_status = 'rejected'
                             GROUP BY repo_id)
SELECT repos.repo_id,
       repos.repo_git_url,
       repos.repo_git_branch,
       repos.repo_last_git_commit_hash,
       COALESCE(total.cnt, 0)    as total_highlight,
       COALESCE(pending.cnt, 0)  as pending_highlight,
       COALESCE(rejected.cnt, 0) as rejected_highlight,
       COALESCE(accepted.cnt, 0) as accepted_highlight
FROM repos as repos
         LEFT JOIN repo_stats_total as total ON repos.repo_id = total.repo_id
         LEFT JOIN repo_stats_pending as pending ON repos.repo_id = pending.repo_id
         LEFT JOIN repo_stats_rejected as rejected ON repos.repo_id = rejected.repo_id
         LEFT JOIN repo_stats_accepted as accepted ON repos.repo_id = accepted.repo_id
ORDER BY accepted_highlight DESC, pending_highlight DESC, rejected_highlight, updated_at DESC
`

type ListBugHuntReposRow struct {
	RepoID                string
	RepoGitUrl            string
	RepoGitBranch         string
	RepoLastGitCommitHash pgtype.Text
	TotalHighlight        int64
	PendingHighlight      int64
	RejectedHighlight     int64
	AcceptedHighlight     int64
}

func (q *Queries) ListBugHuntRepos(ctx context.Context) ([]ListBugHuntReposRow, error) {
	rows, err := q.db.Query(ctx, listBugHuntRepos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListBugHuntReposRow
	for rows.Next() {
		var i ListBugHuntReposRow
		if err := rows.Scan(
			&i.RepoID,
			&i.RepoGitUrl,
			&i.RepoGitBranch,
			&i.RepoLastGitCommitHash,
			&i.TotalHighlight,
			&i.PendingHighlight,
			&i.RejectedHighlight,
			&i.AcceptedHighlight,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const moderateBugHuntHighlight = `-- name: ModerateBugHuntHighlight :exec
UPDATE lint_highlights
SET moderation_status = $5
WHERE lint_id = $1
  AND path = $2
  AND start_line = $3
  AND end_line = $4
`

type ModerateBugHuntHighlightParams struct {
	LintID           string
	Path             string
	StartLine        int32
	EndLine          int32
	ModerationStatus HighlightStatus
}

func (q *Queries) ModerateBugHuntHighlight(ctx context.Context, arg ModerateBugHuntHighlightParams) error {
	_, err := q.db.Exec(ctx, moderateBugHuntHighlight,
		arg.LintID,
		arg.Path,
		arg.StartLine,
		arg.EndLine,
		arg.ModerationStatus,
	)
	return err
}
