CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    created_at date NOT NULL,
    total numeric(10,2) NOT NULL,
    PRIMARY KEY (id, created_at)
);

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    category text NOT NULL,
    price numeric(10,2) NOT NULL,
    PRIMARY KEY (id, category)
);

CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    username text NOT NULL,
    region text NOT NULL
);
