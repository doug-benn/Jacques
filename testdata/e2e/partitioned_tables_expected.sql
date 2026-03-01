CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at date NOT NULL,
    total numeric(10,2) NOT NULL,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    category text NOT NULL,
    price numeric(10,2) NOT NULL,
    PRIMARY KEY (id, category)
) PARTITION BY LIST (category);

CREATE TABLE public.orders_2024 PARTITION OF public.orders
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE public.orders_2025 PARTITION OF public.orders
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

CREATE TABLE public.products_electronics PARTITION OF public.products
    FOR VALUES IN ('electronics', 'computers', 'phones');

CREATE TABLE public.products_clothing PARTITION OF public.products
    FOR VALUES IN ('clothing', 'shoes', 'accessories');
