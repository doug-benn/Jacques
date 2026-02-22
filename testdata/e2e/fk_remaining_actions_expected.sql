CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE
);

CREATE TABLE public.accounts (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) ON DELETE SET DEFAULT NOT NULL,
    account_type text NOT NULL DEFAULT 'basic'
);

CREATE TABLE public.categories (
    id bigint PRIMARY KEY,
    parent_id bigint REFERENCES public.categories(id) ON DELETE RESTRICT,
    name text NOT NULL
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    category_id bigint REFERENCES public.categories(id) ON DELETE NO ACTION NOT NULL,
    name text NOT NULL
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES public.orders(id) ON UPDATE CASCADE NOT NULL,
    product_id bigint NOT NULL
);

CREATE TABLE public.shipments (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES public.orders(id) ON UPDATE SET NULL NOT NULL,
    tracking_number text
);

CREATE TABLE public.addresses (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) NOT NULL,
    address_line1 text NOT NULL,
    city text NOT NULL
);

ALTER TABLE public.categories
    ADD CONSTRAINT categories_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.categories(id) ON DELETE RESTRICT;
