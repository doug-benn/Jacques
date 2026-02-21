CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE VIEW order_summary AS
SELECT o.id, o.total, o.status, o.created_at, u.name as user_name
FROM orders o
JOIN users u ON o.user_id = u.id;

CREATE MATERIALIZED VIEW order_stats AS
SELECT status, COUNT(*) as count, SUM(total) as total_amount
FROM orders
GROUP BY status;

CREATE INDEX idx_order_stats_status ON order_stats(status);

CREATE MATERIALIZED VIEW user_order_summary AS
SELECT u.id as user_id, u.name, COUNT(o.id) as order_count, COALESCE(SUM(o.total), 0) as total_spent
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name;

CREATE INDEX idx_user_order_user_id ON user_order_summary(user_id);
