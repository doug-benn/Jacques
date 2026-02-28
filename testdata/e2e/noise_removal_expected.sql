-- Test fixture for noise removal
-- Transformations tested:
--   - SET statements removed
--   - GRANT/REVOKE statements removed
--   - COMMENT ON statements removed
--   - OWNER TO statements removed
--
-- Edge cases tested:
--   - Empty SET value
--   - Multi-statement SET
--   - GRANT with multiple privileges
--   - Empty COMMENT
--   - Multiple GRANT statements
--   - GRANT/REVOKE ON FUNCTION
--
-- Negative tests (should NOT be removed):
--   - SET inside function body (part of function definition)

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

-- ============================================
-- Edge cases for noise removal
-- ============================================

-- Edge case: Table for edge cases
CREATE TABLE public.edge_cases (
    id bigint NOT NULL
);

-- Negative test: SET inside function body (should NOT be removed - part of function)
CREATE FUNCTION public.set_in_function() RETURNS void AS $$
BEGIN
    SET statement_timeout = 3600;
END;
$$ LANGUAGE plpgsql;

-- Ensure edge case table has PK
ALTER TABLE ONLY public.edge_cases ADD CONSTRAINT edge_cases_pkey PRIMARY KEY (id);
