CREATE TABLE users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    name text NOT NULL,
    status text NOT NULL DEFAULT 'active'
);

CREATE TABLE orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE VIEW active_users AS
SELECT id, email, name
FROM users
WHERE status = 'active';

CREATE VIEW user_orders AS
SELECT u.id as user_id, u.email, o.id as order_id, o.total
FROM users u
JOIN orders o ON u.id = o.user_id;

CREATE VIEW expensive_products AS
SELECT id, name, price
FROM products
WHERE price > (SELECT AVG(price) FROM products);

CREATE VIEW order_details AS
SELECT 
    o.id as order_id,
    u.name as customer_name,
    u.email,
    p.name as product_name,
    o.total as order_total
FROM orders o
JOIN users u ON o.user_id = u.id
JOIN products p ON p.id = 1;

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
