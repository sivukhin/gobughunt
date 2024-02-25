CREATE TABLE IF NOT EXISTS linters
(
    linter_id                   TEXT UNIQUE NOT NULL,
    linter_git_url              TEXT        NOT NULL,
    linter_git_branch           TEXT        NOT NULL,
    linter_last_docker_image    TEXT,
    linter_last_docker_sha_hash TEXT,
    created_at                  TIMESTAMP   NOT NULL,
    updated_at                  TIMESTAMP   NOT NULL
);