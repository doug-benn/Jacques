-- Test fixture for sequence handling
-- Covers: shared sequences, SERIAL conversion, sequence preservation

-- Shared sequence - used by multiple columns/tables, should be preserved
CREATE SEQUENCE global_id_seq START WITH 1 INCREMENT BY 1 NO MAXVALUE CACHE 1;

-- Table with dedicated sequence - should convert to SERIAL
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

-- Table using shared sequence for order_number
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

-- Table using smallint with sequence - should NOT convert to SERIAL (smallint not supported)
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
