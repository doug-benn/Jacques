-- Test fixture for index transformations
-- Covers: Partial indexes (WHERE), Covering indexes (INCLUDE), Expression indexes, 
--         Redundant index removal, and Duplicate index removal.

CREATE TABLE index_test (
    id bigint NOT NULL,
    val text NOT NULL,
    status text NOT NULL,
    data jsonb
);

ALTER TABLE ONLY index_test ADD CONSTRAINT index_test_pkey PRIMARY KEY (id);
ALTER TABLE ONLY index_test ADD CONSTRAINT index_test_val_key UNIQUE (val);

-- 1. Partial Index (should stay)
CREATE INDEX idx_partial ON index_test(status) WHERE status = 'active';

-- 2. Covering Index (should stay)
CREATE INDEX idx_include ON index_test(status) INCLUDE (data);

-- 3. Expression Index (should stay)
CREATE INDEX idx_expr ON index_test((lower(val)));

-- 4. Redundant Index Removal (should be REMOVED)
-- Column 'val' is already UNIQUE, so this plain index is redundant.
CREATE INDEX idx_redundant ON index_test(val);

-- 5. Duplicate Index Removal (should be REMOVED)
-- Same definition as idx_expr.
CREATE INDEX idx_duplicate ON index_test((lower(val)));

-- 6. Non-redundant UNIQUE index with different properties (should stay)
-- Even though 'val' is UNIQUE, this one has INCLUDE, making it distinct.
CREATE UNIQUE INDEX idx_val_include ON index_test(val) INCLUDE (status);
