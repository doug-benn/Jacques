CREATE TYPE contact_info AS (
    email text,
    phone text,
    mobile text,
    fax text
);

CREATE DOMAIN email AS text
CHECK (VALUE ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z2-9]+$');

CREATE DOMAIN positive_int AS integer
CHECK (VALUE > 0);

CREATE DOMAIN phone_number AS text
CHECK (VALUE ~ '^\+?[0-9]{10,15}$');

CREATE DOMAIN order_status AS text
CHECK (VALUE IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled'));

CREATE TABLE customers (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    contact contact_info,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id bigint PRIMARY KEY,
    email email NOT NULL UNIQUE,
    phone phone_number,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    quantity positive_int NOT NULL,
    status order_status NOT NULL DEFAULT 'pending'
);
