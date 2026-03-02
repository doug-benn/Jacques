CREATE TABLE users (
    id bigint PRIMARY KEY,
    username text NOT NULL,
    email text,
    status varchar(50),
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    price integer,
    description text NOT NULL,
    is_active boolean NOT NULL DEFAULT true
);

CREATE TABLE logs (
    id bigint PRIMARY KEY,
    message text NOT NULL,
    severity text
);
