CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    PRIMARY KEY (id)
);
