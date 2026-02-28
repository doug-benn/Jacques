CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE TABLE public.edge_cases (
    id bigint NOT NULL
);

CREATE FUNCTION public.set_in_function() RETURNS void AS $$
BEGIN
    SET statement_timeout = 3600;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE ONLY public.edge_cases ADD CONSTRAINT edge_cases_pkey PRIMARY KEY (id);
