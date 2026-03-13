CREATE SCHEMA app;

CREATE SCHEMA "MySchema";

CREATE TABLE base_table (
    id bigint NOT NULL PRIMARY KEY,
    val text NOT NULL
);

CREATE TABLE app.app_table (
    id bigint NOT NULL PRIMARY KEY,
    base_id bigint REFERENCES base_table(id)
);

CREATE TABLE "MySchema".profile (
    id bigint NOT NULL PRIMARY KEY,
    bio text
);

CREATE VIEW simple_view AS SELECT id, val FROM base_table;

CREATE MATERIALIZED VIEW mat_view AS 
SELECT id, count(*) as cnt FROM base_table GROUP BY id;

CREATE INDEX idx_mat_view_id ON mat_view(id);

CREATE OR REPLACE FUNCTION get_val(row_id bigint) RETURNS text AS $$
BEGIN
    RETURN (SELECT val FROM base_table WHERE id = row_id);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_all_rows() 
RETURNS TABLE(id bigint, val text) AS $$
BEGIN
    -- Internal comments MUST be preserved
    RETURN QUERY SELECT * FROM base_table;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.val = 'updated';
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_before_update
    BEFORE UPDATE ON base_table
    FOR EACH ROW EXECUTE FUNCTION update_timestamp();

ALTER TABLE base_table ENABLE ROW LEVEL SECURITY;

CREATE POLICY select_all ON base_table FOR SELECT USING (true);
