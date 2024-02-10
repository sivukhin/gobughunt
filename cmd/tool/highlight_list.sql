SELECT *
FROM (SELECT t.created_at, h.*
      FROM (
               lint_highlights h INNER JOIN lint_tasks t ON h.lint_id = t.lint_id
               ))
ORDER BY created_at