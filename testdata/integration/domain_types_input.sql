-- Test fixture for domain types
-- Features tested:
--   - CREATE DOMAIN: Custom type definitions with constraints
--   - Domain constraints: CHECK constraints on domains
--   - Regex validation: Email pattern validation
--   - Value constraints: Positive integers, allowed values
--   - Domain usage: Tables using domain types as column types
--   - Schema-qualified domains: public.email, public.positive_int
--
-- Input: pg_dump output with domain definitions
-- Expected: Clean domain type output
--
-- Note: Domain types are folded when --experimental-folding is NOT used

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
