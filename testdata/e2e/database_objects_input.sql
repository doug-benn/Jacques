-- Test fixture for various database objects
-- Covers: Views, Functions, Triggers, RLS Policies, and Schema handling

-- 1. Schemas (Explicit and Inferred)
CREATE SCHEMA app;

-- Base table in public (public prefix should be stripped)
CREATE TABLE public.base_table (
    id bigint NOT NULL PRIMARY KEY,
    val text NOT NULL
);

-- Table in custom schema
CREATE TABLE app.app_table (
    id bigint NOT NULL PRIMARY KEY,
    base_id bigint REFERENCES public.base_table(id)
);

-- Table in quoted schema (should be inferred)
CREATE TABLE "MySchema".profile (
    id bigint NOT NULL PRIMARY KEY,
    bio text
);

-- 2. Views (Regular and Materialized)
CREATE VIEW simple_view AS SELECT id, val FROM public.base_table;

CREATE MATERIALIZED VIEW mat_view AS 
SELECT id, count(*) as cnt FROM public.base_table GROUP BY id;

CREATE INDEX idx_mat_view_id ON mat_view(id);

-- 3. Functions (Scalar and Table-returning)
CREATE OR REPLACE FUNCTION public.get_val(row_id bigint) RETURNS text AS $$
BEGIN
    RETURN (SELECT val FROM public.base_table WHERE id = row_id);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_all_rows() 
RETURNS TABLE(id bigint, val text) AS $$
BEGIN
    -- Internal comments MUST be preserved
    RETURN QUERY SELECT * FROM public.base_table;
END;
$$ LANGUAGE plpgsql;

-- 4. Triggers
CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.val = 'updated';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_before_update
    BEFORE UPDATE ON public.base_table
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- 5. RLS Policies
ALTER TABLE public.base_table ENABLE ROW LEVEL SECURITY;

CREATE POLICY select_all ON public.base_table FOR SELECT USING (true);
