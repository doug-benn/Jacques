-- Test fixture for ALTER folding
-- Transformations tested:
--   - PRIMARY KEY constraint folded into CREATE TABLE
--   - UNIQUE constraint folded into CREATE TABLE
--   - CHECK constraint folded into CREATE TABLE
--
-- Edge cases tested:
--   - Inline NOT NULL preserved
--   - CHECK with complex expression
--   - Multiple constraints on same column
--
-- Negative tests (should NOT be folded):
--   - DEFERRABLE constraints
--   - INITIALLY DEFERRED constraints
--   - USING clause

-- Table with separate PRIMARY KEY (should fold)
CREATE TABLE public.users (
    id bigint NOT NULL,
    email text,
    name text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

-- Table with separate UNIQUE constraint (should fold)
CREATE TABLE public.products (
    id bigint NOT NULL,
    sku text NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_sku_key UNIQUE (sku);

-- Table with multi-column UNIQUE constraint (should fold)
CREATE TABLE public.orders (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    order_number text NOT NULL
);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_user_order_number_key UNIQUE (user_id, order_number);

-- Table with CHECK constraint (should fold)
CREATE TABLE public.accounts (
    id bigint NOT NULL,
    balance numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'active'
);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_balance_check CHECK (balance >= 0);

ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_status_check CHECK (status IN ('active', 'inactive', 'suspended'));

-- Table with multiple CHECK constraints (should all fold)
CREATE TABLE public.inventories (
    id bigint NOT NULL,
    product_id bigint NOT NULL,
    quantity integer NOT NULL,
    reorder_point integer NOT NULL
);

ALTER TABLE ONLY public.inventories
    ADD CONSTRAINT inventories_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.inventories
    ADD CONSTRAINT inventories_quantity_check CHECK (quantity >= 0);

ALTER TABLE ONLY public.inventories
    ADD CONSTRAINT inventories_reorder_check CHECK (reorder_point >= 0);

ALTER TABLE ONLY public.inventories
    ADD CONSTRAINT inventories_reorder_quantity_check CHECK (reorder_point <= quantity);

-- ============================================
-- Edge cases for ALTER folding
-- ============================================

-- Edge case: Table with inline NOT NULL (should preserve)
CREATE TABLE public.inline_notnull (
    id bigint NOT NULL,
    name text NOT NULL,
    email text NOT NULL
);

ALTER TABLE ONLY public.inline_notnull
    ADD CONSTRAINT inline_notnull_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.inline_notnull
    ADD CONSTRAINT inline_notnull_email_key UNIQUE (email);

-- Edge case: CHECK with complex expression (should fold)
CREATE TABLE public.complex_check (
    id bigint NOT NULL,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) NOT NULL,
    final_price numeric(10,2) NOT NULL
);

ALTER TABLE ONLY public.complex_check
    ADD CONSTRAINT complex_check_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.complex_check
    ADD CONSTRAINT complex_check_price_check CHECK (price >= 0 AND price < 10000);

ALTER TABLE ONLY public.complex_check
    ADD CONSTRAINT complex_check_final_check CHECK (final_price = price - discount);

-- Edge case: Multiple constraints on same column
CREATE TABLE public.multi_constraint (
    id bigint NOT NULL,
    code text NOT NULL
);

ALTER TABLE ONLY public.multi_constraint
    ADD CONSTRAINT multi_constraint_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.multi_constraint
    ADD CONSTRAINT multi_constraint_code_key UNIQUE (code);

-- Negative test: DEFERRABLE constraint (should NOT fold - pass through)
CREATE TABLE public.deferrable_unique (
    id bigint NOT NULL,
    email text NOT NULL
);

ALTER TABLE ONLY public.deferrable_unique
    ADD CONSTRAINT deferrable_unique_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.deferrable_unique
    ADD CONSTRAINT deferrable_unique_email_key UNIQUE (email) DEFERRABLE INITIALLY DEFERRED;

-- Negative test: NOT DEFERRABLE (should fold - same as default)
CREATE TABLE public.not_deferrable_unique (
    id bigint NOT NULL,
    email text NOT NULL
);

ALTER TABLE ONLY public.not_deferrable_unique
    ADD CONSTRAINT not_deferrable_unique_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.not_deferrable_unique
    ADD CONSTRAINT not_deferrable_unique_email_key UNIQUE (email) NOT DEFERRABLE;

-- Negative test: USING clause (should NOT fold - pass through)
CREATE TABLE public.using_clause (
    id bigint NOT NULL
);

ALTER TABLE ONLY public.using_clause
    ADD CONSTRAINT using_clause_pkey PRIMARY KEY (id) USING btree;
