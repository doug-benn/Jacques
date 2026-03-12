CREATE TABLE comment_test (
    id bigint
);

CREATE TABLE string_test (
    val text DEFAULT '-- not a comment'
);

CREATE FUNCTION preserved_comments() RETURNS text AS $$
BEGIN
    -- This line comment should stay
    /* This block comment
       should also stay */
    RETURN '/* not a comment */ -- still not a comment';
END;
$$ LANGUAGE plpgsql;
