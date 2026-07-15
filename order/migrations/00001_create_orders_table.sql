-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS platform;

CREATE TABLE platform.orders (
    id             UUID          PRIMARY KEY,
    user_id        UUID          NOT NULL,
    part_ids       UUID[]        NOT NULL CHECK (cardinality(part_ids) > 0),
    total_price    NUMERIC(12,2) NOT NULL CHECK (total_price >= 0),
    transaction_id UUID,
    payment_method TEXT                   CHECK (payment_method IS NULL OR payment_method IN ('CARD', 'SBP', 'CREDIT_CARD', 'INVESTOR_MONEY')),
    status         TEXT          NOT NULL CHECK (status IN ('PENDING_PAYMENT', 'PAID', 'COMPLETED', 'CANCELLED')),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON platform.orders(user_id);
CREATE INDEX idx_orders_status ON platform.orders(status);
CREATE INDEX idx_orders_created_at ON platform.orders(created_at);

CREATE FUNCTION platform.set_orders_updated_at()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = NOW();
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_orders_set_updated_at
        BEFORE UPDATE ON platform.orders
        FOR EACH ROW
        EXECUTE FUNCTION platform.set_orders_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_orders_set_updated_at ON platform.orders;
DROP TABLE IF EXISTS platform.orders;
DROP FUNCTION IF EXISTS platform.set_orders_updated_at();
DROP SCHEMA IF EXISTS platform;
-- +goose StatementEnd
