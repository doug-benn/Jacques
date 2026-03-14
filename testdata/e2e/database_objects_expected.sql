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

CREATE TABLE other_table (
    id bigint NOT NULL PRIMARY KEY,
    name text NOT NULL
);

CREATE VIEW simple_view AS SELECT id, val FROM base_table;

CREATE MATERIALIZED VIEW mat_view AS 
SELECT id, count(*) as cnt FROM base_table GROUP BY id;

CREATE INDEX idx_mat_view_id ON mat_view(id);

CREATE VIEW ordered_view AS
SELECT id, val FROM base_table
ORDER BY id DESC LIMIT 100;

CREATE VIEW joined_view AS
SELECT a.id, a.val, b.name
FROM base_table a
JOIN other_table b ON a.id = b.id;

CREATE VIEW union_view AS
SELECT id, val FROM base_table WHERE val = 'a'
UNION
SELECT id, val FROM base_table WHERE val = 'b';

CREATE VIEW updatable_view AS
SELECT id, val FROM base_table;

CREATE OR REPLACE FUNCTION updatable_view_trigger() RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO base_table (id, val) VALUES (NEW.id, NEW.val);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE INSTEAD OF INSERT ON updatable_view
FOR EACH ROW EXECUTE FUNCTION updatable_view_trigger();

CREATE MATERIALIZED VIEW mat_distinct_view AS
SELECT DISTINCT val FROM base_table;

CREATE INDEX idx_mat_distinct ON mat_distinct_view(val);

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
