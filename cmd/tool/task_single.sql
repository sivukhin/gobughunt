SELECT lint_id,
       linter_id,
       linter_docker_image,
       linter_docker_sha_hash,
       repo_id,
       repo_git_url,
       repo_git_commit_hash,
       locked_at
FROM lint_tasks
WHERE lint_status = 'pending'
  AND (locked_at IS NULL)
ORDER BY created_at
LIMIT 1