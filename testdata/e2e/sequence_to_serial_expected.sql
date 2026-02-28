CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL
);

ALTER TABLE ONLY public.users ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE TABLE public.accounts (
    id integer NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.accounts ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

CREATE TABLE public.tags (
    id smallint NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.tags ADD CONSTRAINT tags_pkey PRIMARY KEY (id);

CREATE SEQUENCE global_id_seq START WITH 1 INCREMENT BY 1 NO MAXVALUE CACHE 1;

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL DEFAULT 0
);

ALTER TABLE ONLY public.orders ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

CREATE TABLE public.custom_seq_table (
    id bigint NOT NULL
);

CREATE TABLE public.unowned_table (
    id bigint NOT NULL
);

CREATE TABLE public.bounded_table (
    id bigint NOT NULL
);

CREATE SEQUENCE public.cached_seq
    START WITH 1
    INCREMENT BY 1
    NO MAXVALUE
    CACHE 100;

CREATE TABLE public.cached_table (
    id bigint NOT NULL
);

ALTER TABLE public.cached_table ALTER COLUMN id SET DEFAULT nextval('public.cached_seq'::regclass);

ALTER TABLE ONLY public.custom_seq_table ADD CONSTRAINT custom_seq_table_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.unowned_table ADD CONSTRAINT unowned_table_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.bounded_table ADD CONSTRAINT bounded_table_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.cached_table ADD CONSTRAINT cached_table_pkey PRIMARY KEY (id);
