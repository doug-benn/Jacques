-- Test fixture for sequence handling
-- Covers: shared sequences, SERIAL conversion, sequence preservation, edge cases
--
-- Transformations tested:
--   - Dedicated sequence → BIGSERIAL/SERIAL/SMALLSERIAL conversion
--   - Shared sequence → preserved (not converted)
--
-- Edge cases tested:
--   - Sequence with explicit START WITH/INCREMENT BY (should convert)
--   - Sequence without explicit OWNED BY (should convert)
--   - Sequence with MINVALUE/MAXVALUE (should convert)
--
-- Negative tests (should NOT be converted):
--   - Sequence with CACHE > 1 (could have gaps)

-- ============================================
-- Shared sequence - used by multiple tables, should be preserved
-- ============================================
CREATE SEQUENCE shared_seq START WITH 1 INCREMENT BY 1 NO MAXVALUE CACHE 1;

CREATE TABLE shared_table_1 (
    id bigint NOT NULL DEFAULT nextval('shared_seq'::regclass)
);

CREATE TABLE shared_table_2 (
    id bigint NOT NULL DEFAULT nextval('shared_seq'::regclass)
);

-- ============================================
-- Dedicated sequences - should convert to SERIAL types
-- ============================================

-- Table testing all three SERIAL types in one go
CREATE TABLE serial_types (
    id_big bigint NOT NULL,
    id_reg integer NOT NULL,
    id_small smallint NOT NULL
);

CREATE SEQUENCE seq_big OWNED BY serial_types.id_big;
CREATE SEQUENCE seq_reg OWNED BY serial_types.id_reg;
CREATE SEQUENCE seq_small OWNED BY serial_types.id_small;

ALTER TABLE serial_types ALTER COLUMN id_big SET DEFAULT nextval('seq_big');
ALTER TABLE serial_types ALTER COLUMN id_reg SET DEFAULT nextval('seq_reg');
ALTER TABLE serial_types ALTER COLUMN id_small SET DEFAULT nextval('seq_small');

-- ============================================
-- Edge cases for sequence handling
-- ============================================

CREATE TABLE edge_case_table (
    id_start bigint NOT NULL,
    id_unowned bigint NOT NULL,
    id_bounded bigint NOT NULL
);

-- Explicit START WITH/INCREMENT BY
CREATE SEQUENCE start_seq START WITH 1000 INCREMENT BY 10 OWNED BY edge_case_table.id_start;
ALTER TABLE edge_case_table ALTER COLUMN id_start SET DEFAULT nextval('start_seq');

-- Without explicit OWNED BY
CREATE SEQUENCE unowned_seq;
ALTER TABLE edge_case_table ALTER COLUMN id_unowned SET DEFAULT nextval('unowned_seq');

-- With MINVALUE/MAXVALUE
CREATE SEQUENCE bounded_seq START WITH 1 MINVALUE 1 MAXVALUE 1000 OWNED BY edge_case_table.id_bounded;
ALTER TABLE edge_case_table ALTER COLUMN id_bounded SET DEFAULT nextval('bounded_seq');

-- ============================================
-- Negative test: CACHE > 1 (should NOT convert)
-- ============================================
CREATE SEQUENCE cached_seq CACHE 100;

CREATE TABLE cached_table (
    id bigint NOT NULL DEFAULT nextval('cached_seq')
);

-- Ensure all tables have PKs to exercise folding + sequence detection
ALTER TABLE shared_table_1 ADD PRIMARY KEY (id);
ALTER TABLE shared_table_2 ADD PRIMARY KEY (id);
ALTER TABLE serial_types ADD PRIMARY KEY (id_big);
ALTER TABLE edge_case_table ADD PRIMARY KEY (id_start);
ALTER TABLE cached_table ADD PRIMARY KEY (id);
