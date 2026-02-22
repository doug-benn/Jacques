-- Test fixture for ALTER SEQUENCE options
-- Covers: ALTER SEQUENCE options (RESTART, INCREMENT, MINVALUE, MAXVALUE, CACHE, etc.)

-- Sequence with options
CREATE SEQUENCE public.order_ids
    START WITH 1000
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 999999999
    CACHE 10;

-- Table using the sequence
CREATE TABLE public.orders (
    id bigint NOT NULL,
    customer_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

ALTER TABLE public.orders ALTER COLUMN id
    SET DEFAULT nextval('public.order_ids'::regclass);

-- ALTER SEQUENCE RESTART (resets sequence)
ALTER SEQUENCE public.order_ids RESTART WITH 2000;

-- Sequence with CYCLE option
CREATE SEQUENCE public.product_ids
    START WITH 1
    INCREMENT BY 1
    MINVALUE 1
    MAXVALUE 1000
    CACHE 5
    CYCLE;

ALTER SEQUENCE public.product_ids INCREMENT BY 2;

ALTER SEQUENCE public.product_ids RESTART WITH 100;

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

ALTER TABLE public.products ALTER COLUMN id
    SET DEFAULT nextval('public.product_ids'::regclass);
