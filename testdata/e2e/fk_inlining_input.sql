-- Test fixture for FK inlining
-- Transformations tested:
--   - Simple FK inlining into column definition
--   - ON DELETE/UPDATE actions (CASCADE, SET NULL, SET DEFAULT, RESTRICT)
--   - Self-referential and circular FKs
--
-- Edge cases tested:
--   - Multi-column FK (should pass through)
--   - NOT VALID FK (should pass through)
--   - MATCH FULL/PARTIAL FK (should inline)
--   - Quoted constraint names

-- Parent table for references
CREATE TABLE fk_parent (
    id bigint NOT NULL
);
ALTER TABLE ONLY fk_parent ADD PRIMARY KEY (id);

-- Actions test: One table to exercise all basic actions
CREATE TABLE fk_actions (
    id bigint NOT NULL,
    col_cascade bigint NOT NULL,
    col_setnull bigint,
    col_setdef bigint DEFAULT 1,
    col_restrict bigint NOT NULL,
    col_upd_cascade bigint NOT NULL
);
ALTER TABLE ONLY fk_actions ADD PRIMARY KEY (id);

ALTER TABLE ONLY fk_actions ADD CONSTRAINT actions_cascade_fkey FOREIGN KEY (col_cascade) REFERENCES fk_parent(id) ON DELETE CASCADE;
ALTER TABLE ONLY fk_actions ADD CONSTRAINT actions_setnull_fkey FOREIGN KEY (col_setnull) REFERENCES fk_parent(id) ON DELETE SET NULL;
ALTER TABLE ONLY fk_actions ADD CONSTRAINT actions_setdef_fkey FOREIGN KEY (col_setdef) REFERENCES fk_parent(id) ON DELETE SET DEFAULT;
ALTER TABLE ONLY fk_actions ADD CONSTRAINT actions_restrict_fkey FOREIGN KEY (col_restrict) REFERENCES fk_parent(id) ON DELETE RESTRICT;
ALTER TABLE ONLY fk_actions ADD CONSTRAINT actions_upd_cascade_fkey FOREIGN KEY (col_upd_cascade) REFERENCES fk_parent(id) ON UPDATE CASCADE;

-- Self-referential and Circular
CREATE TABLE fk_circular_a (
    id bigint NOT NULL,
    parent_id bigint,
    other_id bigint
);
ALTER TABLE ONLY fk_circular_a ADD PRIMARY KEY (id);

CREATE TABLE fk_circular_b (
    id bigint NOT NULL,
    other_id bigint
);
ALTER TABLE ONLY fk_circular_b ADD PRIMARY KEY (id);

-- Self-referential (should inline)
ALTER TABLE ONLY fk_circular_a ADD CONSTRAINT self_fkey FOREIGN KEY (parent_id) REFERENCES fk_circular_a(id);

-- Circular (should inline both)
ALTER TABLE ONLY fk_circular_a ADD CONSTRAINT circ_a_fkey FOREIGN KEY (other_id) REFERENCES fk_circular_b(id);
ALTER TABLE ONLY fk_circular_b ADD CONSTRAINT circ_b_fkey FOREIGN KEY (other_id) REFERENCES fk_circular_a(id);

-- Complex cases (Pass-throughs and Match)
CREATE TABLE fk_complex (
    id bigint NOT NULL,
    sub_id bigint NOT NULL,
    ref_id bigint,
    match_full_id bigint,
    match_partial_id bigint
);
ALTER TABLE ONLY fk_complex ADD PRIMARY KEY (id, sub_id);

-- Multi-column FK (cannot inline - pass through)
ALTER TABLE ONLY fk_complex ADD CONSTRAINT multi_fkey FOREIGN KEY (id, sub_id) REFERENCES fk_complex(id, sub_id);

-- NOT VALID (cannot inline - pass through)
ALTER TABLE ONLY fk_complex ADD CONSTRAINT not_valid_fkey FOREIGN KEY (ref_id) REFERENCES fk_parent(id) NOT VALID;

-- MATCH clauses (should inline)
ALTER TABLE ONLY fk_complex ADD CONSTRAINT match_full_fkey FOREIGN KEY (match_full_id) REFERENCES fk_parent(id) MATCH FULL;
ALTER TABLE ONLY fk_complex ADD CONSTRAINT match_partial_fkey FOREIGN KEY (match_partial_id) REFERENCES fk_parent(id) MATCH PARTIAL;

-- Quoted identifier
ALTER TABLE ONLY fk_complex ADD CONSTRAINT "quoted fk name" FOREIGN KEY (ref_id) REFERENCES fk_parent(id);
