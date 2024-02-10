INSERT INTO linters (linter_id, linter_git_url, linter_git_branch, linter_last_docker_image, linter_last_docker_sha_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (linter_id)
    DO UPDATE SET linter_git_url              = $2,
                  linter_git_branch           = $3,
                  linter_last_docker_image    = $4,
                  linter_last_docker_sha_hash = $5,
                  updated_at                  = $6