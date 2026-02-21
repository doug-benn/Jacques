CREATE TABLE public.orders (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL,
    amount numeric(10,2),
    CONSTRAINT orders_amount_check CHECK (amount > 0)
);

CREATE TABLE public.order_items (
    order_id bigint REFERENCES public.orders(id) NOT NULL,
    product_id bigint NOT NULL,
    qty integer NOT NULL,
    PRIMARY KEY (order_id, product_id)
);
