SELECT linter_id,
       linter_git_url,
       linter_git_branch,
       linter_last_docker_image,
       linter_last_docker_sha_hash
FROM linters
ORDER BY updated_at DESC