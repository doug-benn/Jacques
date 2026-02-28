CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    name text NOT NULL,
    status text NOT NULL DEFAULT 'active'
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'pending'
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE VIEW active_users AS
SELECT id, email, name
FROM public.users
WHERE status = 'active';

CREATE VIEW user_orders AS
SELECT u.id as user_id, u.email, o.id as order_id, o.total
FROM public.users u
JOIN public.orders o ON u.id = o.user_id;

CREATE VIEW user_order_totals AS
SELECT user_id, email, SUM(total) as total_spent
FROM user_orders
GROUP BY user_id, email;

CREATE VIEW expensive_products AS
SELECT id, name, price
FROM public.products
WHERE price > (SELECT AVG(price) FROM public.products);

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
