INSERT INTO repos (repo_id, repo_git_url, repo_git_branch, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())