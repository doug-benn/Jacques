-- Test fixture for block comment removal
/* This is a header block comment */
CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL
);

/* Table for products */
CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);
