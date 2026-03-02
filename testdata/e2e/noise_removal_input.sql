-- Test fixture for noise removal
-- Transformations tested:
--   - SET statements removed
--   - GRANT/REVOKE statements removed
--   - COMMENT ON statements removed
--   - OWNER TO statements removed
--   - psql metacommands removed
--   - Block comments removed
--
-- Edge cases tested:
--   - Empty SET value
--   - Multi-statement SET
--   - GRANT with multiple privileges
--   - Empty COMMENT
--   - Multiple GRANT statements
--   - GRANT/REVOKE ON FUNCTION
--   - \set and \unset variations
--   - Block comments in various positions
--
-- Negative tests (should NOT be removed):
--   - SET inside function body (part of function definition)

/* Header comment for the schema */

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET constraint_exclusion = partition;
SET row_security = on;

/* Table with inline block comment */

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

/* Inline block comment in table definition */

ALTER TABLE public.users OWNER TO testuser;

GRANT ALL ON TABLE public.users TO testuser;
GRANT SELECT ON TABLE public.users TO readonly_user;
GRANT INSERT, UPDATE ON TABLE public.users TO readwrite_user;

COMMENT ON TABLE public.users IS 'Application users table';
COMMENT ON COLUMN public.users.email IS 'User email address';
COMMENT ON COLUMN public.users.name IS 'User full name';

/* Another table for testing */

CREATE TABLE public.products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

/* Multiple block comments between statements */

ALTER TABLE public.products OWNER TO admin;

GRANT ALL ON TABLE public.products TO admin;
GRANT SELECT ON TABLE public.products TO public;

COMMENT ON TABLE public.products IS 'Product catalog';

-- ============================================
-- Edge cases for noise removal
-- ============================================

-- Edge case: Empty SET value
SET statement_timeout = ;

-- Edge case: Multi-statement SET
SET a = 1; SET b = 2;

/* Block comment around edge case */

-- Edge case: GRANT with multiple privileges
GRANT SELECT, INSERT, UPDATE ON public.users TO app_user;

-- Edge case: Multiple GRANT statements
GRANT SELECT ON public.products TO user1;
GRANT INSERT ON public.products TO user1;
GRANT UPDATE ON public.products TO user1;

-- Edge case: Empty COMMENT
COMMENT ON TABLE public.users IS ;

-- Edge case: GRANT ON FUNCTION (should be removed)
GRANT EXECUTE ON FUNCTION public.my_function() TO app_user;

-- Edge case: REVOKE ON FUNCTION (should be removed)
REVOKE EXECUTE ON FUNCTION public.my_function() FROM app_user;

/* Block comment around table */

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

-- ============================================
-- psql metacommands (\set, \unset)
-- ============================================

-- Edge case: \set command (should be removed)
\set QUIET on

-- Edge case: \set with variable (should be removed)
\set HISTSIZE 1000

-- Edge case: \unset command (should be removed)
\unset QUIET

-- Edge case: \set with empty value (should be removed)
\set EMPTY_VAR

\restricted
\unrestricted

/* Footer comment */

-- Ensure edge case table has PK
ALTER TABLE ONLY public.edge_cases ADD CONSTRAINT edge_cases_pkey PRIMARY KEY (id);
