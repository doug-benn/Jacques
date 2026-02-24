CREATE TABLE public.products (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    amount numeric(10,2),
    source bigint,
    CONSTRAINT orders_amount_check CHECK (amount > 0)
);

CREATE TABLE public.order_items (
    order_id bigint REFERENCES public.orders(id) NOT NULL,
    product_id bigint REFERENCES public.products(id) NOT NULL,
    qty integer NOT NULL,
    PRIMARY KEY (order_id, product_id)
);

CREATE TABLE public.queues (
    id bigint PRIMARY KEY,
    name text NOT NULL,
    parent_id bigint REFERENCES queues(id)
);

ALTER TABLE public.queues
    ADD CONSTRAINT "self ref for queues" FOREIGN KEY (parent_id) REFERENCES queues(id) NOT VALID;
