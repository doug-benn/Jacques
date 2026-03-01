-- Test fixture for basic domain types (E2E testable)
-- Covers: Simple DOMAIN without CHECK constraints
-- Note: DOMAIN with CHECK constraints requires ExperimentalFolding

-- Basic domain without constraints
CREATE DOMAIN public.email AS text;

CREATE TABLE public.users (
    id bigint NOT NULL,
    email public.email NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- Basic domain used in multiple tables
CREATE DOMAIN public.status AS text;

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    status public.status NOT NULL DEFAULT 'pending'
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);
