-- Test fixture for noise removal
-- Transformations tested:
--   - Administrative noise (SET, GRANT, OWNER TO)
--   - psql metacommands (\set, \unset, etc.)
--   - ONLY keyword removal
--   - Block and line comment removal (quote-aware)
--   - DROP IF EXISTS addition

-- 1. Administrative Noise & Metacommands
SET statement_timeout = 0;
\set QUIET on
\set EMPTY_VAR
\restricted

-- 2. Comments (Header, Inline, Footer)
/* Header block comment */
-- Line comment

CREATE TABLE noise_test (
    id bigint NOT NULL,
    val text /* Inline comment */
);

-- 3. ONLY Removal & Ownership/Grants
ALTER TABLE ONLY noise_test OWNER TO testuser;
GRANT ALL ON TABLE noise_test TO testuser;
COMMENT ON TABLE noise_test IS 'A noisy table';

-- 4. DROP IF EXISTS
DROP TABLE public.drop_test;
DROP VIEW public.drop_view;

-- Objects for DROP test
CREATE TABLE drop_test (id bigint PRIMARY KEY);
CREATE VIEW drop_view AS SELECT id FROM drop_test;

-- 5. Edge Case: Preservation inside functions/strings
CREATE FUNCTION func_with_noise() RETURNS void AS $$
BEGIN
    -- Internal comments and SET MUST be preserved
    SET statement_timeout = 3600;
    PERFORM 1;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE string_noise (
    val text DEFAULT 'SET timeout; -- not a comment; ONLY stay'
);
