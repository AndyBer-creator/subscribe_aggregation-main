-- 000001_init_schema.up.sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
<<<<<<< HEAD
    email TEXT NOT NULL,
=======
>>>>>>> 78bf63b (updated)
    cost NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
-- 000001_init_schema.down.sql
DROP TABLE IF EXISTS subscriptions;
DROP EXTENSION IF EXISTS "uuid-ossp";
