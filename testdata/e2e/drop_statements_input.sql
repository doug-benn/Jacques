-- Test fixture for DROP statements
-- Covers: Adding IF EXISTS to various DROP statements, and schema stripping

DROP TABLE public.drop_test;
DROP INDEX idx_drop_val;
DROP SEQUENCE drop_seq;
DROP VIEW drop_view;
DROP MATERIALIZED VIEW drop_mat_view;

-- Objects to exist so they can be referenced/dropped in the E2E test
CREATE SEQUENCE drop_seq;
CREATE TABLE public.drop_test (
    id bigint PRIMARY KEY DEFAULT nextval('drop_seq'),
    val text
);
CREATE INDEX idx_drop_val ON public.drop_test(val);
CREATE VIEW drop_view AS SELECT id FROM public.drop_test;
CREATE MATERIALIZED VIEW drop_mat_view AS SELECT id FROM public.drop_test;
