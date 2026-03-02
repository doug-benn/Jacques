CREATE TABLE users (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE administrators (
    role text NOT NULL DEFAULT 'admin'
) INHERITS (users);
