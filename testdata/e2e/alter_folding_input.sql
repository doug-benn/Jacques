-- Test fixture for structural folding
-- Transformations tested:
--   - PRIMARY KEY, UNIQUE, CHECK, EXCLUDE folded into CREATE TABLE
--   - Foreign Key inlining (including actions and MATCH)
--   - Table Inheritance preservation

-- ============================================
-- 1. Basic Constraints & Deferrability
-- ============================================
CREATE TABLE folding_basics (
    id bigint NOT NULL,
    sku text NOT NULL,
    user_id bigint NOT NULL,
    order_number text NOT NULL,
    balance numeric(10,2) NOT NULL,
    id_defer bigint NOT NULL,
    email_defer text NOT NULL
);

ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_pkey PRIMARY KEY (id);
ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_sku_key UNIQUE (sku);
ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_multi_unique UNIQUE (user_id, order_number);
ALTER TABLE ONLY folding_basics ADD CONSTRAINT folding_basics_balance_check CHECK (balance >= 0);
-- Deferrable PK
ALTER TABLE ONLY folding_basics ADD CONSTRAINT defer_pkey PRIMARY KEY (id_defer) DEFERRABLE;
-- Deferrable Unique
ALTER TABLE ONLY folding_basics ADD CONSTRAINT defer_unique UNIQUE (email_defer) DEFERRABLE INITIALLY DEFERRED;

-- ============================================
-- 2. Exclusion Constraints & USING Clause
-- ============================================
CREATE TABLE folding_exclude (
    id bigint NOT NULL,
    room_id bigint NOT NULL,
    booking_date date NOT NULL,
    val text
);

ALTER TABLE ONLY folding_exclude ADD PRIMARY KEY (id) USING btree;
ALTER TABLE ONLY folding_exclude ADD CONSTRAINT no_double_booking EXCLUDE USING btree (room_id WITH =, booking_date WITH =);
ALTER TABLE ONLY folding_exclude ADD CONSTRAINT partial_exclude EXCLUDE USING btree (room_id WITH =) WHERE (val IS NOT NULL);

-- ============================================
-- 3. Foreign Keys (Inlining & Actions)
-- ============================================
CREATE TABLE fk_parent (
    id bigint PRIMARY KEY
);

CREATE TABLE fk_child (
    id bigint PRIMARY KEY,
    parent_id bigint,
    col_cascade bigint NOT NULL,
    col_match_full bigint
);

-- Simple inlining
ALTER TABLE ONLY fk_child ADD CONSTRAINT simple_fkey FOREIGN KEY (parent_id) REFERENCES fk_parent(id);
-- Inlining with action
ALTER TABLE ONLY fk_child ADD CONSTRAINT cascade_fkey FOREIGN KEY (col_cascade) REFERENCES fk_parent(id) ON DELETE CASCADE;
-- Inlining with MATCH
ALTER TABLE ONLY fk_child ADD CONSTRAINT match_fkey FOREIGN KEY (col_match_full) REFERENCES fk_parent(id) MATCH FULL;

-- Self-referential (passes through)
ALTER TABLE ONLY fk_child ADD CONSTRAINT self_fkey FOREIGN KEY (id) REFERENCES fk_child(id);

-- ============================================
-- 4. Inheritance
-- ============================================
CREATE TABLE parent_table (
    id bigint PRIMARY KEY,
    name text NOT NULL
);

-- Basic inheritance
CREATE TABLE child_table (
    val text
) INHERITS (parent_table);

-- ONLY removal in INHERITS
CREATE TABLE only_child (
    val text
) INHERITS (ONLY parent_table);
