-- Test fixture for domain types
-- Covers: CREATE DOMAIN, domain usage in tables

-- Domain for email
CREATE DOMAIN public.email AS text
CHECK (VALUE ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z2-9]+$');

-- Domain for positive numbers
CREATE DOMAIN public.positive_int AS integer
CHECK (VALUE > 0);

-- Domain for phone number
CREATE DOMAIN public.phone_number AS text
CHECK (VALUE ~ '^\+?[0-9]{10,15}$');

-- Domain for status (like enum)
CREATE DOMAIN public.order_status AS text
CHECK (VALUE IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled'));

-- Table using domains
CREATE TABLE public.users (
    id bigint NOT NULL,
    email public.email NOT NULL,
    phone public.phone_number,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

-- Table with positive integer domain
CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    quantity public.positive_int NOT NULL,
    status public.order_status NOT NULL DEFAULT 'pending'
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);
