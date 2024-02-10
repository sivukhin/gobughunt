WITH highlights AS (SELECT t.linter_id,
                           t.linter_docker_sha_hash,
                           t.repo_id,
                           t.repo_git_commit_hash,
                           h.path,
                           h.start_line,
                           h.end_line,
                           h.moderation_status
                    FROM lint_highlights as h
                             JOIN lint_tasks as t ON h.lint_id = t.lint_id),
     linter_stats_total AS (SELECT linter_id,
                                   COUNT(DISTINCT (repo_id, repo_git_commit_hash, path, start_line, end_line)) as cnt,
                                   COUNT(DISTINCT (repo_id, path, start_line, end_line))                       as cnt_dedup
                            FROM highlights
                            GROUP BY linter_id),
     linter_stats_pending AS (SELECT linter_id,
                                     COUNT(DISTINCT (repo_id, repo_git_commit_hash, path, start_line, end_line)) as cnt,
                                     COUNT(DISTINCT (repo_id, path, start_line, end_line))                       as cnt_dedup
                              FROM highlights
                              WHERE moderation_status = 'pending'
                              GROUP BY linter_id),
     linter_stats_accepted AS (SELECT linter_id,
                                      COUNT(DISTINCT (repo_id, repo_git_commit_hash, path, start_line, end_line)) as cnt,
                                      COUNT(DISTINCT (repo_id, path, start_line, end_line))                       as cnt_dedup
                               FROM highlights
                               WHERE moderation_status = 'accepted'
                               GROUP BY linter_id),
     linter_stats_rejected AS (SELECT linter_id,
                                      COUNT(DISTINCT (repo_id, repo_git_commit_hash, path, start_line, end_line)) as cnt,
                                      COUNT(DISTINCT (repo_id, path, start_line, end_line))                       as cnt_dedup
                               FROM highlights
                               WHERE moderation_status = 'rejected'
                               GROUP BY linter_id)
SELECT linters.linter_id,
       linters.linter_git_url,
       linters.linter_git_branch,
       linters.linter_last_docker_image,
       linters.linter_last_docker_sha_hash,
       COALESCE(total.cnt, 0)          as total_highlight,
       COALESCE(total.cnt_dedup, 0)    as total_highlight_dedup,
       COALESCE(pending.cnt, 0)        as pending_highlight,
       COALESCE(pending.cnt_dedup, 0)  as pending_highlight_dedup,
       COALESCE(rejected.cnt, 0)       as rejected_highlight,
       COALESCE(rejected.cnt_dedup, 0) as rejected_highlight_dedup,
       COALESCE(accepted.cnt, 0)       as accepted_highlight,
       COALESCE(accepted.cnt_dedup, 0) as accepted_highlight_dedup
FROM linters as linters
         LEFT JOIN linter_stats_total as total ON linters.linter_id = total.linter_id
         LEFT JOIN linter_stats_pending as pending ON linters.linter_id = pending.linter_id
         LEFT JOIN linter_stats_rejected as rejected ON linters.linter_id = rejected.linter_id
         LEFT JOIN linter_stats_accepted as accepted ON linters.linter_id = accepted.linter_id
ORDER BY updated_at DESC