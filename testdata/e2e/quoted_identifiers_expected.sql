CREATE TABLE public.users (
    id bigint PRIMARY KEY,
    email text
);

CREATE TABLE public.order_items (
    id bigint PRIMARY KEY,
    user_id bigint REFERENCES public.users(id) NOT NULL,
    product_id bigint NOT NULL
);
