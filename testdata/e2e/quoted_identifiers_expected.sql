CREATE TABLE public.users (
    id bigint NOT NULL,
    email text,
    CONSTRAINT "users pk" PRIMARY KEY (id)
);

CREATE TABLE public.order_items (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    product_id bigint NOT NULL,
    CONSTRAINT "order items pk" PRIMARY KEY (id)
);

ALTER TABLE public.order_items
    ADD CONSTRAINT "order items user fk" FOREIGN KEY (user_id) REFERENCES public.users(id);
