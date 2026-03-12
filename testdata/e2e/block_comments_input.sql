-- Test fixture for block comment removal
-- Transformations tested:
--   - /* ... */ block comments removed
--   - Block comments in various positions (header, inline, footer)
--
-- Edge cases:
--   - Comments inside function bodies (MUST be preserved)
--   - Comments inside string literals (MUST be preserved)

/* Header comment */

CREATE TABLE comment_test (
    id bigint /* Inline comment */
);

-- Comment inside a function (should NOT be removed)
CREATE FUNCTION preserved_comments() RETURNS text AS $$
BEGIN
    -- This line comment should stay
    /* This block comment
       should also stay */
    RETURN '/* not a comment */ -- still not a comment';
END;
$$ LANGUAGE plpgsql;

-- Comment inside a string literal (should NOT be removed)
CREATE TABLE string_test (
    val text DEFAULT '-- not a comment'
);

/* Footer comment */
