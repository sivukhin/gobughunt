CREATE TABLE IF NOT EXISTS moderators
(
    moderator_id   text unique not null,
    moderator_name text
);

CREATE TYPE moderation_status as ENUM ('pending', 'rejected', 'accepted');

CREATE TABLE IF NOT EXISTS hunters
(
    hunter_id          text unique       not null,
    language           text              not null,
    docker_image       text              not null,
    created_at         timestamp         not null,
    last_check_hash    text,
    last_check_time    timestamp,

    moderation_status  moderation_status not null default 'pending',
    moderator_id       text,
    moderated_at       timestamp,
    moderation_comment text,

    CONSTRAINT fk_moderator FOREIGN KEY (moderator_id) REFERENCES moderators (moderator_id)
);

CREATE TYPE hunt_status as ENUM ('pending', 'failed', 'succeed', 'skipped');

CREATE TABLE IF NOT EXISTS hunts
(
    hunt_id     text unique not null,
    hunt_dt     timestamp   not null,
    prey_id     text        not null,
    hunter_id   text        not null,
    prey_hash   text        not null,
    hunter_hash text        not null,
    hunt_status hunt_status not null default 'pending',
    created_at  timestamp   not null,
    locked_at   timestamp,

    CONSTRAINT fk_prey FOREIGN KEY (prey_id) REFERENCES preys (prey_id),
    CONSTRAINT fk_hunter FOREIGN KEY (hunter_id) REFERENCES hunters (hunter_id)
);
CREATE TABLE IF NOT EXISTS hunt_details
(
    hunt_id            text              not null,
    file               text              not null,
    function           text              not null,
    line               integer           not null,
    snippet            text              not null,
    issue_ref          text,

    moderation_status  moderation_status not null default 'pending',
    moderator_id       text,
    moderated_at       timestamp,
    moderation_comment text,

    CONSTRAINT fk_hunt FOREIGN KEY (hunt_id) REFERENCES hunts (hunt_id),
    CONSTRAINT fk_moderator FOREIGN KEY (moderator_id) REFERENCES moderators (moderator_id)
);

CREATE TABLE IF NOT EXISTS prey_stats
(
    prey_id text unique not null,
    stat    jsonb,

    CONSTRAINT fk_prey FOREIGN KEY (prey_id) REFERENCES preys (prey_id)
);

CREATE TABLE IF NOT EXISTS hunter_stats
(
    hunter_id text unique not null,
    stat      jsonb,

    CONSTRAINT fk_hunter FOREIGN KEY (hunter_id) REFERENCES hunters (hunter_id)
);

CREATE TABLE IF NOT EXISTS preys
(
    prey_id            text unique       not null,
    language           text              not null,
    git_repo_url       text              not null,
    git_repo_branch    text              not null,
    project_path       text              not null,
    created_at         timestamp         not null,
    last_check_hash    text,

    moderation_status  moderation_status not null default 'pending',
    moderator_id       text,
    moderated_at       timestamp,
    moderation_comment text,

    CONSTRAINT fk_moderator FOREIGN KEY (moderator_id) REFERENCES moderators (moderator_id)
);
