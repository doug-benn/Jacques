-- Test fixture for XML type columns
-- Covers: xml data type in PostgreSQL

CREATE TABLE public.documents (
    id bigint NOT NULL,
    title text NOT NULL,
    content xml,
    metadata xml,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

ALTER TABLE ONLY public.documents
    ADD CONSTRAINT documents_pkey PRIMARY KEY (id);

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    specifications xml,
    category text NOT NULL
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);
