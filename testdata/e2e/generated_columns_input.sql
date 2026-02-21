-- Test fixture for generated columns
-- Covers: GENERATED ALWAYS AS (expression) STORED

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) DEFAULT 0,
    final_price numeric(10,2) GENERATED ALWAYS AS (price - discount) STORED
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

CREATE TABLE public.users (
    id bigint NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    full_name text GENERATED ALWAYS AS (first_name || ' ' || last_name) STORED
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp without time zone NOT NULL DEFAULT NOW(),
    age interval GENERATED ALWAYS AS (updated_at - created_at) STORED
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

CREATE TABLE public.events (
    id bigint NOT NULL,
    data jsonb NOT NULL,
    event_type text NOT NULL,
    is_urgent boolean GENERATED ALWAYS AS (data->>'urgent' = 'true') STORED
);

ALTER TABLE ONLY public.events
    ADD CONSTRAINT events_pkey PRIMARY KEY (id);
