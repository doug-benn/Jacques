-- Test fixture for multiple schemas
-- Covers: CREATE SCHEMA, schema-qualified names, cross-schema FKs

CREATE SCHEMA app;

CREATE TABLE public.countries (
    id bigint NOT NULL,
    code text NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.countries
    ADD CONSTRAINT countries_pkey PRIMARY KEY (id);

CREATE TABLE app.users (
    id bigint NOT NULL,
    email text NOT NULL,
    country_id bigint
);

ALTER TABLE ONLY app.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

ALTER TABLE ONLY app.users
    ADD CONSTRAINT users_country_fkey FOREIGN KEY (country_id) REFERENCES public.countries(id);
