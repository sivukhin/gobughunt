CREATE TYPE LintStatus AS ENUM ('pending', 'locked', 'succeed', 'failed', 'skipped');
CREATE TABLE IF NOT EXISTS lint_tasks
(
    lint_id                TEXT       NOT NULL UNIQUE,
    linter_id              TEXT       NOT NULL,
    linter_docker_image    TEXT       NOT NULL,
    linter_docker_sha_hash TEXT       NOT NULL,
    repo_id                TEXT       NOT NULL,
    repo_git_url           TEXT       NOT NULL,
    repo_git_commit_hash   TEXT       NOT NULL,

    lint_status            LintStatus NOT NULL DEFAULT 'pending',
    lint_status_comment    TEXT,
    lint_duration          INTERVAL,
    created_at             TIMESTAMP  NOT NULL,
    locked_at              TIMESTAMP,
    linted_at              TIMESTAMP
);
CREATE UNIQUE INDEX hash_unique ON lint_tasks
    (linter_docker_image, linter_docker_sha_hash, repo_git_url, repo_git_commit_hash);