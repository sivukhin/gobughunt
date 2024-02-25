CREATE TABLE IF NOT EXISTS repos
(
    repo_id                   TEXT UNIQUE NOT NULL,
    repo_git_url              TEXT        NOT NULL,
    repo_git_branch           TEXT        NOT NULL,
    repo_last_git_commit_hash TEXT,
    created_at                TIMESTAMP   NOT NULL,
    updated_at                TIMESTAMP   NOT NULL
);
