-- Test fixture for Views
-- Covers: Regular views, materialized views, indexes on materialized views

-- Base table
CREATE TABLE public.base_table (
    id bigint NOT NULL,
    val text
);
ALTER TABLE ONLY public.base_table ADD PRIMARY KEY (id);

-- Regular View
CREATE VIEW simple_view AS
SELECT id, val FROM public.base_table;

-- View with options (e.g. WITH CASCADED CHECK OPTION - typically passed through)
CREATE VIEW view_with_option AS
SELECT id, val FROM public.base_table WHERE id > 10 WITH CASCADED CHECK OPTION;

-- Materialized View
CREATE MATERIALIZED VIEW mat_view AS
SELECT id, count(*) as cnt FROM public.base_table GROUP BY id;

-- Index on Materialized View
CREATE INDEX idx_mat_view_id ON mat_view(id);
