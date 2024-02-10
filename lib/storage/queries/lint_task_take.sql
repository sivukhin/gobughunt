WITH available_tasks AS (SELECT lint_id,
                                linter_id,
                                linter_docker_image,
                                linter_docker_sha_hash,
                                repo_id,
                                repo_git_url,
                                repo_git_commit_hash,
                                locked_at
                         FROM lint_tasks
                         WHERE lint_status = 'pending'
                           AND (locked_at IS NULL OR locked_at <= $1)
                         ORDER BY created_at
                         LIMIT 1 FOR UPDATE)
UPDATE lint_tasks as t
SET locked_at = $2
FROM available_tasks
WHERE t.lint_id = available_tasks.lint_id
RETURNING
    t.lint_id,
    t.linter_id,
    t.linter_docker_image,
    t.linter_docker_sha_hash,
    t.repo_id,
    t.repo_git_url,
    t.repo_git_commit_hash
