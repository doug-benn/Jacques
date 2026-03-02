CREATE SEQUENCE global_id_seq START WITH 1 INCREMENT BY 1 NO MAXVALUE CACHE 1;

CREATE SEQUENCE public.tags_id_seq
    START WITH 1 INCREMENT BY 1 NO MINVALUE NO MAXVALUE CACHE 1;

CREATE TABLE public.orders (
    id bigint PRIMARY KEY DEFAULT nextval('global_id_seq'::regclass),
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL DEFAULT 0,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY DEFAULT nextval('global_id_seq'::regclass),
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL DEFAULT 1
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY DEFAULT nextval('global_id_seq'::regclass),
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE TABLE public.users (
    id BIGSERIAL PRIMARY KEY,
    email text NOT NULL UNIQUE,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.accounts (
    id SERIAL PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.tags (
    id SMALLSERIAL PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.custom_seq_table (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE public.unowned_table (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE public.bounded_table (
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE public.cached_table (
    id BIGSERIAL PRIMARY KEY
);
