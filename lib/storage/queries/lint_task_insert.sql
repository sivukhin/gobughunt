INSERT INTO lint_tasks
    (lint_id, lint_status, linter_id, linter_docker_image, linter_docker_sha_hash, repo_id, repo_git_url, repo_git_commit_hash, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)