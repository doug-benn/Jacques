-- Test fixture for schema handling
-- Covers: CREATE SCHEMA preservation, ordering, inference, and public stripping

-- Explicit schema
CREATE SCHEMA app;

-- Table in public schema (public should be stripped)
CREATE TABLE public.base_table (
    id bigint NOT NULL
);
ALTER TABLE ONLY public.base_table ADD PRIMARY KEY (id);

-- Table in explicit schema
CREATE TABLE app.app_table (
    id bigint NOT NULL,
    base_id bigint
);
ALTER TABLE ONLY app.app_table ADD PRIMARY KEY (id);
ALTER TABLE ONLY app.app_table ADD CONSTRAINT app_table_fkey FOREIGN KEY (base_id) REFERENCES public.base_table(id);

-- Table in quoted schema (should be inferred)
CREATE TABLE "MySchema".profile (
    id bigint NOT NULL,
    bio text
);
ALTER TABLE ONLY "MySchema".profile ADD PRIMARY KEY (id);
