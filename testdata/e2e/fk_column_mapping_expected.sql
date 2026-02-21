CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE,
    name text NOT NULL,
    UNIQUE (id, name)
);

CREATE TABLE public.user_profiles (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    user_name text NOT NULL UNIQUE,
    bio text
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) NOT NULL,
    total numeric(10,2) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    sku text NOT NULL UNIQUE,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES public.orders(id) NOT NULL,
    product_sku text REFERENCES public.products(sku) NOT NULL,
    quantity integer NOT NULL
);

ALTER TABLE public.user_profiles
    ADD CONSTRAINT user_profiles_user_fkey FOREIGN KEY (user_id, user_name) REFERENCES public.users(id, name);
