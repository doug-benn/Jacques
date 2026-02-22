-- Test fixture for foreign key and CHECK constraint handling
-- Features tested:
--   - ALTER folding: PK, CHECK folded into CREATE TABLE
--   - FK inlining: Inline FK into column definition where possible
--   - ONLY removal: ALTER TABLE ONLY → ALTER TABLE
--
-- Input: pg_dump output with separate ALTER TABLE statements
-- Expected: Constraints folded into CREATE TABLE, FK inlined

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    amount numeric(10,2)
);

CREATE TABLE public.order_items (
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    qty integer NOT NULL
);

ALTER TABLE ONLY public.orders       -- ONLY removed
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);  -- Folded into table

ALTER TABLE ONLY public.order_items  -- ONLY removed
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (order_id, product_id);  -- Folded into table

ALTER TABLE ONLY public.orders       -- ONLY removed
    ADD CONSTRAINT orders_amount_check CHECK (amount > 0);  -- Folded into table

ALTER TABLE ONLY public.order_items   -- ONLY removed
    ADD CONSTRAINT order_items_order_id_fkey
    FOREIGN KEY (order_id) REFERENCES public.orders(id);  -- Passed through (not foldable)
