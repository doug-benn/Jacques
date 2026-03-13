CREATE TABLE drop_test (
    id BIGSERIAL PRIMARY KEY,
    val text
);

DROP TABLE IF EXISTS drop_test;

DROP INDEX IF EXISTS idx_drop_val;

DROP SEQUENCE IF EXISTS drop_seq;

DROP VIEW IF EXISTS drop_view;

DROP MATERIALIZED VIEW IF EXISTS drop_mat_view;

CREATE INDEX idx_drop_val ON drop_test(val);

CREATE VIEW drop_view AS SELECT id FROM drop_test;

CREATE MATERIALIZED VIEW drop_mat_view AS SELECT id FROM drop_test;
