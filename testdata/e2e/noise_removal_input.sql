-- Test fixture for noise removal
-- Transformations tested:
--   - SET statements removed
--   - GRANT statements removed
--   - COMMENT statements removed
--   - OWNER TO statements removed
--   - psql metacommands removed
--
-- Edge cases:
--   - Administrative commands inside function body (MUST be preserved)
--   - Administrative commands in string literal (MUST be preserved)
--   - Administrative commands in quoted identifiers (MUST be preserved)

SET statement_timeout = 0;

CREATE TABLE noise_test (
    id bigint
);

ALTER TABLE noise_test OWNER TO testuser;
GRANT ALL ON TABLE noise_test TO testuser;
COMMENT ON TABLE noise_test IS 'A noisy table';

-- Administrative commands inside function body
CREATE FUNCTION func_with_admin() RETURNS void AS $$
BEGIN
    SET statement_timeout = 3600;
    -- GRANT SELECT ON some_table TO user; (should be preserved since it's inside a comment inside a function)
    PERFORM 1;
END;
$$ LANGUAGE plpgsql;

-- Administrative commands in string literal
CREATE TABLE string_noise (
    val text DEFAULT 'SET statement_timeout = 0;'
);

-- Administrative commands in quoted identifiers (highly unlikely but possible)
CREATE TABLE "OWNER TO testuser" (
    "GRANT ALL" bigint
);
