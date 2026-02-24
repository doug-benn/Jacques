-- Test fixture for quoted identifier handling
-- Features tested:
--   - Quoted constraint names: "pk constraint", "fk constraint"
--   - ALTER folding with quoted constraint names
--
-- Input: pg_dump output with quoted constraint names
-- Expected: Quoted constraint names preserved, constraints folded

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text
);

CREATE TABLE public.order_items (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    product_id bigint NOT NULL
);

-- ALTER with quoted constraint names
ALTER TABLE ONLY public.users
    ADD CONSTRAINT "users pk" PRIMARY KEY (id);

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT "order items pk" PRIMARY KEY (id);

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT "order items user fk" FOREIGN KEY (user_id) REFERENCES public.users(id);
