CREATE SEQUENCE public.tags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

CREATE TABLE public.users (
    id BIGSERIAL PRIMARY KEY,
    email text NOT NULL
);

CREATE TABLE public.accounts (
    id SERIAL PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.tags (
    id SMALLSERIAL PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.orders (
    id BIGSERIAL PRIMARY KEY,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL DEFAULT 0
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
