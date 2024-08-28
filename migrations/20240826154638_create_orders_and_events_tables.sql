-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

CREATE TABLE orders (
    order_id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    is_final BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE order_events (
    event_id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(order_id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    order_status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_final BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_order_events_order_id ON order_events(order_id);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.

DROP TABLE IF EXISTS order_events;
DROP TABLE IF EXISTS orders;
