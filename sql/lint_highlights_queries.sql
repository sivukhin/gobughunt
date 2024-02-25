-- name: AddLintHighlight :batchexec
INSERT INTO lint_highlights (lint_id, path, start_line, end_line, explanation, snippet_start_line, snippet_end_line, snippet_code)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);