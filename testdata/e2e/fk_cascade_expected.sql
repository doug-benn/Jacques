CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.categories (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    parent_id bigint
);

CREATE TABLE public.accounts (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    balance numeric(10,2) NOT NULL DEFAULT 0
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    category_id bigint REFERENCES public.categories(id) ON UPDATE CASCADE
);

CREATE TABLE public.posts (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) ON DELETE CASCADE NOT NULL,
    title text NOT NULL
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) ON DELETE RESTRICT NOT NULL,
    total numeric(10,2) NOT NULL
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES public.orders(id) ON DELETE CASCADE NOT NULL,
    product_id bigint REFERENCES public.products(id) ON DELETE SET DEFAULT NOT NULL,
    quantity integer NOT NULL DEFAULT 1
);

CREATE TABLE public.comments (
    id bigint PRIMARY KEY,
    post_id bigint REFERENCES public.posts(id) ON DELETE SET NULL,
    user_id bigint REFERENCES public.users(id) ON DELETE SET NULL,
    content text NOT NULL
);

CREATE TABLE public.transactions (
    id bigint PRIMARY KEY,
    account_id bigint REFERENCES public.accounts(id) ON DELETE CASCADE ON UPDATE RESTRICT NOT NULL,
    amount numeric(10,2) NOT NULL,
    type text NOT NULL
);

CREATE TABLE public.user_profiles (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) MATCH FULL NOT NULL,
    bio text
);
