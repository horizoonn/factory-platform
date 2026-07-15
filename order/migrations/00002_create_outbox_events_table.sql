-- +goose Up
-- +goose StatementBegin

CREATE TABLE platform.order_outbox_events (
    id              UUID        PRIMARY KEY,
    aggregate_id    UUID        NOT NULL,
    event_type      TEXT        NOT NULL CHECK (BTRIM(event_type) <> ''),
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

    UNIQUE(aggregate_id, event_type),
    CHECK ((locked_by IS NULL) = (locked_until IS NULL)),
    CHECK (NOT (published_at IS NOT NULL AND failed_at IS NOT NULL))
);

CREATE INDEX idx_order_outbox_events_pending
    ON platform.order_outbox_events (available_at, next_attempt_at, created_at)
    WHERE published_at IS NULL AND failed_at IS NULL;

CREATE TABLE platform.order_inbox_events (
    event_id     UUID        PRIMARY KEY,
    aggregate_id UUID        NOT NULL,
    event_type   TEXT        NOT NULL CHECK (BTRIM(event_type) <> ''),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_inbox_events_aggregate_id
    ON platform.order_inbox_events (aggregate_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS platform.order_inbox_events;
DROP TABLE IF EXISTS platform.order_outbox_events;
-- +goose StatementEnd
