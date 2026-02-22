-- Test fixture for functions
-- Covers: Scalar function, table-returning function

CREATE TABLE public.users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL
);

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);

CREATE OR REPLACE FUNCTION public.get_user_by_id(user_id bigint)
RETURNS TABLE(id bigint, email text, name text) AS $$
BEGIN
    RETURN QUERY
    SELECT u.id, u.email, u.name
    FROM public.users u
    WHERE u.id = user_id;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION public.get_user_email(user_id bigint)
RETURNS text AS $$
DECLARE
    email_text text;
BEGIN
    SELECT u.email INTO email_text
    FROM public.users u
    WHERE u.id = user_id;
    RETURN email_text;
END;
$$ LANGUAGE plpgsql;
