UPDATE lint_tasks
SET lint_status = 'pending',
    locked_at   = NULL
WHERE lint_id = $1;

--- DELIMITER ---

DELETE
FROM lint_highlights
WHERE lint_id = $1;