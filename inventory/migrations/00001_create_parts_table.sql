-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS platform;

CREATE TABLE platform.parts (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price NUMERIC(12, 2) NOT NULL CHECK (price >= 0),
    stock_quantity BIGINT NOT NULL CHECK (stock_quantity >= 0),
    category SMALLINT NOT NULL,
    dimensions JSONB,
    manufacturer JSONB,
    tags TEXT[] NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_parts_name ON platform.parts (name);
CREATE INDEX idx_parts_category ON platform.parts (category);
CREATE INDEX idx_parts_manufacturer_country ON platform.parts((manufacturer->>'country'));
CREATE INDEX idx_parts_tags ON platform.parts USING GIN (tags);

CREATE FUNCTION platform.set_parts_updated_at()
    RETURNS TRIGGER AS $$
    BEGIN
        NEW.updated_at = NOW();
        RETURN NEW;
    END;
    $$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_parts_set_updated_at
        BEFORE UPDATE ON platform.parts
        FOR EACH ROW
        EXECUTE FUNCTION platform.set_parts_updated_at();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_parts_set_updated_at ON platform.parts;
DROP TABLE IF EXISTS platform.parts;
DROP FUNCTION IF EXISTS platform.set_parts_updated_at();
DROP SCHEMA IF EXISTS platform;
-- +goose StatementEnd
