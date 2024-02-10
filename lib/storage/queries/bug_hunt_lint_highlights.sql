SELECT h.repo_id,
       repos.repo_git_url,
       repos.repo_git_branch,
       h.repo_git_commit_hash,

       h.linter_id,
       linters.linter_git_url,
       linters.linter_git_branch,
       h.linter_docker_image,
       h.linter_docker_sha_hash,

       h.lint_status,
       h.lint_status_comment,
       h.lint_duration,

       h.path,
       h.start_line,
       h.end_line,
       h.explanation,
       h.snippet_start_line,
       h.snippet_end_line,
       h.snippet_code,
       h.moderation_status,
       h.moderation_comment,
       h.moderated_at
FROM (SELECT lint_tasks.repo_id,
             lint_tasks.repo_git_commit_hash,
             lint_tasks.linter_id,
             lint_tasks.linter_docker_image,
             lint_tasks.linter_docker_sha_hash,
             lint_tasks.lint_status,
             lint_tasks.lint_status_comment,
             lint_tasks.lint_duration,

             lint_highlights.path,
             lint_highlights.start_line,
             lint_highlights.end_line,
             lint_highlights.explanation,
             lint_highlights.snippet_start_line,
             lint_highlights.snippet_end_line,
             lint_highlights.snippet_code,
             lint_highlights.moderation_status,
             lint_highlights.moderation_comment,
             lint_highlights.moderated_at
      FROM lint_highlights as lint_highlights
               JOIN lint_tasks as lint_tasks ON lint_highlights.lint_id = lint_tasks.lint_id
      WHERE lint_highlights.lint_id = $1) as h
         JOIN linters as linters ON h.linter_id = linters.linter_id
         JOIN repos as repos ON h.repo_id = repos.repo_id