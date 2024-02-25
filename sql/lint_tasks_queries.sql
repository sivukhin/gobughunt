-- name: AddLintTask :exec
INSERT INTO lint_tasks
(lint_id, lint_status, linter_id, linter_docker_image, linter_docker_sha_hash, repo_id, repo_git_url, repo_git_commit_hash, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: SetLintTask :exec
UPDATE lint_tasks
SET lint_status         = $2,
    lint_status_comment = $3,
    lint_duration       = $4,
    linted_at           = $5,
    locked_at           = NULL
WHERE lint_id = $1;

-- name: TryTakeLintTask :one
WITH available_tasks AS (SELECT lint_id,
    linter_id,
    linter_docker_image,
    linter_docker_sha_hash,
    repo_id,
    repo_git_url,
    repo_git_commit_hash,
    locked_at
    FROM lint_tasks t
    WHERE lint_status = 'pending'
    AND (t.locked_at IS NULL OR t.locked_at <= @lock_time_lower_bound)
    ORDER BY created_at
    LIMIT 1 FOR UPDATE)
UPDATE lint_tasks as t
SET locked_at = @locked_at
FROM available_tasks
WHERE t.lint_id = available_tasks.lint_id
    RETURNING
    t.lint_id,
    t.linter_id,
    t.linter_docker_image,
    t.linter_docker_sha_hash,
    t.repo_id,
    t.repo_git_url,
    t.repo_git_commit_hash;