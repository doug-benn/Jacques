CREATE TABLE users (
    id bigint NOT NULL,
    email text NOT NULL
);

CREATE TABLE orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL
);

DROP TABLE IF EXISTS users;

DROP TABLE IF EXISTS orders;

DROP INDEX IF EXISTS idx_users_email;

DROP INDEX IF EXISTS idx_orders_user;

CREATE INDEX idx_users_email ON users(email);

CREATE INDEX idx_orders_user ON orders(user_id);
