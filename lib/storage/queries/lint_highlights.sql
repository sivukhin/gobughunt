CREATE TYPE HighlightStatus AS ENUM ('pending', 'accepted', 'rejected');

--- DELIMITER ---

CREATE TABLE IF NOT EXISTS lint_highlights
(
    lint_id            TEXT            NOT NULL,
    path               TEXT            NOT NULL,
    start_line         INT             NOT NULL,
    end_line           INT             NOT NULL,
    explanation        TEXT            NOT NULL,
    snippet_start_line INT             NOT NULL,
    snippet_end_line   INT             NOT NULL,
    snippet_code       TEXT            NOT NULL,

    moderation_status  HighlightStatus NOT NULL DEFAULT 'pending',
    moderation_comment TEXT,
    moderated_at       TIMESTAMP,

    CONSTRAINT fk_lint_id FOREIGN KEY (lint_id) REFERENCES lint_tasks (lint_id)
);
