-- Test fixture for ALTER folding
-- Transformations tested:
--   - PRIMARY KEY, UNIQUE, CHECK constraints folded into CREATE TABLE
--
-- Edge cases tested:
--   - Multi-column constraints
--   - Complex CHECK expressions
--   - Multiple constraints on same column
--   - USING clause (index method)
--   - DEFERRABLE and INITIALLY DEFERRED constraints
--   - Negative test: NOT DEFERRABLE (should fold)

-- ============================================
-- Basic Folding: PK, UNIQUE, CHECK
-- ============================================
CREATE TABLE folding_basics (
    id bigint NOT NULL,
    sku text NOT NULL,
    user_id bigint NOT NULL,
    order_number text NOT NULL,
    balance numeric(10,2) NOT NULL,
    status text NOT NULL DEFAULT 'active'
);

ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_pkey PRIMARY KEY (id);
ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_sku_key UNIQUE (sku);
ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_multi_unique UNIQUE (user_id, order_number);
ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_balance_check CHECK (balance >= 0);
ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_status_check CHECK (status IN ('active', 'inactive'));

-- ============================================
-- Complex Expressions & USING Clause
-- ============================================
CREATE TABLE folding_complex (
    id bigint NOT NULL,
    price numeric(10,2) NOT NULL,
    discount numeric(10,2) NOT NULL,
    final_price numeric(10,2) NOT NULL,
    code text NOT NULL
);

ALTER TABLE ONLY folding_complex ADD PRIMARY KEY (id) USING btree;
ALTER TABLE ONLY folding_complex ADD CONSTRAINT complex_price_check CHECK (price >= 0 AND price < 10000);
ALTER TABLE ONLY folding_complex ADD CONSTRAINT complex_final_check CHECK (final_price = price - discount);
ALTER TABLE ONLY folding_complex ADD CONSTRAINT multi_constraint_code_key UNIQUE (code);

-- ============================================
-- Deferrability Edge Cases
-- ============================================
CREATE TABLE folding_deferrable (
    id_defer bigint NOT NULL,
    email_defer text NOT NULL,
    email_not_defer text NOT NULL
);

-- Primary Key with DEFERRABLE
ALTER TABLE ONLY folding_deferrable ADD CONSTRAINT folding_deferrable_pkey PRIMARY KEY (id_defer) DEFERRABLE;

-- Unique with DEFERRABLE INITIALLY DEFERRED
ALTER TABLE ONLY folding_deferrable ADD CONSTRAINT defer_unique UNIQUE (email_defer) DEFERRABLE INITIALLY DEFERRED;

-- Unique with NOT DEFERRABLE (should fold as normal)
ALTER TABLE ONLY folding_deferrable ADD CONSTRAINT not_defer_unique UNIQUE (email_not_defer) NOT DEFERRABLE;
