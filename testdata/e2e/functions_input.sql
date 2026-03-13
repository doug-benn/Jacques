-- Test fixture for functions
-- Covers: Scalar function, table-returning function, preservation of internal content

CREATE TABLE public.base_table (
    id bigint NOT NULL,
    val text NOT NULL
);
ALTER TABLE ONLY public.base_table ADD PRIMARY KEY (id);

-- 1. Simple scalar function with schema prefix
CREATE OR REPLACE FUNCTION public.get_val(row_id bigint) RETURNS text AS $$
BEGIN
    RETURN (SELECT val FROM public.base_table WHERE id = row_id);
END;
$$ LANGUAGE plpgsql;

-- 2. Table-returning function with internal comments and SET (MUST be preserved)
CREATE OR REPLACE FUNCTION public.get_all_rows() 
RETURNS TABLE(id bigint, val text) AS $$
BEGIN
    -- Internal comment: this should stay
    /* Block comment: this should also stay */
    SET statement_timeout = '1s';
    RETURN QUERY SELECT * FROM public.base_table;
END;
$$ LANGUAGE plpgsql;

-- 3. Function with different language
CREATE FUNCTION add_one(integer) RETURNS integer
    AS 'select $1 + 1;'
    LANGUAGE SQL;
