UPDATE lint_tasks
SET lint_status         = $2,
    lint_status_comment = $3,
    lint_duration       = $4,
    linted_at           = $5,
    locked_at           = NULL
WHERE lint_id = $1
