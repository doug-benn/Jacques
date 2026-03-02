-- Test fixture for Views
-- Covers: Regular views, materialized views, views with joins, views with subqueries

-- Base tables for views
CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    status text NOT NULL DEFAULT 'active'
);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'pending',
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.orders ADD CONSTRAINT orders_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.products ADD CONSTRAINT products_pkey PRIMARY KEY (id);

-- ============================================
-- Regular Views
-- ============================================

-- Simple view
CREATE VIEW active_users AS
SELECT id, email, name
FROM public.users
WHERE status = 'active';

-- View with join
CREATE VIEW user_orders AS
SELECT u.id as user_id, u.email, o.id as order_id, o.total
FROM public.users u
JOIN public.orders o ON u.id = o.user_id;
-- TODO (pg-schema-diff limitation): Views depending on other views
-- CREATE VIEW user_order_totals AS
-- SELECT user_id, email, SUM(total) as total_spent
-- FROM user_orders
-- GROUP BY user_id, email;

-- View with subquery
CREATE VIEW expensive_products AS
SELECT id, name, price
FROM public.products
WHERE price > (SELECT AVG(price) FROM public.products);

-- View using multiple joins
CREATE VIEW order_details AS
SELECT 
    o.id as order_id,
    u.name as customer_name,
    u.email,
    p.name as product_name,
    o.total as order_total
FROM public.orders o
JOIN public.users u ON o.user_id = u.id
JOIN public.products p ON p.id = 1;

-- Regular view in materialized_views style
CREATE VIEW order_summary AS
SELECT o.id, o.total, o.status, o.created_at, u.name as user_name
FROM orders o
JOIN users u ON o.user_id = u.id;

-- ============================================
-- Materialized Views
-- ============================================

-- Materialized view with aggregation
CREATE MATERIALIZED VIEW order_stats AS
SELECT status, COUNT(*) as count, SUM(total) as total_amount
FROM orders
GROUP BY status;

-- Index on materialized view
CREATE INDEX idx_order_stats_status ON order_stats(status);

-- Another materialized view with JOIN
CREATE MATERIALIZED VIEW user_order_summary AS
SELECT u.id as user_id, u.name, COUNT(o.id) as order_count, COALESCE(SUM(o.total), 0) as total_spent
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
GROUP BY u.id, u.name;

CREATE INDEX idx_user_order_user_id ON user_order_summary(user_id);
