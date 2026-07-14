-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS platform;

CREATE TABLE platform.assembly_outbox_events (
    id              UUID        PRIMARY KEY,
    source_event_id UUID        NOT NULL UNIQUE,
    aggregate_id    UUID        NOT NULL,
    topic           TEXT        NOT NULL CHECK (BTRIM(topic) <> ''),
    message_key     BYTEA       NOT NULL CHECK (OCTET_LENGTH(message_key) > 0),
    payload         BYTEA       NOT NULL CHECK (OCTET_LENGTH(payload) > 0),
    headers         JSONB       NOT NULL DEFAULT '{}'::JSONB CHECK (JSONB_TYPEOF(headers) = 'object'),
    available_at    TIMESTAMPTZ NOT NULL,
    next_attempt_at TIMESTAMPTZ NOT NULL,
    attempts        INTEGER     NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    locked_by       TEXT,
    locked_until    TIMESTAMPTZ,
    published_at    TIMESTAMPTZ,
    failed_at       TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CHECK ((locked_by IS NULL) = (locked_until IS NULL)),
    CHECK (NOT (published_at IS NOT NULL AND failed_at IS NOT NULL))
);

CREATE INDEX idx_assembly_outbox_events_pending
    ON platform.assembly_outbox_events (available_at, next_attempt_at, created_at)
    WHERE published_at IS NULL AND failed_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS platform.assembly_outbox_events;
DROP SCHEMA IF EXISTS platform;
-- +goose StatementEnd
