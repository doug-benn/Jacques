CREATE TYPE order_status AS ENUM ('pending', 'processing', 'shipped', 'delivered', 'cancelled');

CREATE TYPE priority AS ENUM ('low', 'medium', 'high', 'urgent');

CREATE DOMAIN email AS text;

CREATE DOMAIN status AS text;

CREATE TABLE tickets (
    id bigint PRIMARY KEY,
    title text NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
    priority priority NOT NULL DEFAULT 'medium',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE projects (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    tags text[],
    milestones text[],
    budget numeric(10,2)
);

CREATE TABLE events (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    metadata jsonb,
    payload json,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE task_assignments (
    id bigint PRIMARY KEY,
    task_id bigint NOT NULL,
    assignees bigint[],
    status order_status NOT NULL DEFAULT 'pending',
    labels text[],
    extra_data jsonb,
    notes text
);

CREATE TABLE users (
    id bigint PRIMARY KEY,
    email email NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    status status NOT NULL DEFAULT 'pending'
);

CREATE TABLE reservations (
    id bigint PRIMARY KEY,
    room_number text NOT NULL,
    stay_period tstzrange NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE inventory (
    id bigint PRIMARY KEY,
    product_id bigint NOT NULL,
    quantity_range int4range NOT NULL,
    last_updated timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE user_profiles (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    birth_date date,
    age_range int4range,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE order_versions (
    id bigint PRIMARY KEY,
    order_id bigint NOT NULL,
    status_range daterange NOT NULL,
    status order_status NOT NULL
);
