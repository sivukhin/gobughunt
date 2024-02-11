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
