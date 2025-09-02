-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL,
    user_id UUID NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE
);

-- +goose Down
DROP TABLE subscriptions;
