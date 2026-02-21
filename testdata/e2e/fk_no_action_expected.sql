CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text NOT NULL UNIQUE,
    name text NOT NULL
);

CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE TABLE public.posts (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) ON DELETE NO ACTION NOT NULL,
    title text NOT NULL,
    content text
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) ON DELETE NO ACTION NOT NULL,
    total numeric(10,2) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES public.orders(id) ON DELETE CASCADE NOT NULL,
    product_id bigint REFERENCES public.products(id) ON DELETE NO ACTION ON UPDATE NO ACTION NOT NULL,
    quantity integer NOT NULL DEFAULT 1
);

CREATE TABLE public.comments (
    id bigint PRIMARY KEY,
    post_id bigint REFERENCES public.posts(id) ON DELETE NO ACTION,
    user_id bigint REFERENCES public.users(id) ON DELETE NO ACTION ON UPDATE NO ACTION,
    content text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);
