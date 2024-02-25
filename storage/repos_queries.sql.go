// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: repos_queries.sql

package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const deleteRepo = `-- name: DeleteRepo :exec
DELETE FROM repos WHERE repo_id = $1
`

func (q *Queries) DeleteRepo(ctx context.Context, repoID string) error {
	_, err := q.db.Exec(ctx, deleteRepo, repoID)
	return err
}

const getRepo = `-- name: GetRepo :one
SELECT repo_id,
       repo_git_url,
       repo_git_branch,
       repo_last_git_commit_hash
FROM repos
WHERE repo_id = $1
`

type GetRepoRow struct {
	RepoID                string
	RepoGitUrl            string
	RepoGitBranch         string
	RepoLastGitCommitHash pgtype.Text
}

func (q *Queries) GetRepo(ctx context.Context, repoID string) (GetRepoRow, error) {
	row := q.db.QueryRow(ctx, getRepo, repoID)
	var i GetRepoRow
	err := row.Scan(
		&i.RepoID,
		&i.RepoGitUrl,
		&i.RepoGitBranch,
		&i.RepoLastGitCommitHash,
	)
	return i, err
}

const listRepos = `-- name: ListRepos :many
SELECT repo_id,
       repo_git_url,
       repo_git_branch,
       repo_last_git_commit_hash
FROM repos
ORDER BY updated_at DESC
`

type ListReposRow struct {
	RepoID                string
	RepoGitUrl            string
	RepoGitBranch         string
	RepoLastGitCommitHash pgtype.Text
}

func (q *Queries) ListRepos(ctx context.Context) ([]ListReposRow, error) {
	rows, err := q.db.Query(ctx, listRepos)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListReposRow
	for rows.Next() {
		var i ListReposRow
		if err := rows.Scan(
			&i.RepoID,
			&i.RepoGitUrl,
			&i.RepoGitBranch,
			&i.RepoLastGitCommitHash,
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

const upsertRepo = `-- name: UpsertRepo :exec
INSERT INTO repos (repo_id, repo_git_url, repo_git_branch, repo_last_git_commit_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $5)
ON CONFLICT (repo_id)
    DO UPDATE SET repo_git_url              = $2,
    repo_git_branch           = $3,
    repo_last_git_commit_hash = $4,
    updated_at                = $5
`

type UpsertRepoParams struct {
	RepoID                string
	RepoGitUrl            string
	RepoGitBranch         string
	RepoLastGitCommitHash pgtype.Text
	CreatedAt             pgtype.Timestamp
}

func (q *Queries) UpsertRepo(ctx context.Context, arg UpsertRepoParams) error {
	_, err := q.db.Exec(ctx, upsertRepo,
		arg.RepoID,
		arg.RepoGitUrl,
		arg.RepoGitBranch,
		arg.RepoLastGitCommitHash,
		arg.CreatedAt,
	)
	return err
}
