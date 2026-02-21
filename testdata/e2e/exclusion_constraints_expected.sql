CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    email text NOT NULL
);
