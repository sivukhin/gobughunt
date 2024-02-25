-- name: ListBugHuntLinters :many
WITH highlights AS (SELECT h.linter_id,
                           h.repo_id,
                           h.path,
                           h.start_line,
                           h.end_line,
                           max(h.moderation_status) as moderation_status
                    FROM (SELECT t.lint_id,
                                 t.linter_id,
                                 t.linter_docker_sha_hash,
                                 t.repo_id,
                                 t.repo_git_commit_hash,
                                 h.path,
                                 h.start_line,
                                 h.end_line,
                                 h.moderation_status
                          FROM lint_highlights as h
                                   JOIN lint_tasks as t ON h.lint_id = t.lint_id) h
                    GROUP BY h.linter_id,
                             h.repo_id,
                             h.path,
                             h.start_line,
                             h.end_line),
     linter_stats_total AS (SELECT linter_id,
                                   COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                            FROM highlights
                            GROUP BY linter_id),
     linter_stats_pending AS (SELECT linter_id,
                                     COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                              FROM highlights
                              WHERE moderation_status = 'pending'
                              GROUP BY linter_id),
     linter_stats_accepted AS (SELECT linter_id,
                                      COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                               FROM highlights
                               WHERE moderation_status = 'accepted'
                               GROUP BY linter_id),
     linter_stats_rejected AS (SELECT linter_id,
                                      COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                               FROM highlights
                               WHERE moderation_status = 'rejected'
                               GROUP BY linter_id)
SELECT linters.linter_id,
       linters.linter_git_url,
       linters.linter_git_branch,
       linters.linter_last_docker_image,
       linters.linter_last_docker_sha_hash,
       COALESCE(total.cnt, 0)    as total_highlight,
       COALESCE(pending.cnt, 0)  as pending_highlight,
       COALESCE(rejected.cnt, 0) as rejected_highlight,
       COALESCE(accepted.cnt, 0) as accepted_highlight
FROM linters as linters
         LEFT JOIN linter_stats_total as total ON linters.linter_id = total.linter_id
         LEFT JOIN linter_stats_pending as pending ON linters.linter_id = pending.linter_id
         LEFT JOIN linter_stats_rejected as rejected ON linters.linter_id = rejected.linter_id
         LEFT JOIN linter_stats_accepted as accepted ON linters.linter_id = accepted.linter_id
ORDER BY accepted_highlight DESC, pending_highlight DESC, rejected_highlight, updated_at DESC;

-- name: ListBugHuntRepos :many
WITH highlights AS (SELECT h.linter_id,
                           h.repo_id,
                           h.path,
                           h.start_line,
                           h.end_line,
                           max(h.moderation_status) as moderation_status
                    FROM (SELECT t.lint_id,
                                 t.linter_id,
                                 t.linter_docker_sha_hash,
                                 t.repo_id,
                                 t.repo_git_commit_hash,
                                 h.path,
                                 h.start_line,
                                 h.end_line,
                                 h.moderation_status
                          FROM lint_highlights as h
                                   JOIN lint_tasks as t ON h.lint_id = t.lint_id) h
                    GROUP BY h.linter_id,
                             h.repo_id,
                             h.path,
                             h.start_line,
                             h.end_line),
     repo_stats_total AS (SELECT repo_id,
                                 COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                          FROM highlights
                          GROUP BY repo_id),
     repo_stats_pending AS (SELECT repo_id,
                                   COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                            FROM highlights
                            WHERE moderation_status = 'pending'
                            GROUP BY repo_id),
     repo_stats_accepted AS (SELECT repo_id,
                                    COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                             FROM highlights
                             WHERE moderation_status = 'accepted'
                             GROUP BY repo_id),
     repo_stats_rejected AS (SELECT repo_id,
                                    COUNT(DISTINCT (repo_id, path, start_line, end_line)) as cnt
                             FROM highlights
                             WHERE moderation_status = 'rejected'
                             GROUP BY repo_id)
SELECT repos.repo_id,
       repos.repo_git_url,
       repos.repo_git_branch,
       repos.repo_last_git_commit_hash,
       COALESCE(total.cnt, 0)    as total_highlight,
       COALESCE(pending.cnt, 0)  as pending_highlight,
       COALESCE(rejected.cnt, 0) as rejected_highlight,
       COALESCE(accepted.cnt, 0) as accepted_highlight
FROM repos as repos
         LEFT JOIN repo_stats_total as total ON repos.repo_id = total.repo_id
         LEFT JOIN repo_stats_pending as pending ON repos.repo_id = pending.repo_id
         LEFT JOIN repo_stats_rejected as rejected ON repos.repo_id = rejected.repo_id
         LEFT JOIN repo_stats_accepted as accepted ON repos.repo_id = accepted.repo_id
ORDER BY accepted_highlight DESC, pending_highlight DESC, rejected_highlight, updated_at DESC;

-- name: ListBugHuntLintTasks :many
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
LIMIT $2 OFFSET $1;

-- name: ListBugHuntHighlights :many
WITH highlights AS (SELECT h.repo_id,
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

                           h.lint_id,
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

                                 lint_highlights.lint_id,
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
                                   JOIN lint_tasks as lint_tasks ON lint_highlights.lint_id = lint_tasks.lint_id) as h
                             JOIN linters as linters ON h.linter_id = linters.linter_id
                             JOIN repos as repos ON h.repo_id = repos.repo_id
                    WHERE (@lint_id = '' OR h.lint_id = @lint_id)
                      AND (@linter_id = '' OR h.linter_id = @linter_id)
                      AND (@repo_id = '' OR h.repo_id = @repo_id))
SELECT *
FROM highlights as t
WHERE moderation_status = (SELECT MAX(moderation_status)
                           FROM highlights as h
                           WHERE t.repo_id = h.repo_id
                             AND t.path = h.path
                             AND t.start_line = h.start_line
                             AND t.end_line = h.end_line)
ORDER BY (t.moderation_status, t.repo_id, t.path, t.start_line);

-- name: ModerateBugHuntHighlight :exec
UPDATE lint_highlights
SET moderation_status = $5
WHERE lint_id = $1
  AND path = $2
  AND start_line = $3
  AND end_line = $4;
