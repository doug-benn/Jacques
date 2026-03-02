CREATE TABLE users (
    id bigint NOT NULL,
    email text NOT NULL,
    name text NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
    id bigint NOT NULL,
    name text NOT NULL,
    price numeric(10,2) NOT NULL
);

CREATE TABLE edge_cases (
    id bigint NOT NULL
);

CREATE FUNCTION set_in_function() RETURNS void AS $$
BEGIN
    SET statement_timeout = 3600;
END;
$$ LANGUAGE plpgsql;
