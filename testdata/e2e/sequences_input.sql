-- Test fixture for sequence handling
-- Covers: shared sequences, SERIAL conversion, sequence preservation, edge cases
--
-- Transformations tested:
--   - Dedicated sequence → BIGSERIAL conversion
--   - Dedicated sequence → SERIAL conversion
--   - Dedicated sequence → SMALLSERIAL conversion
--   - Shared sequence → preserved (not converted)
--
-- Edge cases tested:
--   - Sequence with explicit START WITH
--   - Sequence with explicit INCREMENT BY
--   - Sequence without OWNED BY
--   - Sequence with MINVALUE/MAXVALUE
--
-- Negative tests (should NOT be converted):
--   - Sequence with CACHE > 1 (could have gaps)

-- ============================================
-- Shared sequence - used by multiple tables, should be preserved
-- ============================================
CREATE SEQUENCE global_id_seq START WITH 1 INCREMENT BY 1 NO MAXVALUE CACHE 1;

-- Table using shared sequence - should preserve sequence (not convert to SERIAL)
CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL DEFAULT 0,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE public.orders ALTER COLUMN id
    SET DEFAULT nextval('global_id_seq'::regclass);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

-- Table using shared sequence for order_items
CREATE TABLE public.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL DEFAULT 1
);

ALTER TABLE public.order_items ALTER COLUMN id
    SET DEFAULT nextval('global_id_seq'::regclass);

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);

-- Another table using the same shared sequence
CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

ALTER TABLE public.products ALTER COLUMN id
    SET DEFAULT nextval('global_id_seq'::regclass);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

-- ============================================
-- Dedicated sequences - should convert to SERIAL types
-- ============================================

-- Table with dedicated sequence - should convert to BIGSERIAL
CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE SEQUENCE public.users_id_seq
    START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;

ALTER TABLE public.users ALTER COLUMN id
    SET DEFAULT nextval('public.users_id_seq'::regclass);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);

-- Table using integer with dedicated sequence - should convert to SERIAL
CREATE TABLE public.accounts (
    id integer NOT NULL,
    name text NOT NULL
);

CREATE SEQUENCE public.accounts_id_seq
    START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

ALTER SEQUENCE public.accounts_id_seq OWNED BY public.accounts.id;

ALTER TABLE public.accounts ALTER COLUMN id
    SET DEFAULT nextval('public.accounts_id_seq'::regclass);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

-- Table using smallint with dedicated sequence - should convert to SMALLSERIAL
CREATE TABLE public.tags (
    id smallint NOT NULL,
    name text NOT NULL
);

CREATE SEQUENCE public.tags_id_seq
    START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

ALTER SEQUENCE public.tags_id_seq OWNED BY public.tags.id;

ALTER TABLE public.tags ALTER COLUMN id
    SET DEFAULT nextval('public.tags_id_seq'::regclass);

ALTER TABLE ONLY public.tags
    ADD CONSTRAINT tags_pkey PRIMARY KEY (id);

-- ============================================
-- Edge cases for sequence handling
-- ============================================

-- Edge case: Sequence with explicit START WITH (should convert)
CREATE SEQUENCE public.custom_start_seq
    START WITH 1000
    INCREMENT BY 10
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.custom_seq_table (
    id bigint NOT NULL
);

ALTER SEQUENCE public.custom_start_seq OWNED BY public.custom_seq_table.id;
ALTER TABLE public.custom_seq_table ALTER COLUMN id SET DEFAULT nextval('public.custom_start_seq');

-- Edge case: Sequence without explicit OWNED BY (should convert)
CREATE SEQUENCE public.unowned_seq;

CREATE TABLE public.unowned_table (
    id bigint NOT NULL DEFAULT nextval('public.unowned_seq'::regclass)
);

-- Edge case: Sequence with MINVALUE/MAXVALUE (should convert)
CREATE SEQUENCE public.bounded_seq
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 1000000
    CACHE 1;

CREATE TABLE public.bounded_table (
    id bigint NOT NULL
);

ALTER SEQUENCE public.bounded_seq OWNED BY public.bounded_table.id;
ALTER TABLE public.bounded_table ALTER COLUMN id SET DEFAULT nextval('public.bounded_seq'::regclass);

-- Negative test: Sequence with CACHE > 1 (should NOT convert - could have gaps)
CREATE SEQUENCE public.cached_seq
    START WITH 1
    INCREMENT BY 1
    NO MAXVALUE
    CACHE 100;

CREATE TABLE public.cached_table (
    id bigint NOT NULL
);

ALTER SEQUENCE public.cached_seq OWNED BY public.cached_table.id;
ALTER TABLE public.cached_table ALTER COLUMN id SET DEFAULT nextval('public.cached_seq'::regclass);

-- Ensure tables have PKs
ALTER TABLE ONLY public.custom_seq_table ADD CONSTRAINT custom_seq_table_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.unowned_table ADD CONSTRAINT unowned_table_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.bounded_table ADD CONSTRAINT bounded_table_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.cached_table ADD CONSTRAINT cached_table_pkey PRIMARY KEY (id);
