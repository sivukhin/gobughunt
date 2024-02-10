INSERT INTO repos (repo_id, repo_git_url, repo_git_branch, repo_last_git_commit_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $5)
ON CONFLICT (repo_id)
    DO UPDATE SET repo_git_url              = $2,
                  repo_git_branch           = $3,
                  repo_last_git_commit_hash = $4,
                  updated_at                = $5