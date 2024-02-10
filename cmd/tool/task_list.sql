SELECT *
FROM lint_tasks
WHERE lint_status = 'pending'
ORDER BY created_at;