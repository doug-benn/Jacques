CREATE TABLE base_table (
    id bigint PRIMARY KEY,
    val text NOT NULL
);

CREATE OR REPLACE FUNCTION get_val(row_id bigint) RETURNS text AS $$
BEGIN
    RETURN (SELECT val FROM base_table WHERE id = row_id);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_all_rows() 
RETURNS TABLE(id bigint, val text) AS $$
BEGIN
    -- Internal comment: this should stay
    /* Block comment: this should also stay */
    SET statement_timeout = '1s';
    RETURN QUERY SELECT * FROM base_table;
END;
$$ LANGUAGE plpgsql;

CREATE FUNCTION add_one(integer) RETURNS integer
    AS 'select $1 + 1;'
    LANGUAGE SQL;
