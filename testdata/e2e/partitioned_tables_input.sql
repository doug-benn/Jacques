-- Test fixture for partitioned tables (E2E testable)
-- Covers: Basic PARTITION BY RANGE, PARTITION BY LIST
-- Note: Complex partitioning still requires ExperimentalFolding

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at date NOT NULL,
    total numeric(10,2) NOT NULL
) PARTITION BY RANGE (created_at);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id, created_at);

CREATE TABLE public.orders_2024 PARTITION OF public.orders
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE public.orders_2025 PARTITION OF public.orders
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    category text NOT NULL,
    price numeric(10,2) NOT NULL
) PARTITION BY LIST (category);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id, category);

CREATE TABLE public.products_electronics PARTITION OF public.products
    FOR VALUES IN ('electronics', 'computers', 'phones');

CREATE TABLE public.products_clothing PARTITION OF public.products
    FOR VALUES IN ('clothing', 'shoes', 'accessories');

-- Partition by HASH
CREATE TABLE public.users (
    id bigint NOT NULL,
    name text NOT NULL
) PARTITION BY HASH (id);

CREATE TABLE public.users_1 PARTITION OF public.users FOR VALUES WITH (MODULUS 4, REMAINDER 0);
