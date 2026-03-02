CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email text UNIQUE,
    created_at timestamp without time zone
);
