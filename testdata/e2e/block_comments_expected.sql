CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL
);
