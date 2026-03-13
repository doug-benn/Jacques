CREATE TABLE base_table (
    id bigint PRIMARY KEY,
    val text
);

CREATE VIEW simple_view AS
SELECT id, val FROM base_table;

CREATE VIEW view_with_option AS
SELECT id, val FROM base_table WHERE id > 10 WITH CASCADED CHECK OPTION;

CREATE MATERIALIZED VIEW mat_view AS
SELECT id, count(*) as cnt FROM base_table GROUP BY id;

CREATE INDEX idx_mat_view_id ON mat_view(id);
