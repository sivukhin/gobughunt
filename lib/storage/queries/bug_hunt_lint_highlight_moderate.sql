UPDATE lint_highlights
SET moderation_status = $5
WHERE lint_id = $1
  AND path = $2
  AND start_line = $3
  AND end_line = $4
