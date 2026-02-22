CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) ON DELETE CASCADE NOT NULL,
    total numeric(10,2) NOT NULL
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY,
    order_id bigint REFERENCES public.orders(id) ON DELETE CASCADE NOT NULL,
    product_id bigint REFERENCES public.users(id) ON DELETE SET NULL NOT NULL,
    quantity integer NOT NULL DEFAULT 1
);
