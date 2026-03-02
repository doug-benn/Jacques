CREATE TABLE users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    name text NOT NULL
);

CREATE OR REPLACE FUNCTION get_user_by_id(user_id bigint)
RETURNS TABLE(id bigint, email text, name text) AS $$
BEGIN
    RETURN QUERY
    SELECT u.id, u.email, u.name
    FROM users u
    WHERE u.id = user_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION get_user_email(user_id bigint)
RETURNS text AS $$
DECLARE
    email_text text;
BEGIN
    SELECT u.email INTO email_text
    FROM users u
    WHERE u.id = user_id;
    RETURN email_text;
END;
$$ LANGUAGE plpgsql;
