CREATE TABLE noise_test (
    id bigint
);

CREATE TABLE string_noise (
    val text DEFAULT 'SET statement_timeout = 0;'
);

CREATE TABLE "OWNER TO testuser" (
    "GRANT ALL" bigint
);

CREATE FUNCTION func_with_admin() RETURNS void AS $$
BEGIN
    SET statement_timeout = 3600;
    -- GRANT SELECT ON some_table TO user; (should be preserved since it's inside a comment inside a function)
    PERFORM 1;
END;
$$ LANGUAGE plpgsql;
