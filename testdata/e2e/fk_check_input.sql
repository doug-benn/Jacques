-- Test fixture for foreign key and CHECK constraint handling
-- Features tested:
--   - ALTER folding: PK, CHECK folded into CREATE TABLE
--   - FK inlining: Inline FK into column definition where possible
--   - ONLY removal: ALTER TABLE ONLY → ALTER TABLE
--   - Quoted constraint names: Handle constraints with spaces in names
--   - NOT VALID FK: Pass through FK constraints that are NOT VALID
--
-- Input: pg_dump output with separate ALTER TABLE statements
-- Expected: Constraints folded into CREATE TABLE, FK inlined

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    amount numeric(10,2),
    source bigint
);

CREATE TABLE public.order_items (
    order_id bigint NOT NULL,
    product_id bigint NOT NULL,
    qty integer NOT NULL
);

-- Table with self-referential FK (quoted constraint name)
CREATE TABLE public.queues (
    id bigint NOT NULL,
    name text NOT NULL,
    parent_id bigint
);

ALTER TABLE ONLY public.queues
    ADD CONSTRAINT queues_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.orders       -- ONLY removed
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);  -- Folded into table

ALTER TABLE ONLY public.order_items  -- ONLY removed
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (order_id, product_id);  -- Folded into table

ALTER TABLE ONLY public.orders       -- ONLY removed
    ADD CONSTRAINT orders_amount_check CHECK (amount > 0);  -- Folded into table

ALTER TABLE ONLY public.order_items   -- ONLY removed
    ADD CONSTRAINT order_items_order_id_fkey
    FOREIGN KEY (order_id) REFERENCES public.orders(id);  -- Passed through (not foldable)

-- FK with quoted constraint name (should be inlined)
ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT "order_items product fk" FOREIGN KEY (product_id) REFERENCES public.products(id);

-- FK with NOT VALID (should pass through - can't inline unvalidated FKs)
ALTER TABLE ONLY public.queues
    ADD CONSTRAINT "self ref for queues" FOREIGN KEY (parent_id) REFERENCES queues(id) NOT VALID;
