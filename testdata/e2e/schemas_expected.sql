CREATE SCHEMA app;

CREATE TABLE public.countries (
    id bigint PRIMARY KEY,
    code text NOT NULL,
    name text NOT NULL
);

CREATE TABLE app.users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    country_id bigint REFERENCES public.countries(id)
);
