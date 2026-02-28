-- Test fixture for FK inlining
-- Transformations tested:
--   - Simple FK inlined into column definition
--   - FK with ON DELETE inlined
--   - FK with ON UPDATE inlined
--   - Multi-column FK (cannot inline - passes through)
--   - FK with NOT VALID (cannot inline - passes through)
--   - FK with quoted constraint name (inlined)
--
-- Edge cases tested:
--   - FK with ON DELETE SET DEFAULT
--   - FK with ON DELETE SET NULL
--   - Self-referential FK
--
-- Negative tests (should NOT be inlined):
--   - FK with MATCH FULL
--   - FK with MATCH PARTIAL
--   - Multi-column FK

-- Simple FK (should inline)
CREATE TABLE public.categories (
    id bigint NOT NULL,
    name text NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.products (
    id bigint NOT NULL,
    category_id bigint NOT NULL REFERENCES public.categories(id),
    name text NOT NULL,
    PRIMARY KEY (id)
);

-- FK with ON DELETE CASCADE (should inline)
CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    total numeric(10,2) NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.order_items (
    id bigint NOT NULL,
    order_id bigint NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    product_id bigint NOT NULL,
    quantity integer NOT NULL,
    PRIMARY KEY (id)
);

-- FK with ON UPDATE SET NULL (should inline)
CREATE TABLE public.accounts (
    id bigint NOT NULL,
    parent_id bigint REFERENCES public.accounts(id) ON UPDATE SET NULL,
    PRIMARY KEY (id)
);

-- FK with quoted constraint name (should inline)
CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.comments (
    id bigint NOT NULL,
    user_id bigint NOT NULL REFERENCES public.users(id),
    content text NOT NULL,
    PRIMARY KEY (id)
);

-- Multi-column FK (cannot inline - passes through as ALTER TABLE)
CREATE TABLE public.order_shipments (
    order_id bigint NOT NULL,
    shipment_id bigint NOT NULL,
    tracking_number text,
    PRIMARY KEY (order_id, shipment_id)
);

ALTER TABLE ONLY public.order_shipments
    ADD CONSTRAINT order_shipments_order_fkey FOREIGN KEY (order_id, shipment_id) REFERENCES public.orders(id, id);

-- FK with NOT VALID (cannot inline - passes through)
CREATE TABLE public.queues (
    id bigint NOT NULL,
    name text NOT NULL,
    parent_id bigint
);

ALTER TABLE ONLY public.queues
    ADD CONSTRAINT queues_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.queues
    ADD CONSTRAINT queues_parent_fkey FOREIGN KEY (parent_id) REFERENCES queues(id) NOT VALID;

-- ============================================
-- Edge cases for FK inlining
-- ============================================

-- Edge case: FK with ON DELETE SET DEFAULT (should inline)
CREATE TABLE public.set_default_child (
    id bigint NOT NULL,
    parent_id bigint DEFAULT 1 REFERENCES public.set_default_child(id) ON DELETE SET DEFAULT,
    PRIMARY KEY (id)
);

-- Edge case: FK with ON DELETE SET NULL (should inline)
CREATE TABLE public.set_null_child (
    id bigint NOT NULL,
    parent_id bigint REFERENCES public.set_null_child(id) ON DELETE SET NULL,
    PRIMARY KEY (id)
);

-- Edge case: Self-referential FK (should inline)
CREATE TABLE public.employee_hierarchy (
    id bigint NOT NULL,
    name text NOT NULL,
    manager_id bigint REFERENCES public.employee_hierarchy(id),
    PRIMARY KEY (id)
);

-- ============================================
-- Negative tests for FK inlining
-- ============================================

-- Negative test: FK with MATCH FULL (cannot inline - passes through)
CREATE TABLE public.match_full_a (
    id bigint NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.match_full_b (
    id bigint NOT NULL,
    ref_id bigint NOT NULL
);

ALTER TABLE ONLY public.match_full_b ADD CONSTRAINT match_full_b_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.match_full_b
    ADD CONSTRAINT match_full_fkey FOREIGN KEY (ref_id) REFERENCES public.match_full_a(id) MATCH FULL;

-- Negative test: FK with MATCH PARTIAL (cannot inline - passes through)
CREATE TABLE public.match_partial_a (
    id bigint NOT NULL,
    PRIMARY KEY (id)
);

CREATE TABLE public.match_partial_b (
    id bigint NOT NULL,
    ref_id bigint
);

ALTER TABLE ONLY public.match_partial_b ADD CONSTRAINT match_partial_b_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.match_partial_b
    ADD CONSTRAINT match_partial_fkey FOREIGN KEY (ref_id) REFERENCES public.match_partial_a(id) MATCH PARTIAL;

-- Negative test: FK with ON DELETE NO ACTION (should pass through)
CREATE TABLE public.no_action_child (
    id bigint NOT NULL,
    parent_id bigint
);

ALTER TABLE ONLY public.no_action_child ADD CONSTRAINT no_action_child_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.no_action_child
    ADD CONSTRAINT no_action_child_parent_fkey FOREIGN KEY (parent_id) REFERENCES public.no_action_child(id) ON DELETE NO ACTION;
