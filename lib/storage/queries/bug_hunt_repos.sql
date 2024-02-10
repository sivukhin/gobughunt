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
     repo_stats_total AS (SELECT repo_id,
                                 COUNT(DISTINCT (repo_id, repo_git_commit_hash, path, start_line, end_line)) as cnt,
                                 COUNT(DISTINCT (repo_id, path, start_line, end_line))                       as cnt_dedup
                          FROM highlights
                          GROUP BY repo_id),
     repo_stats_pending AS (SELECT repo_id,
                                   COUNT(DISTINCT (repo_id, repo_git_commit_hash, path, start_line, end_line)) as cnt,
                                   COUNT(DISTINCT (repo_id, path, start_line, end_line))                       as cnt_dedup
                            FROM highlights
                            WHERE moderation_status = 'pending'
                            GROUP BY repo_id),
     repo_stats_accepted AS (SELECT repo_id,
                                    COUNT(DISTINCT (repo_id, repo_git_commit_hash, path, start_line, end_line)) as cnt,
                                    COUNT(DISTINCT (repo_id, path, start_line, end_line))                       as cnt_dedup
                             FROM highlights
                             WHERE moderation_status = 'accepted'
                             GROUP BY repo_id),
     repo_stats_rejected AS (SELECT repo_id,
                                    COUNT(DISTINCT (repo_id, repo_git_commit_hash, path, start_line, end_line)) as cnt,
                                    COUNT(DISTINCT (repo_id, path, start_line, end_line))                       as cnt_dedup
                             FROM highlights
                             WHERE moderation_status = 'rejected'
                             GROUP BY repo_id)
SELECT repos.repo_id,
       repos.repo_git_url,
       repos.repo_git_branch,
       repos.repo_last_git_commit_hash,
       COALESCE(total.cnt, 0)          as total_highlight,
       COALESCE(total.cnt_dedup, 0)    as total_highlight_dedup,
       COALESCE(pending.cnt, 0)        as pending_highlight,
       COALESCE(pending.cnt_dedup, 0)  as pending_highlight_dedup,
       COALESCE(rejected.cnt, 0)       as rejected_highlight,
       COALESCE(rejected.cnt_dedup, 0) as rejected_highlight_dedup,
       COALESCE(accepted.cnt, 0)       as accepted_highlight,
       COALESCE(accepted.cnt_dedup, 0) as accepted_highlight_dedup
FROM repos as repos
         LEFT JOIN repo_stats_total as total ON repos.repo_id = total.repo_id
         LEFT JOIN repo_stats_pending as pending ON repos.repo_id = pending.repo_id
         LEFT JOIN repo_stats_rejected as rejected ON repos.repo_id = rejected.repo_id
         LEFT JOIN repo_stats_accepted as accepted ON repos.repo_id = accepted.repo_id
ORDER BY updated_at DESC